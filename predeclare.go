package main

// built-in types
const sizeOfInterface = 8 * 3

var sInterface = Gtype{kind: G_INTERFACE, size: sizeOfInterface}
var gInterface = &sInterface
var sInt = Gtype{kind: G_INT, size: 8}
var gInt = &sInt
var gByte = &Gtype{kind: G_BYTE, size: 1}
var gBool = &Gtype{kind: G_BOOL, size: 8} // we treat bool as quad length data for now
var gString = &Gtype{
	kind: G_CLIKE_STRING,
}

var builtinTypesAsString []gostring = []gostring{
	gostring("bool"),
	gostring("byte"),
	gostring("int"),
	gostring("string"),
	gostring("func"),
}

var eIota = &ExprConstVariable{
	name: goidentifier("iota"),
}

var builtinPanic = &DeclFunc{
	rettypes: []*Gtype{},
}

var builtinLen = &DeclFunc{
	rettypes: []*Gtype{&sInt},
}

var builtinCap = &DeclFunc{
	rettypes: []*Gtype{&sInt},
}

var builtinAppend = &DeclFunc{
	rettypes: []*Gtype{&sInt},
}

var builtinMakeSlice = &DeclFunc{
	rettypes: []*Gtype{&sBuiltinRunTimeArgsRettypes1},
}

var builtinDumpSlice = &DeclFunc{
	rettypes: []*Gtype{},
}

var builtinDumpInterface = &DeclFunc{
	rettypes: []*Gtype{},
}

var builtinAssertInterface = &DeclFunc{
	rettypes: []*Gtype{},
}

var builtinAsComment = &DeclFunc{
	rettypes: []*Gtype{},
}

var sBuiltinRunTimeArgsRettypes1 Gtype = Gtype{
	kind: G_SLICE,
	size: IntSize * 3,
	elementType: &Gtype{
		kind: G_CLIKE_STRING,
	},
}

var builtinPrintstring = &DeclFunc{
	rettypes: []*Gtype{},
}

func newUniverse() *Scope {
	universe := newScope(nil, S("universe"))
	setPredeclaredIdentifiers(universe)
	return universe
}

// https://golang.org/ref/spec#Predeclared_identifiers
func setPredeclaredIdentifiers(universe *Scope) {
	predeclareNil(universe)
	predeclareTypes(universe)
	predeclareConsts(universe)
	predeclareLibcFuncs(universe)

	universe.setFunc(goidentifier("panic"), &ExprFuncRef{
		funcdef: builtinPanic,
	})
	universe.setFunc(goidentifier("len"), &ExprFuncRef{
		funcdef: builtinLen,
	})
	universe.setFunc(goidentifier("cap"), &ExprFuncRef{
		funcdef: builtinCap,
	})
	universe.setFunc(goidentifier("append"), &ExprFuncRef{
		funcdef: builtinAppend,
	})
	universe.setFunc(goidentifier("makeSlice"), &ExprFuncRef{
		funcdef: builtinMakeSlice,
	})

	universe.setFunc(goidentifier("dumpSlice"), &ExprFuncRef{
		funcdef: builtinDumpSlice,
	})

	universe.setFunc(goidentifier("dumpInterface"), &ExprFuncRef{
		funcdef: builtinDumpInterface,
	})

	universe.setFunc(goidentifier("assertInterface"), &ExprFuncRef{
		funcdef: builtinAssertInterface,
	})

	universe.setFunc(goidentifier("asComment"), &ExprFuncRef{
		funcdef: builtinAsComment,
	})

	universe.setFunc(goidentifier("printstring"), &ExprFuncRef{
		funcdef: builtinPrintstring,
	})

}

// Zero value:
// nil
func predeclareNil(universe *Scope) {
	universe.set(goidentifier("nil"), &IdentBody{
		expr: &ExprNilLiteral{},
	})
}

// Types:
// bool byte complex64 complex128 error float32 float64
// int int8 int16 int32 int64 rune string
// uint uint8 uint16 uint32 uint64 uintptr
func predeclareTypes(universe *Scope) {
	universe.setGtype(goidentifier("bool"), gBool)
	universe.setGtype(goidentifier("byte"), gByte)
	universe.setGtype(goidentifier("int"), gInt)
	universe.setGtype(goidentifier("string"), gString)
	universe.setGtype(goidentifier("uint8"), gByte)
}

// Constants:
// true false iota
func predeclareConsts(universe *Scope) {
	universe.setConst(goidentifier("true"), &ExprConstVariable{
		name:  goidentifier("true"),
		gtype: gBool,
		val:   &ExprNumberLiteral{val: 1},
	})
	universe.setConst(goidentifier("false"), &ExprConstVariable{
		name:  goidentifier("false"),
		gtype: gBool,
		val:   &ExprNumberLiteral{val: 0},
	})

	universe.setConst(goidentifier("iota"), eIota)
}

func predeclareLibcFuncs(universe *Scope) {
	libc := goidentifier(S("libc"))

	universe.setFunc(goidentifier("exit"), &ExprFuncRef{
		funcdef: &DeclFunc{
			pkg: libc,
		},
	})
	universe.setFunc(goidentifier("open"), &ExprFuncRef{
		funcdef: &DeclFunc{
			pkg: libc,
			rettypes: []*Gtype{gInt},
		},
	})

	universe.setFunc(goidentifier("read"), &ExprFuncRef{
		funcdef: &DeclFunc{
			pkg: libc,
			rettypes: []*Gtype{gInt},
		},
	})

	universe.setFunc(goidentifier("write"), &ExprFuncRef{
		funcdef: &DeclFunc{
			pkg: libc,
			rettypes: []*Gtype{gInt},
		},
	})
}
