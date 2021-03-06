package main

// Assignment: a,b,c = expr1,expr2,expr3
func emitAssignMultiToMulti(ast *StmtAssignment) {
	emit("# emitAssignMultiToMulti")
	// The number of operands on the left hand side must match the number of values.
	if len(ast.lefts) != len(ast.rights) {
		errorft(ast.token(), "number of exprs does not match")
	}

	length := len(ast.lefts)
	for i := 0; i < length; i++ {
		right := ast.rights[i]
		left := ast.lefts[i]
		switch right.(type) {
		case *ExprFuncallOrConversion, *ExprMethodcall:
			rettypes := getRettypes(right)
			assert(len(rettypes) == 1, ast.token(), "return values should be one")
		}
		emitAssignOne(left, right)
	}
}

// https://golang.org/ref/spec#Assignments
// A tuple assignment assigns the individual elements of a multi-valued operation to a list of variables.
// There are two forms.
//
// In the first,
// the right hand operand is a single multi-valued expression such as a function call, a channel or map operation, or a type assertion.
// The number of operands on the left hand side must match the number of values.
// For instance, if f is a function returning two values,
//
//	x, y = f()
//
// assigns the first value to x and the second to y.
//
// In the second form,
// the number of operands on the left must equal the number of expressions on the right,
// each of which must be single-valued, and the nth expression on the right is assigned to the nth operand on the left:
//
//  one, two, three = '一', '二', '三'
//

func emitAssignOneRightToMultiLeft(ast *StmtAssignment) {
	var numLeft int = len(ast.lefts)
	emit("# multi(%d) = expr", numLeft)
	// a,b,c = expr
	var numRight int = 0
	right := ast.rights[0]

	var leftsMayBeTwo bool // a(,b) := expr // map index or type assertion
	switch right.(type) {
	case *ExprFuncallOrConversion, *ExprMethodcall:
		rettypes := getRettypes(right)
		numRight += len(rettypes)
	case *ExprTypeAssertion:
		leftsMayBeTwo = true
		numRight++
	case *ExprIndex:
		indexExpr := right.(*ExprIndex)
		if indexExpr.collection.getGtype().getKind() == G_MAP {
			// map get
			emit("# v, ok = map[k]")
			leftsMayBeTwo = true
		}
		numRight++
	default:
		numRight++
	}

	if leftsMayBeTwo {
		if numLeft > 2 {
			errorft(ast.token(), "number of exprs does not match. numLeft=%d", numLeft)
		}
	} else {
		if numLeft != numRight {
			errorft(ast.token(), "number of exprs does not match: %d <=> %d", numLeft, numRight)
		}
	}

	left := ast.lefts[0]
	switch right.(type) {
	case *ExprFuncallOrConversion, *ExprMethodcall:
		rettypes := getRettypes(right)
		if len(rettypes) > 1 {
			// a,b,c = f()
			emit("# a,b,c = f()")
			right.emit()
			var retRegiLen int
			for _, rettype := range rettypes {
				retSize := rettype.getSize()
				if retSize < 8 {
					retSize = 8
				}
				retRegiLen += retSize / 8
			}
			emit("# retRegiLen=%d\n", retRegiLen)
			var i int
			for i = retRegiLen - 1; i >= 0; i-- {
				emit("pushq %%%s # %d", retRegi[i], i)
			}
			for _, left := range ast.lefts {
				if isUnderScore(left) {
					continue
				}
				if left == nil {
					// what is this case ???
					continue
				}
				assert(left.getGtype() != nil, left.token(), "should not be nil")
				emitPop(left.getGtype())
				emitOffsetSave(left, 0)
			}
			return
		}
	}

	emitAssignOne(left, right)
	if leftsMayBeTwo && len(ast.lefts) == 2 {
		okVariable := ast.lefts[1]
		//emit("# lefts[0] type = %s", ast.lefts[0].getGtype().String())
		okRegister := mapOkRegister(ast.lefts[0].getGtype().is24WidthType())
		emit("movq %%%s, %%rax # emit okValue", okRegister)
		emitSavePrimitive(okVariable)
	}
}

func emitAssignOne(lhs Expr, rhs Expr) {
	if lhs == nil {
		// what is this case ???
		return
	}
	gtype := lhs.getGtype()
	switch {
	case gtype == nil:
		// suppose lhs is "_"
		rhs.emit()
	case gtype.getKind() == G_ARRAY:
		assignToArray(lhs, rhs)
	case gtype.getKind() == G_SLICE || gtype.getKind() == G_STRING:
		assignToSlice(lhs, rhs)
	case gtype.getKind() == G_STRUCT:
		assignToStruct(lhs, rhs)
	case gtype.getKind() == G_INTERFACE:
		assignToInterface(lhs, rhs)
	default:
		// suppose primitive
		emitAssignPrimitive(lhs, rhs)
	}
}
func (ast *StmtAssignment) emit() {
	emit("# StmtAssignment")
	// the right hand operand is a single multi-valued expression
	// such as a function call, a channel or map operation, or a type assertion.
	// The number of operands on the left hand side must match the number of values.
	if len(ast.rights) > 1 {
		emitAssignMultiToMulti(ast)
	} else {
		emitAssignOneRightToMultiLeft(ast)
	}
}

func emitAssignPrimitive(lhs Expr, rhs Expr) {
	if rhs == nil {
		if lhs.getGtype().isString() {
			assertNotReached(lhs.token())
		} else {
			// assign zero value
			rhs = &ExprNumberLiteral{}
		}
	}

	assert(lhs.getGtype().getSize() <= 8, lhs.token(), "invalid type for lhs")
	assert(rhs != nil || rhs.getGtype().getSize() <= 8, rhs.token(), "invalid type for rhs")
	rhs.emit()             //   expr => %rax
	emitSavePrimitive(lhs) //   %rax => memory
}

func assignToStruct(lhs Expr, rhs Expr) {
	emit("# assignToStruct start")
	lhs = unwrapRel(lhs)
	assert(rhs == nil || (rhs.getGtype().getKind() == G_STRUCT),
		lhs.token(), "rhs should be struct type")
	// initializes with zero values
	emit("# initialize struct with zero values: start")
	for _, fieldtype := range lhs.getGtype().relation.gtype.fields {
		if fieldtype.is24WidthType() {
			emit("LOAD_EMPTY_24")
			emitSave24(lhs, fieldtype.offset)
			continue
		}
		switch fieldtype.getKind() {
		case G_ARRAY:
			arrayType := fieldtype
			elementType := arrayType.elementType
			elmSize := arrayType.elementType.getSize()
			switch {
			case elementType.getKind() == G_STRUCT:
				left := &ExprStructField{
					strct:     lhs,
					fieldname: fieldtype.fieldname,
				}
				assignToArray(left, nil)
			default:
				assert(0 <= elmSize && elmSize <= 8, lhs.token(), "invalid size")
				for i := 0; i < arrayType.length; i++ {
					emit("LOAD_EMPTY_8")
					emitOffsetSavePrimitive(lhs, elmSize, fieldtype.offset+i*elmSize)
				}
			}
		case G_STRUCT:
			left := &ExprStructField{
				strct:     lhs,
				fieldname: fieldtype.fieldname,
			}
			assignToStruct(left, nil)
		default:
			emit("LOAD_EMPTY_8")
			regSize := fieldtype.getSize()
			assert(0 < regSize && regSize <= 8, lhs.token(), "%s", fieldtype.String())
			emitOffsetSavePrimitive(lhs, regSize, fieldtype.offset)
		}
	}
	emit("# initialize struct with zero values: end")

	if rhs == nil {
		return
	}
	variable := lhs

	strcttyp := rhs.getGtype().Underlying()
	rhs = unwrapRel(rhs)
	switch rhs.(type) {
	case *ExprVariable:
		emitAddress(lhs)
		emit("PUSH_8")
		emitAddress(rhs)
		emit("PUSH_8")
		emitCopyStructFromStack(lhs.getGtype().getSize())
	case *ExprUop:
		re := rhs.(*ExprUop)
		if re.op == "*" {
			// copy struct
			emitAddress(lhs)
			emit("PUSH_8")
			re.operand.emit()
			emit("PUSH_8")
			emitCopyStructFromStack(lhs.getGtype().getSize())
		} else {
			TBI(rhs.token(), "assign to struct")
		}
	case *ExprStructLiteral:
		structliteral, ok := rhs.(*ExprStructLiteral)
		assert(ok || rhs == nil, rhs.token(), "invalid rhs")

		// do assignment for each field
		for _, field := range structliteral.fields {
			emit("# .%s", field.key)
			fieldtype := strcttyp.getField(field.key)

			switch fieldtype.getKind() {
			case G_ARRAY:
				initvalues, ok := field.value.(*ExprArrayLiteral)
				assert(ok, nil, "ok")
				arrayType := strcttyp.getField(field.key)
				elementType := arrayType.elementType
				elmSize := elementType.getSize()
				switch {
				case elementType.getKind() == G_STRUCT:
					left := &ExprStructField{
						strct:     lhs,
						fieldname: fieldtype.fieldname,
					}
					assignToArray(left, field.value)
				default:
					for i, val := range initvalues.values {
						val.emit()
						emitOffsetSavePrimitive(variable, elmSize, arrayType.offset+i*elmSize)
					}
				}
			case G_SLICE, G_STRING:
				left := &ExprStructField{
					tok:       variable.token(),
					strct:     lhs,
					fieldname: field.key,
				}
				assignToSlice(left, field.value)
			case G_INTERFACE:
				left := &ExprStructField{
					tok:       lhs.token(),
					strct:     lhs,
					fieldname: field.key,
				}
				assignToInterface(left, field.value)
			case G_STRUCT:
				left := &ExprStructField{
					tok:       variable.token(),
					strct:     lhs,
					fieldname: field.key,
				}
				assignToStruct(left, field.value)
			default:
				if field.value == nil {
					field.value = &ExprNumberLiteral{}
				}
				field.value.emit()

				regSize := fieldtype.getSize()
				assert(0 < regSize && regSize <= 8, variable.token(), "%s", fieldtype.String())
				emitOffsetSavePrimitive(variable, regSize, fieldtype.offset)
			}
		}
	default:
		TBI(rhs.token(), "assign to struct")
	}

	emit("# assignToStruct end")
}

func assignToInterface(lhs Expr, rhs Expr) {
	emit("# assignToInterface")
	if rhs == nil || isNil(rhs) {
		emit("LOAD_EMPTY_INTERFACE")
		emitSave24(lhs, 0)
		return
	}

	assert(rhs.getGtype() != nil, rhs.token(), "rhs gtype is nil")
	if rhs.getGtype().getKind() == G_INTERFACE {
		rhs.emit()
		emitSave24(lhs, 0)
		return
	}

	emitConversionToInterface(rhs)
	emitSave24(lhs, 0)
}

func assignToSlice(lhs Expr, rhs Expr) {
	emit("# assignToSlice")
	assertInterface(lhs)
	rhs = unwrapRel(rhs)

	switch rhs.(type) {
	case nil:
		emit("LOAD_EMPTY_SLICE")
	case *ExprNilLiteral:
		emit("LOAD_EMPTY_SLICE")
	case *IrExprConversion:
		emit("# IrExprConversion in assignToSlice")
		// https://golang.org/ref/spec#Conversions
		// Converting a value of a string type to a slice of bytes type
		// yields a slice whose successive elements are the bytes of the string.
		//
		// see also https://blog.golang.org/strings
		conversion := rhs.(*IrExprConversion)
		fromExpr := unwrapRel(conversion.arg)
		if fromExpr.getGtype().getKind() == G_SLICE || fromExpr.getGtype().getKind() == G_STRING {
			// emit as it is
			fromExpr.emit()
		} else {
			assertNotReached(lhs.token())
		}
	default:
		//emit("# emit rhs of type %T %s", rhs, rhs.getGtype().String())
		rhs.emit() // it should put values to rax,rbx,rcx
	}

	emitSave24(lhs, 0)
}

// copy each element
func assignToArray(lhs Expr, rhs Expr) {
	rhs = unwrapRel(rhs)
	emit("# assignToArray")
	lhs = unwrapRel(lhs)
	arrayType := lhs.getGtype()
	elementType := arrayType.elementType
	elmSize := elementType.getSize()
	assert(rhs == nil || rhs.getGtype().getKind() == G_ARRAY, nil, "rhs should be array")
	switch elementType.getKind() {
	case G_STRUCT:
		//TBI
		for i := 0; i < arrayType.length; i++ {
			left := &ExprIndex{
				collection: lhs,
				index:      &ExprNumberLiteral{val: i},
			}
			if rhs == nil {
				assignToStruct(left, nil)
				continue
			}
			arrayLiteral, ok := rhs.(*ExprArrayLiteral)
			assert(ok, nil, "ok")
			assignToStruct(left, arrayLiteral.values[i])
		}
		return
	default: // prrimitive type or interface
		for i := 0; i < arrayType.length; i++ {
			offsetByIndex := i * elmSize
			switch rhs.(type) {
			case nil:
				// assign zero values
				if elementType.getKind() == G_INTERFACE {
					emit("LOAD_EMPTY_INTERFACE")
					emitSave24(lhs, offsetByIndex)
					continue
				} else {
					emit("LOAD_EMPTY_8")
				}
			case *ExprArrayLiteral:
				arrayLiteral := rhs.(*ExprArrayLiteral)
				if elementType.getKind() == G_INTERFACE {
					if i >= len(arrayLiteral.values) {
						// zero value
						emit("LOAD_EMPTY_INTERFACE")
						emitSave24(lhs, offsetByIndex)
						continue
					} else if arrayLiteral.values[i].getGtype().getKind() != G_INTERFACE {
						// conversion of dynamic type => interface type
						dynamicValue := arrayLiteral.values[i]
						emitConversionToInterface(dynamicValue)
						emit("LOAD_EMPTY_INTERFACE")
						emitSave24(lhs, offsetByIndex)
						continue
					} else {
						arrayLiteral.values[i].emit()
						emitSave24(lhs, offsetByIndex)
						continue
					}
				}

				if i >= len(arrayLiteral.values) {
					// zero value
					emit("LOAD_EMPTY_8")
				} else {
					val := arrayLiteral.values[i]
					val.emit()
				}
			case *ExprVariable:
				arrayVariable := rhs.(*ExprVariable)
				arrayVariable.emitOffsetLoad(elmSize, offsetByIndex)
			case *ExprStructField:
				strctField := rhs.(*ExprStructField)
				strctField.emitOffsetLoad(elmSize, offsetByIndex)
			default:
				TBI(rhs.token(), "no supporetd %T", rhs)
			}

			emitOffsetSavePrimitive(lhs, elmSize, offsetByIndex)
		}
	}
}

func getRettypes(call Expr) []*Gtype {
	switch call.(type) {
	case *ExprFuncallOrConversion:
		return call.(*ExprFuncallOrConversion).getRettypes()
	case *ExprMethodcall:
		return call.(*ExprMethodcall).getRettypes()
	}
	assertNotReached(call.token())
	return nil
}
