// builder builds packages
package main

import (
	"./stdlib/fmt"
	"./stdlib/path"
	"./stdlib/strings"
	"./stdlib/io/ioutil"
	"os"
)

// "fmt" => "/stdlib/fmt"
// "./stdlib/fmt" => "/stdlib/fmt"
// "./mylib"      => "./mylib"
func normalizeImportPath(currentPath string, pth string) normalizedPackagePath {
	if strings.HasPrefix(pth, "./stdlib/") {
		// "./stdlib/fmt" => "/stdlib/fmt"
		return normalizedPackagePath(pth[1:])
	} else if strings.HasPrefix(pth, "./") {
		// parser relative pth
		// "./mylib" => "/mylib"
		return normalizedPackagePath("./" + currentPath + pth[1:])
	} else {
		// "fmt" => "/stdlib/fmt"
		return normalizedPackagePath("/stdlib/" + pth)
	}
}

func getParsingDir(sourceFile string) string {
	found := strings.LastIndexByte(sourceFile, '/')
	if found == -1 {
		return "."
	}
	return path.Dir(sourceFile)
}

// analyze imports of a given go source
func parseImportsFromFile(sourceFile string) importMap {
	p := &parser{
		parsingDir: getParsingDir(sourceFile),
	}
	astFile := p.ParseFile(sourceFile, nil, true)
	return astFile.imports
}

// analyze imports of given go files
func parseImports(sourceFiles []string) importMap {

	var imported importMap = make(map[normalizedPackagePath]bool)
	for _, sourceFile := range sourceFiles {
		importsInFile := parseImportsFromFile(sourceFile)
		for name, _ := range importsInFile {
			imported[name] = true
		}
	}

	return imported
}

// inject builtin functions into the universe scope
func compileUniverse(universe *Scope) *AstPackage {
	p := &parser{
		packagePath: normalizeImportPath("", "builtin"), // anything goes
		packageName: identifier("builtin"),
	}
	f := p.ParseFile("stdlib/builtin/builtin.go", universe, false)
	attachMethodsToTypes(f.methods, p.packageBlockScope)
	inferTypes(f.uninferredGlobals, f.uninferredLocals)
	calcStructSize(f.dynamicTypes)
	return &AstPackage{
		name:           identifier(""),
		files:          []*AstFile{f},
		stringLiterals: f.stringLiterals,
		dynamicTypes:   f.dynamicTypes,
	}
}

// "path/dir" => {"path/dir/a.go", ...}
func getPackageFiles(pkgDir string) []string {
	f, err := os.Open(pkgDir)
	if err != nil {
		panic(err)
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	var sourceFiles []string
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		// minigo ignores ignore.go
		if name == "ignore.go" {
			continue
		}
		sourceFiles = append(sourceFiles, pkgDir+"/"+name)
	}
	return sourceFiles
}

// inject unsafe package
func compileUnsafe(universe *Scope) *AstPackage {
	pkgName := identifier("unsafe")
	pkgPath := normalizeImportPath("", "unsafe") // need to be normalized because it's imported by iruntime
	pkgScope := newScope(nil, pkgName)
	symbolTable.allScopes[pkgPath] = pkgScope
	sourceFiles := getPackageFiles(convertStdPath(pkgPath))
	pkg := ParseFiles(pkgName, pkgPath, pkgScope, sourceFiles)
	makePkg(pkg, universe)
	return pkg
}

const IRuntimePath normalizedPackagePath = "iruntime"
const MainPath normalizedPackagePath = "main"
const IRuntimePkgName identifier = "iruntime"
var pkgIRuntime *AstPackage

// inject runtime things into the universe scope
func compileRuntime(universe *Scope) *AstPackage {
	pkgName := IRuntimePkgName
	pkgPath := IRuntimePath
	pkgScope := newScope(nil, pkgName)
	symbolTable.allScopes[pkgPath] = pkgScope
	sourceFiles := getPackageFiles("internal/runtime")
	pkg := ParseFiles(pkgName, pkgPath, universe, sourceFiles)
	makePkg(pkg, universe)
	pkgIRuntime = pkg

	// read asm files
	for _, asmfile := range []string{"internal/runtime/asm_amd64.s", "internal/runtime/runtime.s"} {
		buf, _ := ioutil.ReadFile(asmfile)
		pkgIRuntime.asm = append(pkgIRuntime.asm, string(buf))
	}

	return pkg
}

func makePkg(pkg *AstPackage, universe *Scope) *AstPackage {
	resolveIdents(pkg, universe)
	attachMethodsToTypes(pkg.methods, pkg.scope)
	inferTypes(pkg.uninferredGlobals, pkg.uninferredLocals)
	calcStructSize(pkg.dynamicTypes)
	return pkg
}

// compileMainFiles parses files into *AstPackage
func compileMainFiles(universe *Scope, sourceFiles []string) *AstPackage {
	// compile the main package
	pkgName := identifier("main")
	pkgScope := newScope(nil, pkgName)
	mainPkg := ParseFiles("main", MainPath, pkgScope, sourceFiles)
	if parseOnly {
		if debugAst {
			mainPkg.dump()
		}
		return nil
	}
	mainPkg = makePkg(mainPkg, universe)
	if debugAst {
		mainPkg.dump()
	}

	if resolveOnly {
		return nil
	}
	return mainPkg
}

type importMap map[normalizedPackagePath]bool

func parseImportRecursive(dep map[normalizedPackagePath]importMap, directDependencies importMap) {
	for pth, _ := range directDependencies {
		files := getPackageFiles(convertStdPath(pth))
		var imports importMap = make(map[normalizedPackagePath]bool)
		for _, file := range files {
			imprts := parseImportsFromFile(file)
			for k, v := range imprts {
				imports[k] = v
			}
		}
		dep[pth] = imports
		parseImportRecursive(dep, imports)
	}
}

func removeResolvedPkg(dep map[normalizedPackagePath]importMap, pkgToRemove normalizedPackagePath) map[normalizedPackagePath]importMap {
	var dep2 map[normalizedPackagePath]importMap = make(map[normalizedPackagePath]importMap)

	for pkg1, imports := range dep {
		if pkg1 == pkgToRemove {
			continue
		}
		var newimports importMap = make(map[normalizedPackagePath]bool)
		for pkg2, _ := range imports {
			if pkg2 == pkgToRemove {
				continue
			}
			newimports[pkg2] = true
		}
		dep2[pkg1] = newimports
	}

	return dep2
}

func removeResolvedPackages(dep map[normalizedPackagePath]importMap, sortedUniqueImports []normalizedPackagePath) map[normalizedPackagePath]importMap {
	for _, resolved := range sortedUniqueImports {
		dep = removeResolvedPkg(dep, resolved)
	}
	return dep
}

func dumpDep(dep map[string]importMap) {
	debugf("#------------- dep -----------------")
	for spkgName, imports := range dep {
		debugf("#  %s has %d imports:", spkgName, len(imports))
		for sspkgName, _ := range imports {
			debugf("#    %s", sspkgName)
		}
	}
}

func get0dependentPackages(dep map[normalizedPackagePath]importMap) []normalizedPackagePath {
	var moved []normalizedPackagePath
	if len(dep) == 0 {
		return nil
	}
	for spkgName, imports := range dep {
		var numDepends int
		for _, flg := range imports {
			if flg {
				numDepends++
			}
		}
		if numDepends == 0 {
			moved = append(moved, spkgName)
		}
	}
	return moved
}

func resolveDependency(directDependencies importMap) []normalizedPackagePath {
	var sortedUniqueImports []normalizedPackagePath
	var dep map[normalizedPackagePath]importMap = make(map[normalizedPackagePath]importMap)
	parseImportRecursive(dep, directDependencies)

	for {
		//dumpDep(dep)
		moved := get0dependentPackages(dep)
		if len(moved) == 0 {
			break
		}
		dep = removeResolvedPackages(dep, moved)
		for _, pkg := range moved {
			sortedUniqueImports = append(sortedUniqueImports, pkg)
		}

	}
	return sortedUniqueImports
}

// if "/stdlib/foo" => "./stdlib/foo"
func convertStdPath(pth normalizedPackagePath) string {
	if strings.HasPrefix(string(pth), "/stdlib/") {
		return fmt.Sprintf(".%s", string(pth))
	}
	return string(pth)
}

// Compile standard libraries
func compileStdLibs(universe *Scope, directDependencies importMap) map[normalizedPackagePath]*AstPackage {

	sortedUniqueImports := resolveDependency(directDependencies)

	var compiledStdPkgs map[normalizedPackagePath]*AstPackage = make(map[normalizedPackagePath]*AstPackage)

	for _, pth := range sortedUniqueImports {
		files := getPackageFiles(convertStdPath(pth))
		pkgScope := newScope(nil, identifier(pth))
		symbolTable.allScopes[pth] = pkgScope
		pkgShortName := path.Base(string(pth))
		pkg := ParseFiles(identifier(pkgShortName), pth, pkgScope, files)
		pkg = makePkg(pkg, universe)
		compiledStdPkgs[pth] = pkg
	}

	return compiledStdPkgs
}

type Program struct {
	packages    []*AstPackage
	methodTable map[int][]string
}

func build(pkgUniverse *AstPackage, pkgUnsafe *AstPackage, pkgIRuntime *AstPackage, stdPkgs map[normalizedPackagePath]*AstPackage, pkgMain *AstPackage) *Program {
	var packages []*AstPackage

	packages = append(packages, pkgUniverse)
	packages = append(packages, pkgUnsafe)
	packages = append(packages, pkgIRuntime)

	for _, pkg := range stdPkgs {
		packages = append(packages, pkg)
	}

	packages = append(packages, pkgMain)

	var dynamicTypes []*Gtype
	var funcs []*DeclFunc

	for _, pkg := range packages {
		collectDecls(pkg)
		if pkg == pkgUniverse {
			setStringLables(pkg, "pkgUniverse")
		} else {
			setStringLables(pkg, pkg.name)
		}
		for _, dt := range pkg.dynamicTypes {
			dynamicTypes = append(dynamicTypes, dt)
		}
		for _, fn := range pkg.funcs {
			funcs = append(funcs, fn)
		}
		setTypeIds(pkg.namedTypes)
	}

	//  Do restructuring of local nodes
	for _, pkg := range packages {
		for _, fnc := range pkg.funcs {
			fnc = walkFunc(fnc)
		}
	}

	symbolTable.uniquedDTypes = uniqueDynamicTypes(dynamicTypes)

	program := &Program{}
	program.packages = packages
	program.methodTable = composeMethodTable(funcs)
	return program
}
