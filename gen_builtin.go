package main

import "fmt"

func (e *ExprLen) emit() {
	emit("# emit len()")
	arg := unwrapRel(e.arg)
	gtype := arg.getGtype()
	assert(gtype != nil, e.token(), "gtype should not be  nil:\n")

	switch gtype.getKind() {
	case G_ARRAY:
		emit("LOAD_NUMBER %d", gtype.length)
	case G_SLICE, G_STRING:
		emit("# len(slice|string)")
		switch arg.(type) {
		case *ExprStringLiteral:
			sLiteral := arg.(*ExprStringLiteral)
			length := countStrlen(sLiteral.val)
			emit("LOAD_NUMBER %d", length)
		case *ExprVariable, *ExprStructField, *ExprIndex:
			emitOffsetLoad(arg, 8, ptrSize)
		case *ExprSliceLiteral:
			emit("# ExprSliceLiteral")
			_arg := arg.(*ExprSliceLiteral)
			var length int = len(_arg.values)
			emit("LOAD_NUMBER %d", length)
		case *ExprSlice:
			sliceExpr := arg.(*ExprSlice)
			uop := &ExprBinop{
				op:    "-",
				left:  sliceExpr.high,
				right: sliceExpr.low,
			}
			uop.emit()
		case *IrExprConversion:
			conv := arg.(*IrExprConversion)
			e.arg = conv.arg
			e.emit()
		default:
			bs := fmt.Sprintf("unable to handle %T", arg)
			TBI(arg.token(), string(bs))
		}
	case G_MAP:
		emit("# emit len(map)")
		arg.emit()

		// if not nil
		// then 0
		// else len
		labelNil := makeLabel()
		labelEnd := makeLabel()
		emit("cmpq $0, %%rax # map && map (check if map is nil)")
		emit("je %s # jump if map is nil", labelNil)
		// not nil case
		emit("movq 8(%%rax), %%rax # load map len")
		emit("jmp %s", labelEnd)
		// nil case
		emit("%s:", labelNil)
		emit("LOAD_NUMBER 0")
		emit("%s:", labelEnd)
	default:
		TBI(arg.token(), "unable to handle %s", gtype)
	}
}

type IrLowLevelCall struct {
	token         *Token
	symbol        string
	argsFromStack int // args are taken from the stack
}

func (e *IrLowLevelCall) emit() {
	var i int
	for i = e.argsFromStack - 1; i >= 0; i-- {
		emit("POP_TO_ARG_%d", i)
	}
	emit("FUNCALL %s", e.symbol)
}

func (e *ExprCap) emit() {
	emit("# emit cap()")
	arg := unwrapRel(e.arg)
	gtype := arg.getGtype()
	switch gtype.getKind() {
	case G_ARRAY:
		emit("LOAD_NUMBER %d", gtype.length)
	case G_SLICE:
		switch arg.(type) {
		case *ExprVariable, *ExprStructField, *ExprIndex:
			emitOffsetLoad(arg, 8, ptrSize*2)
		case *ExprSliceLiteral:
			emit("# ExprSliceLiteral")
			_arg := arg.(*ExprSliceLiteral)
			var length int = len(_arg.values)
			emit("LOAD_NUMBER %d", length)
		case *ExprSlice:
			sliceExpr := arg.(*ExprSlice)
			if sliceExpr.collection.getGtype().getKind() == G_ARRAY {
				cp := &ExprBinop{
					tok: e.tok,
					op:  "-",
					left: &ExprLen{
						tok: e.tok,
						arg: sliceExpr.collection,
					},
					right: sliceExpr.low,
				}
				cp.emit()
			} else {
				TBI(arg.token(), "unable to handle %T", arg)
			}
		case *IrExprConversion:
			conv := arg.(*IrExprConversion)
			e.arg = conv.arg
			e.emit()
		default:
			TBI(arg.token(), "unable to handle %T", arg)
		}
	case G_MAP:
		errorft(arg.token(), "invalid argument for cap")
	default:
		TBI(e.token(), "unable to handle %s", gtype.String())
	}
}
