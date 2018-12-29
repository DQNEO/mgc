package main

type Expr interface {
	emit()
	dump()
}

type Stmt interface {
	emit()
}

type ExprNumberLiteral struct {
	val int
}

type ExprStringLiteral struct {
	val    string
	slabel string
}

// local or global variable
type ExprVariable struct {
	varname  identifier
	gtype    *Gtype
	offset   int // for local variable
	isGlobal bool
}

type ExprConstVariable struct {
	name  identifier
	gtype *Gtype
	val   Expr // like ExprConstExpr ?
}

type ExprFuncall struct {
	fname string
	args  []Expr
}

type ExprMethodcall struct {
	receiver Expr
	fname identifier
	args  []Expr
}

type ExprBinop struct {
	op    string
	left  Expr
	right Expr
}

type ExprUop struct {
	op      string
	operand Expr
}

// local or global
type AstVarDecl struct {
	variable *ExprVariable
	initval  Expr
}

type AstConstDecl struct {
	variable *ExprConstVariable
}

type AstAssignment struct {
	lefts  []Expr
	rights []Expr
}

type ForRangeClause struct {
	indexvar  *Relation
	valuevar  *Relation
	rangeexpr Expr
}

type ForForClause struct {
	init Stmt
	cond Stmt
	post Stmt
}

type AstForStmt struct {
	// either of rng or cls is set
	rng   *ForRangeClause
	cls   *ForForClause
	block *AstCompountStmt
}

type AstIfStmt struct {
	cond Expr
	then *AstCompountStmt
	els  Stmt
}

type AstReturnStmt struct {
	expr Expr
}

type AstIncrStmt struct {
	operand Expr
}

type AstDecrStmt struct {
	operand Expr
}

type AstPackageClause struct {
	name identifier
}

type AstImportSpec struct {
	packageName identifier
	path        string
}
type AstImportDecl struct {
	specs []*AstImportSpec
}

type AstCompountStmt struct {
	stmts []Stmt
}

type AstFuncDecl struct {
	receiver  *ExprVariable
	fname     identifier
	rettype   *Gtype
	params    []*ExprVariable
	localvars []*ExprVariable
	body      *AstCompountStmt
}

type AstTopLevelDecl struct {
	// either of followings
	funcdecl  *AstFuncDecl
	vardecl   *AstVarDecl
	constdecl *AstConstDecl
	typedecl  *AstTypeDecl
}

type AstSourceFile struct {
	pkg     *AstPackageClause
	imports []*AstImportDecl
	decls   []*AstTopLevelDecl
}

type AstPackageRef struct {
	name identifier
	path string
}

type AstTypeDecl struct {
	name  identifier
	gtype *Gtype
}

// https://golang.org/ref/spec#Operands
type AstOperandName struct {
	pkg   identifier
	ident identifier
}

type ExprSliced struct {
	ref  *AstOperandName
	low  Expr
	high Expr
}

func (e *ExprSliced) dump() {
	errorf("TBD")
}
func (e *ExprSliced) emit() {
	errorf("TBD")
}

// Expr e.g. array[2]
type ExprArrayIndex struct {
	rel   *Relation
	index Expr
}

func (e *ExprArrayIndex) dump() {
	errorf("TBD")

}

type ExprArrayLiteral struct {
	gtype  *Gtype
	values []Expr
}

func (e ExprArrayLiteral) emit() {
	errorf("DO NOT EMIT")
}
