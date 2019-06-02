package main

import "fmt"

// e.g. *x = 1, or *x++
func (uop *ExprUop) emitSave() {
	emit("# *ExprUop.emitSave()")
	assert(uop.op == "*", uop.tok, "uop op should be *")
	emit("PUSH_8")
	uop.operand.emit()
	emit("PUSH_8")
	emit("STORE_8_INDIRECT_FROM_STACK")
}

// e.g. x = 1
func (rel *Relation) emitSave() {
	assert(rel.expr != nil, rel.token(), "left.rel.expr is nil")
	variable := rel.expr.(*ExprVariable)
	variable.emitOffsetSave(variable.getGtype().getSize(), 0, false)
}

func (variable *ExprVariable) emitOffsetSave(size int, offset int, forceIndirection bool) {
	emit("# ExprVariable.emitOffsetSave(size %d, offset %d)", size, offset)
	assert(0 <= size && size <= 8, variable.token(), fmt.Sprintf("invalid size %d", size))
	if variable.getGtype().kind == G_POINTER && (offset > 0 || forceIndirection) {
		assert(variable.getGtype().kind == G_POINTER, variable.token(), "")
		emit("PUSH_8")
		variable.emit()
		emit("ADD_NUMBER %d", offset)
		emit("PUSH_8")
		emit("STORE_8_INDIRECT_FROM_STACK")
		return
	}
	if variable.isGlobal {
		emit("STORE_%d_TO_GLOBAL %s %d", size, variable.varname, offset)
	} else {
		emit("STORE_%d_TO_LOCAL %d+%d", size, variable.offset, offset)
	}
}


func emitAssignPrimitive(left Expr, right Expr) {
	assert(left.getGtype().getSize() <= 8, left.token(), fmt.Sprintf("invalid type for lhs: %s", left.getGtype()))
	assert(right != nil || right.getGtype().getSize() <= 8, right.token(), fmt.Sprintf("invalid type for rhs: %s", right.getGtype()))
	right.emit()   //   expr => %rax
	emitSave(left) //   %rax => memory
}

// Each left-hand side operand must be addressable,
// a map index expression,
// or (for = assignments only) the blank identifier.
func emitSave(left Expr) {
	switch left.(type) {
	case *Relation:
		emit("# %s %s = ", left.(*Relation).name, left.getGtype().String())
		left.(*Relation).emitSave()
	case *ExprIndex:
		left.(*ExprIndex).emitSave()
	case *ExprStructField:
		left.(*ExprStructField).emitSave()
	case *ExprUop:
		left.(*ExprUop).emitSave()
	default:
		left.dump()
		errorft(left.token(), "Unknown case %T", left)
	}
}

// save data from stack
func (e *ExprIndex) emitSave24() {
	// load head address of the array
	// load index
	// multi index * size
	// calc address = head address + offset
	// copy value to the address

	collectionType := e.collection.getGtype()
	switch {
	case collectionType.getKind() == G_ARRAY, collectionType.getKind() == G_SLICE, collectionType.getKind() == G_STRING:
		e.collection.emit() // head address
	case collectionType.getKind() == G_MAP:
		e.emitMapSet(true)
		return
	default:
		TBI(e.token(), "unable to handle %s", collectionType)
	}
	emit("PUSH_8 # head address of collection")
	e.index.emit()
	emit("PUSH_8 # index")
	var elmType *Gtype
	if collectionType.isString() {
		elmType = gByte
	} else {
		elmType = collectionType.elementType
	}
	size := elmType.getSize()
	assert(size > 0, nil, "size > 0")
	emit("LOAD_NUMBER %d # elementSize", size)
	emit("PUSH_8")
	emit("IMUL_FROM_STACK # index * elementSize")
	emit("PUSH_8 # index * elementSize")
	emit("SUM_FROM_STACK # (index * size) + address")
	emit("PUSH_8")
	emit("STORE_24_INDIRECT_FROM_STACK")
}

func (e *ExprIndex) emitSave() {
	collectionType := e.collection.getGtype()
	switch {
	case collectionType.getKind() == G_ARRAY, collectionType.getKind() == G_SLICE, collectionType.getKind() == G_STRING:
		emitCollectIndexSave(e.collection, e.index, 0)
	case collectionType.getKind() == G_MAP:
		emit("PUSH_8") // push RHS value
		e.emitMapSet(false)
		return
	default:
		TBI(e.token(), "unable to handle %s", collectionType)
	}
}

func (e *ExprStructField) emitSave() {
	fieldType := e.getGtype()
	if e.strct.getGtype().kind == G_POINTER {
		emit("PUSH_8 # rhs")

		e.strct.emit()
		emit("ADD_NUMBER %d", fieldType.offset)
		emit("PUSH_8")

		emit("STORE_8_INDIRECT_FROM_STACK")
	} else {
		emitOffsetSave(e.strct, 8, fieldType.offset)
	}
}

func emitOffsetSave(lhs Expr, size int, offset int) {
	switch lhs.(type) {
	case *Relation:
		rel := lhs.(*Relation)
		assert(rel.expr != nil, rel.token(), "left.rel.expr is nil")
		emitOffsetSave(rel.expr, size, offset)
	case *ExprVariable:
		variable := lhs.(*ExprVariable)
		variable.emitOffsetSave(size, offset, false)
	case *ExprStructField:
		structfield := lhs.(*ExprStructField)
		fieldType := structfield.getGtype()
		emitOffsetSave(structfield.strct, size, fieldType.offset+offset)
	case *ExprIndex:
		indexExpr := lhs.(*ExprIndex)
		emitCollectIndexSave(indexExpr.collection, indexExpr.index, offset)

	default:
		errorft(lhs.token(), "unkonwn type %T", lhs)
	}
}

// take slice values from stack
func emitSave24(lhs Expr, offset int) {
	assertInterface(lhs)
	//emit("# emitSave24(%T, offset %d)", lhs, offset)
	emit("# emitSave24(?, offset %d)", offset)
	switch lhs.(type) {
	case *Relation:
		rel := lhs.(*Relation)
		emitSave24(rel.expr, offset)
	case *ExprVariable:
		variable := lhs.(*ExprVariable)
		variable.emitSave24(offset)
	case *ExprStructField:
		structfield := lhs.(*ExprStructField)
		fieldType := structfield.getGtype()
		fieldOffset := fieldType.offset
		emit("# fieldOffset=%d (%s)", fieldOffset, fieldType.fieldname)
		emitSave24(structfield.strct, fieldOffset+offset)
	case *ExprIndex:
		indexExpr := lhs.(*ExprIndex)
		indexExpr.emitSave24()
	default:
		errorft(lhs.token(), "unkonwn type %T", lhs)
	}
}

func (variable *ExprVariable) emitSave24(offset int) {
	emit("# *ExprVariable.emitSave24()")
	emit("pop %%rax # 3rd")
	variable.emitOffsetSave(8, offset+16, false)
	emit("pop %%rax # 2nd")
	variable.emitOffsetSave(8, offset+8, false)
	emit("pop %%rax # 1st")
	variable.emitOffsetSave(8, offset+0, true)
}

func emitCollectIndexSave(collection Expr, index Expr, offset int) {
	collectionType := collection.getGtype()
	assert(collectionType.getKind() == G_ARRAY ||collectionType.getKind() == G_SLICE || collectionType.getKind() == G_STRING, collection.token(), "should be collection")

	var elmType *Gtype
	if collectionType.isString() {
		elmType = gByte
	} else {
		elmType = collectionType.elementType
	}
	elmSize := elmType.getSize()
	assert(elmSize > 0, nil, "elmSize > 0")

	emit("PUSH_8 # rhs")

	collection.emit()
	emit("PUSH_8 # addr")

	index.emit()
	emit("IMUL_NUMBER %d # index * elmSize", elmSize)
	emit("PUSH_8")

	emit("SUM_FROM_STACK # (index * elmSize) + addr")
	emit("ADD_NUMBER %d # offset", offset)
	emit("PUSH_8")

	if elmSize == 1 {
		emit("STORE_1_INDIRECT_FROM_STACK")
	} else {
		emit("STORE_8_INDIRECT_FROM_STACK")
	}
	emitNewline()
}

