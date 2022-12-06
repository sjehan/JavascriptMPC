package compiler

import (
	"fmt"
	"os"

	typ "ixxoprivacy/pkg/types"
	vb "ixxoprivacy/pkg/variables"
	wr "ixxoprivacy/pkg/wires"

	"github.com/robertkrimen/otto/ast"
	tk "github.com/robertkrimen/otto/token"
)

// outExpressionNode is used to produce outputs for nodes implementing
// the ast.Expression interface.
// Not every king of Expression is accepted in Freegates.
func outExpressionNode(n ast.Expression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outExpressionNode")
	}
	switch exp := n.(type) {
	case *ast.ArrayLiteral:
		return outArrayLiteral(exp, fc)

	case *ast.AssignExpression:
		if call, ok := exp.Right.(*ast.CallExpression); ok {
			if _, res := vb.ReservedFunc[call.Callee.(*ast.Identifier).Name]; !res {
				return outCallAndAssign(exp, fc)
			}
		}
		return outAssignNode(exp, fc)

	case *ast.BinaryExpression:
		switch exp.Operator {
		case tk.OR:
			return outBitwiseORNode(exp, fc)
		case tk.AND:
			return outBitwiseANDNode(exp, fc)
		case tk.EXCLUSIVE_OR:
			return outBitwiseXORNode(exp, fc)
		case tk.PLUS:
			return outArithPlusNode(exp, fc)
		case tk.MINUS:
			return outArithMinusNode(exp, fc)
		case tk.MULTIPLY:
			return outArithMultNode(exp, fc)
		case tk.SLASH:
			return outArithDivNode(exp, fc)
		case tk.REMAINDER:
			return outArithModuloNode(exp, fc)
		case tk.LESS:
			return outConditionalLessNode(exp, fc)
		case tk.GREATER:
			return outConditionalGreaterNode(exp, fc)
		case tk.LESS_OR_EQUAL:
			return outConditionalLessEqualNode(exp, fc)
		case tk.GREATER_OR_EQUAL:
			return outConditionalGreaterEqualNode(exp, fc)
		case tk.EQUAL:
			return outConditionalEqualNode(exp, fc)
		case tk.NOT_EQUAL:
			return outConditionalNotEqualNode(exp, fc)
		case tk.SHIFT_LEFT:
			return outShiftLeftNode(exp, fc)
		case tk.SHIFT_RIGHT:
			return outShiftRightNode(exp, fc)
		case tk.LOGICAL_AND:
			return outLogicalANDNode(exp, fc)
		case tk.LOGICAL_OR:
			return outLogicalORNode(exp, fc)
		}

	case *ast.BooleanLiteral:
		return outBooleanLiteral(exp, fc)

	case *ast.BracketExpression:
		return outBracketExpression(exp, fc)

	case *ast.CallExpression:
		fname := exp.Callee.(*ast.Identifier).Name
		switch fname {
		case "RotateLeft":
			return outRotateLeftNode(exp.ArgumentList[0], exp.ArgumentList[1], fc)
		case "GetWire":
			return outGetWireNode(exp.ArgumentList[0], exp.ArgumentList[1], fc)
		case "SetWire":
			return outSetWireNode(exp.ArgumentList[0], exp.ArgumentList[1], exp.ArgumentList[2], fc)
		default:
			return outCallExpression(exp, fc)
		}

	case *ast.ConditionalExpression:

	case *ast.DotExpression:
		return outDotExpression(exp, fc)

	case *ast.EmptyExpression:

	case *ast.FunctionLiteral:
		return outFunctionLiteral(exp, fc)

	case *ast.Identifier:
		return outIdentifier(exp, fc)

	case *ast.NewExpression:

	case *ast.NullLiteral:

	case *ast.NumberLiteral:
		return outNumberLiteral(exp, fc)

	case *ast.ObjectLiteral:
		return outObjectLiteral(exp, fc)

	case *ast.RegExpLiteral:

	case *ast.SequenceExpression:
		for _, exp2 := range exp.Sequence {
			outExpressionNode(exp2, fc)
		}
		return nil

	case *ast.StringLiteral:

	case *ast.ThisExpression:

	case *ast.UnaryExpression:
		switch exp.Operator {
		case tk.NOT:
			return outUnaryNOTNode(exp, fc)
		case tk.MINUS:
			return outUnaryMinusNode(exp, fc)
		case tk.INCREMENT:
			return outUnaryPostPlusPlusNode(exp, fc)
		case tk.DECREMENT:
			return outUnaryPostMinusMinusNode(exp, fc)
		}

	case *ast.VariableExpression:
		return outVariableExpression(exp, fc)

	}
	fmt.Println("Unused node")
	return nil
}

/*        Binary integer operators           */
/*********************************************/

func auxIntegersOperands(n *ast.BinaryExpression, fc vb.FunctionContext) (t *typ.Type, leftv, rightv vb.IntVariable) {
	leftv = outExpressionNode(n.Left, fc).(vb.IntVariable)
	rightv = outExpressionNode(n.Right, fc).(vb.IntVariable)
	t = typ.MaxType(leftv.GetType(), rightv.GetType())
	return t, leftv, rightv
}

func cleanUpBinaryInt(l, r vb.IntVariable, d vb.VarInterface) {
	if unlockVar(l) {
		pool.FreeSet(l.WSet())
	}
	if unlockVar(r) {
		pool.FreeSet(r.WSet())
	}
	lockVar(d)
}

// outArithPlusNode is used for the output in case of a "+" operator
func outArithPlusNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outArithPlusNode")
	}
	// preparation
	t, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() + rightv.Val())
	}
	destv := vb.NewIntVariable(t, "+OP")
	destv.FillInWires(&pool)

	// outputting the circuit
	outputAddition(leftv.WSet(), rightv.WSet(), destv.Wires)

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outArithMinusNode is used for the output in case of a "-" operator
func outArithMinusNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outArithMinusNode")
	}
	// preparation
	t, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() - rightv.Val())
	}
	destv := vb.NewIntVariable(t, "-OP")
	destv.FillInWires(&pool)

	// outputting the circuit
	outputSubtract(leftv.WSet(), rightv.WSet(), destv.Wires)

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outArithMultNode is used for the output in case of a "*" operator
func outArithMultNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outArithMultNode")
	}
	// preparation
	t, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() * rightv.Val())
	}
	destv := vb.NewIntVariable(t, "×OP")
	destv.FillInWires(&pool)

	// outputting the circuit
	if t.IsIntType() {
		outputMultSigned(leftv.WSet(), rightv.WSet(), destv.Wires)
	} else if t.IsUIntType() {
		outputMultUnsigned(leftv.WSet(), rightv.WSet(), destv.Wires)
	}

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outArithModuloNode is used for the output in case of a "%" operator
func outArithModuloNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outArithModuloNode")
	}
	// preparation
	t, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() % rightv.Val())
	}
	destv := vb.NewIntVariable(t, "%OP")
	destv.FillInWires(&pool)

	// outputting the circuit
	if t.IsIntType() {
		outputDivideSigned(leftv.WSet(), rightv.WSet(), destv.Wires, true)
	} else if t.IsUIntType() {
		outputDivideUnsigned(leftv.WSet(), rightv.WSet(), destv.Wires, true)
	}

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outArithDivNode is used for the output in case of a "/" operator
func outArithDivNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outArithDivNode")
	}
	// preparation
	t, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() / rightv.Val())
	}
	destv := vb.NewIntVariable(t, "÷OP")
	destv.FillInWires(&pool)

	// outputting the circuit
	if t.IsIntType() {
		outputDivideSigned(leftv.WSet(), rightv.WSet(), destv.Wires, false)
	} else if t.IsUIntType() {
		outputDivideUnsigned(leftv.WSet(), rightv.WSet(), destv.Wires, false)
	}

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

/*              Bitwise operators            */
/*********************************************/

func cleanUpAny(l, r, d vb.VarInterface) {
	unlockVar(l)
	unlockVar(r)
	lockVar(d)
	pool.FreeIfNoRefs()
}

// outBitwiseORNode is used for the output in case of a "|" operator
func outBitwiseORNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outBitwiseORNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc)
	rightv := outExpressionNode(n.Right, fc)
	if leftv.IsInt() && leftv.(vb.IntVariable).IsExt() && rightv.IsInt() && rightv.(vb.IntVariable).IsExt() {
		return vb.SimpleExtInt(leftv.(*vb.ExtInt).Val() | rightv.(*vb.ExtInt).Val())
	}

	destv := vb.VarFromType(leftv.GetType(), "|OP")
	destv.FillInWires(&pool)

	var d *wr.Wire
	for i := typ.Num(0); i < leftv.Size(); i++ {
		d = outputGate(14, leftv.GetWire(i), rightv.GetWire(i))
		assignWire(destv.GetWire(i), d)
	}

	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outBitwiseANDNode is used for the output in case of a "&" operator
func outBitwiseANDNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outBitwiseANDNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc)
	rightv := outExpressionNode(n.Right, fc)
	if leftv.IsInt() && leftv.(vb.IntVariable).IsExt() && rightv.IsInt() && rightv.(vb.IntVariable).IsExt() {
		return vb.SimpleExtInt(leftv.(*vb.ExtInt).Val() & rightv.(*vb.ExtInt).Val())
	}

	destv := vb.VarFromType(leftv.GetType(), "&OP")
	destv.FillInWires(&pool)

	var d *wr.Wire
	for i := typ.Num(0); i < leftv.Size(); i++ {
		d = outputGate(8, leftv.GetWire(i), rightv.GetWire(i))
		assignWire(destv.GetWire(i), d)
	}

	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outBitwiseXORNode is used for the output in case of a "^" operator
func outBitwiseXORNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outBitwiseXORNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc)
	rightv := outExpressionNode(n.Right, fc)
	if leftv.IsInt() && leftv.(vb.IntVariable).IsExt() && rightv.IsInt() && rightv.(vb.IntVariable).IsExt() {
		return vb.SimpleExtInt(leftv.(vb.IntVariable).Val() ^ rightv.(vb.IntVariable).Val())
	}

	destv := vb.VarFromType(leftv.GetType(), "^OP")
	destv.FillInWires(&pool)

	var d *wr.Wire
	for i := typ.Num(0); i < leftv.Size(); i++ {
		d = outputGate(6, leftv.GetWire(i), rightv.GetWire(i))
		assignWire(destv.GetWire(i), d)
	}

	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outShiftLeftNode is used for the output in case of a "<<" operator
func outShiftLeftNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outShiftLeftNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc).(vb.IntVariable)
	rightv := outExpressionNode(n.Right, fc).(vb.IntVariable)

	if leftv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() << uint(rightv.Val()))
	}
	destv := vb.NewIntVariable(leftv.GetType(), "<<OP")
	destv.FillInWires(&pool)
	lsize := leftv.Size()
	shift := typ.Num(rightv.Val())

	// shifting
	for i := typ.Num(0); i < shift && i < lsize; i++ {
		assignWire(destv.Wires[i], W_0)
	}
	for i := typ.Num(0); i+shift < lsize; i++ {
		assignWire(destv.Wires[i+shift], leftv.GetWire(i))
		makeWireContainValueNoONEZEROcopy(destv.Wires[i+shift])
	}

	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outRotateLeftNode is used for the output in case of a "<<>" operator
func outRotateLeftNode(left, right ast.Expression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outRotateLeftNode")
	}
	// preparation
	leftv := outExpressionNode(left, fc).(*vb.RegularInt)
	rightv := outExpressionNode(right, fc).(vb.IntVariable)
	destv := vb.NewIntVariable(leftv.GetType(), "<<>OP")
	destv.FillInWires(&pool)

	// rotating
	lsize := int(leftv.Size())
	if leftv.IsConst() {
		for i, w := range leftv.Wires {
			assignWire(destv.Wires[(i+rightv.Val())%lsize], w)
		}
	} else {
		for i, w := range leftv.Wires {
			assignWire(destv.Wires[(i+rightv.Val())%lsize], w)
			makeWireContainValueNoONEZEROcopy(destv.Wires[(i+rightv.Val())%lsize])
		}
	}
	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outShiftRightNode is used for the output in case of a ">>" operator
func outShiftRightNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outShiftRightNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc).(vb.IntVariable)
	rightv := outExpressionNode(n.Right, fc).(vb.IntVariable)

	if leftv.IsExt() {
		return vb.SimpleExtInt(leftv.Val() >> uint(rightv.Val()))
	}
	leftv = leftv.(*vb.RegularInt)
	destv := vb.NewIntVariable(leftv.GetType(), ">>OP")
	destv.FillInWires(&pool)
	lsize := leftv.Size()
	shift := typ.Num(rightv.Val())

	// shifting
	i := typ.Num(0)
	for ; i < lsize-shift; i++ {
		assignWire(destv.Wires[i], leftv.GetWire(i+shift))
		makeWireContainValue(destv.Wires[i])
	}
	for ; i < lsize; i++ {
		assignWire(destv.Wires[i], W_0)
	}

	cleanUpAny(leftv, rightv, destv)
	return destv
}

/*              Logical operators            */
/*********************************************/

// outLogicalORNode is used for the output in case of a "||" operator
func outLogicalORNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outLogicalORNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc).(*vb.BoolVariable)
	rightv := outExpressionNode(n.Right, fc).(*vb.BoolVariable)

	destv := vb.NewBoolVariable("&&OP")
	destv.FillInWires(&pool)

	d := outputGate(14, leftv.GetWire(0), rightv.GetWire(0))
	assignWire(destv.W, d)

	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outLogicalANDNode is used for the output in case of a "&&" operator
func outLogicalANDNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outLogicalANDNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc).(*vb.BoolVariable)
	rightv := outExpressionNode(n.Right, fc).(*vb.BoolVariable)

	destv := vb.NewBoolVariable("||OP")
	destv.FillInWires(&pool)

	d := outputGate(8, leftv.GetWire(0), rightv.GetWire(0))
	assignWire(destv.W, d)

	cleanUpAny(leftv, rightv, destv)
	return destv
}

/*        Binary comparison operators        */
/*********************************************/

// outConditionalLessNode is used for the output in case of a "<" operator
func outConditionalLessNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outConditionalLessNode")
	}
	// preparation
	_, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		if leftv.Val() < rightv.Val() {
			return vb.TrueV
		} else {
			return vb.FalseV
		}
	}
	destv := vb.NewBoolVariable("<OP")

	// outputting the circuit
	destv.W = outputLessThan(leftv.WSet(), rightv.WSet())

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outConditionalGreaterNode is used for the output in case of a ">" operator
func outConditionalGreaterNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outConditionalGreaterNode")
	}
	// preparation
	_, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		if leftv.Val() > rightv.Val() {
			return vb.TrueV
		} else {
			return vb.FalseV
		}
	}
	destv := vb.NewBoolVariable(">OP")

	// outputting the circuit
	destv.W = outputLessThan(rightv.WSet(), leftv.WSet())

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outConditionalLessEqualNode is used for the output in case of a "<=" operator
func outConditionalLessEqualNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outConditionalLessEqualNode")
	}
	// preparation
	_, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		if leftv.Val() <= rightv.Val() {
			return vb.TrueV
		} else {
			return vb.FalseV
		}
	}
	destv := vb.NewBoolVariable("<=OP")

	// outputting the circuit
	// notice the parameter reversal for a > operation
	destv.W = outputLessThan(rightv.WSet(), leftv.WSet())
	destv.W = invertWire(destv.W)

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outConditionalGreaterEqualNode is used for the output in case of a ">=" operator
func outConditionalGreaterEqualNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outConditionalGreaterEqualNode")
	}
	// preparation
	_, leftv, rightv := auxIntegersOperands(n, fc)
	if leftv.IsExt() && rightv.IsExt() {
		if leftv.Val() >= rightv.Val() {
			return vb.TrueV
		} else {
			return vb.FalseV
		}
	}
	destv := vb.NewBoolVariable(">=OP")

	// outputting the circuit
	// notice the parameter reversal for a > operation
	destv.W = outputLessThan(leftv.WSet(), rightv.WSet())
	destv.W = invertWire(destv.W)

	cleanUpBinaryInt(leftv, rightv, destv)
	return destv
}

// outConditionalEqualNode is used for the output in case of a "==" operator
func outConditionalEqualNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outConditionalEqualNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc).(vb.IntVariable)
	rightv := outExpressionNode(n.Right, fc).(vb.IntVariable)
	if leftv.IsExt() && rightv.IsExt() {
		if leftv.Val() == rightv.Val() {
			return vb.TrueV
		} else {
			return vb.FalseV
		}
	}
	// TODO: add operator for arrays and objects
	destv := vb.NewBoolVariable("==OP")

	// outputting the circuit
	destv.W = outputEquals(leftv.WSet(), rightv.WSet())

	cleanUpAny(leftv, rightv, destv)
	return destv
}

// outConditionalNotEqualNode is used for the output in case of a "!=" operator
func outConditionalNotEqualNode(n *ast.BinaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outConditionalNotEqualNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc).(vb.IntVariable)
	rightv := outExpressionNode(n.Right, fc).(vb.IntVariable)
	if leftv.IsExt() && rightv.IsExt() {
		if leftv.Val() != rightv.Val() {
			return vb.TrueV
		} else {
			return vb.FalseV
		}
	}
	// TODO: add operator for arrays and objects
	destv := vb.NewBoolVariable("!=OP")

	// outputting the circuit
	destv.W = outputEquals(leftv.WSet(), rightv.WSet())
	destv.W = invertWire(destv.W)

	cleanUpAny(leftv, rightv, destv)
	return destv
}

/*             Unary operators               */
/*********************************************/

// outUnaryNOTNode is used for the output in case of a "!" operator
func outUnaryNOTNode(n *ast.UnaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outUnaryNOTNode")
	}
	leftv := outExpressionNode(n.Operand, fc)
	if leftv.IsInt() && leftv.(vb.IntVariable).IsExt() {
		return vb.SimpleExtInt(^leftv.(vb.IntVariable).Val())
	}
	destv := vb.VarFromType(leftv.GetType(), "!OP")
	destv.FillInWires(&pool)

	var d1, d2 *wr.Wire
	for i := typ.Num(0); i < leftv.Size(); i++ {
		d1 = leftv.GetWire(i)
		d2 = destv.GetWire(i)
		assignWire(d2, invertWire(d1))
		d1.Locked = false
		d2.Locked = true
	}
	pool.FreeIfNoRefs()
	return destv
}

// outUnaryMinusNode is used for the output in case of a unary "-" operator
func outUnaryMinusNode(n *ast.UnaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outUnaryMinusNode")
	}
	leftv := outExpressionNode(n.Operand, fc).(vb.IntVariable)
	if leftv.IsExt() {
		return vb.SimpleExtInt(-leftv.Val())
	}
	leftv = leftv.(*vb.RegularInt)
	destv := vb.NewIntVariable(leftv.GetType(), "-")
	destv.FillInWires(&pool)

	// outputting the circuit
	outputSubtract(vb.ZeroExt.Wires, leftv.WSet(), destv.Wires)

	unlockVar(leftv)
	destv.Lock()
	pool.FreeIfNoRefs()
	return destv
}

// outUnaryPostPlusPlusNode is used for the output in case of a "++" operator
func outUnaryPostPlusPlusNode(n *ast.UnaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outUnaryPostPlusPlusNode")
	}
	leftv := outExpressionNode(n.Operand, fc).(vb.IntVariable)
	ifvar := fc["-+IFCOND+-"]

	if evl, ok := leftv.(*vb.ExtInt); ok {
		if ifvar != nil {
			fmt.Println("Warning: the use of dollar variables in if conditions may cause errors.")
		}
		evl.ChangeValue(evl.Val() + 1)
		return evl
	}
	ivl := leftv.(*vb.RegularInt)

	destv := vb.NewIntVariable(leftv.GetType(), "++")
	destv.FillInWires(&pool)

	// outputting the circuit
	outputAddition(ivl.Wires, vb.OneExt.Wires, destv.Wires)

	destv.Lock()
	pool.FreeIfNoRefs()
	if ifvar == nil {
		// assign destv to variable (i.e. NOT THE LEFT CORV)
		for i, dw := range destv.Wires {
			assignWire(ivl.Wires[i], dw)
			dw.Locked = false
			makeWireContainValueNoONEZEROcopy(ivl.Wires[i])
		}
	} else {
		cond := ifvar.GetWire(0)
		// assign destv to variable (i.e. NOT THE LEFT CORV)
		for i, dw := range destv.Wires {
			assignWireCond(ivl.Wires[i], dw, cond)
			dw.Locked = false
			makeWireContainValueNoONEZEROcopy(ivl.Wires[i])
		}
	}
	// cleanup
	pool.FreeIfNoRefs()
	return ivl
}

// outUnaryPostMinusMinusNode is used for the output in case of a "--" operator
func outUnaryPostMinusMinusNode(n *ast.UnaryExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outUnaryPostMinusMinusNode")
	}
	leftv := outExpressionNode(n.Operand, fc)
	ifvar := fc["-+IFCOND+-"]

	if evl, ok := leftv.(*vb.ExtInt); ok {
		if ifvar != nil {
			fmt.Println("Warning: the use of dollar variables in if conditions may cause errors.")
		}
		evl.ChangeValue(evl.Val() - 1)
		return evl
	}
	ivl := leftv.(*vb.RegularInt)

	destv := vb.NewIntVariable(leftv.GetType(), "--")
	destv.FillInWires(&pool)

	// outputting the circuit
	outputSubtract(ivl.Wires, vb.OneExt.Wires, destv.Wires)

	destv.Lock()
	pool.FreeIfNoRefs()
	if ifvar == nil {
		// assign destv to variable (i.e. NOT THE LEFT CORV)
		for i, dw := range destv.Wires {
			assignWire(ivl.Wires[i], dw)
			dw.Locked = false
			makeWireContainValueNoONEZEROcopy(ivl.Wires[i])
		}
	} else {
		cond := ifvar.GetWire(0)
		// assign destv to variable (i.e. NOT THE LEFT CORV)
		for i, dw := range destv.Wires {
			assignWireCond(ivl.Wires[i], dw, cond)
			dw.Locked = false
			makeWireContainValueNoONEZEROcopy(ivl.Wires[i])
		}
	}
	// cleanup
	pool.FreeIfNoRefs()
	return ivl
}

/*                  Literals                 */
/*********************************************/

// outArrayLiteral is used to output the CORV representation of an array
func outArrayLiteral(n *ast.ArrayLiteral, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outArrayLiteral")
	}
	t := fc.GetNodeType(n)
	av := vb.NewEmptyArray(t, "GENERATED_ARRAY")
	for i, val := range n.Value {
		av.Av[i] = outExpressionNode(val, fc)
	}
	return av
}

// outObjectLiteral is used to output the CORV representation of an object
func outObjectLiteral(n *ast.ObjectLiteral, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outObjectLiteral")
	}
	t := fc.GetNodeType(n)
	ov := vb.NewEmptyObject(t, "GENERATED_OBJECT")
	for _, prop := range n.Value {
		ov.Map[prop.Key] = outExpressionNode(prop.Value, fc)
	}
	return ov
}

// outFunctionLiteral is used to output the CORV representation of a function
func outFunctionLiteral(n *ast.FunctionLiteral, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outFunctionLiteral")
	}
	// funcvar := context.FunctionContext[n.Name.Name].(*vb.FunctionVariable)

	if n.Body != nil {
		outStatementNode(n.Body, fc)
	}
	pool.FreeIfNoRefs()
	pool.PrintUsedPoolState()
	return nil
}

// outNumberLiteral is used to output the CORV representation of a number
func outNumberLiteral(n *ast.NumberLiteral, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outNumberLiteral")
	}
	name := "NUM_VAR_$$_" + n.Literal
	if fc[name] == nil {
		fc[name] = vb.NewExtInt(vb.GetIntt(), name, int(n.Value.(int64)))
	}
	return fc[name]
}

// outBooleanLiteral is used to output the CORV representation of a boolean
func outBooleanLiteral(n *ast.BooleanLiteral, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outBooleanLiteral")
	}
	if n.Value {
		return vb.TrueV
	}
	return vb.FalseV
}

/*                   Others                  */
/*********************************************/

// outAssignNode is used in case of assignment using "="
func outAssignNode(n *ast.AssignExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outAssignNode")
	}
	// preparation
	leftv := outExpressionNode(n.Left, fc)
	rightv := outExpressionNode(n.Right, fc)
	ifvar := fc["-+IFCOND+-"]

	if evl, ok := leftv.(*vb.ExtInt); ok {
		if ifvar != nil {
			fmt.Println("Warning: the use of dollar variables in if conditions may cause errors.")
		}
		if evr, ok := rightv.(*vb.ExtInt); ok {
			evl.ChangeValue(evr.Val())
			return evl
		} else {
			fmt.Println("Error in outAssignNode: non ext right side assigned to ext right side")
		}
	}

	if leftv == nil {
		fmt.Println("Wire variable is 0. This should not happen (assignnode)")
	}

	if ifvar == nil {
		for i := typ.Num(0); i < leftv.Size(); i++ {
			w1 := leftv.GetWire(i)
			w2 := rightv.GetWire(i)
			assignWire(w1, w2)
			makeWireContainValue(w1)
		}
	} else {
		cond := ifvar.GetWire(0)
		for i := typ.Num(0); i < leftv.Size(); i++ {
			w1 := leftv.GetWire(i)
			w2 := rightv.GetWire(i)
			assignWireCond(w1, w2, cond)
			makeWireContainValueNoONEZEROcopy(w1)
		}
	}
	unlockVar(rightv)
	pool.FreeIfNoRefs()
	return leftv
}

// outBracketExpression is used in case of access to an array
func outBracketExpression(n *ast.BracketExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outBracketExpression")
	}
	arrv := outExpressionNode(n.Left, fc).(*vb.ArrayVariable)
	indv := outExpressionNode(n.Member, fc).(vb.IntVariable)

	if indv.Val() < 0 || indv.Val() >= len(arrv.Av) {
		fmt.Println("Array index out of range for array access. Length is ", len(arrv.Av), " and received index: ", indv.Val())
		os.Exit(64)
	}
	pickedVar := arrv.Av[indv.Val()]
	if unlockVar(indv) {
		pool.FreeSet(indv.WSet())
	}
	if unlockVar(arrv) {
		pool.FreeIfNoRefs()
	}
	return pickedVar
}

// outDotExpression is used in case of access to an object's item
func outDotExpression(n *ast.DotExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outDotExpression")
	}
	//left side must be variable.
	objv := outExpressionNode(n.Left, fc).(*vb.ObjectVariable)
	return objv.Map[n.Identifier.Name]
}

// outCallExpression is used in case of call to a function
func outCallExpression(n *ast.CallExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outCallExpression with callee", n.Callee.(*ast.Identifier).Name)
	}
	id, ok := n.Callee.(*ast.Identifier)
	if !ok {
		fmt.Println("Callee should be identifier, exiting now.")
		os.Exit(64)
	}
	funcvar := context.FunctionContext[id.Name].(*vb.FunctionVariable)

	//|algorithm:
	//--|copy parameters to function paramter slots
	//--|call function
	//--|copy returnv to CORV v and return

	// copy param
	for i, arg := range n.ArgumentList {
		argv := outExpressionNode(arg, fc)
		paramv := funcvar.Argsv[i]
		larg := argv.Size()
		lparam := paramv.Size()

		if argv.IsPerm() {
			writer.AddMassCopy(vb.Wirebase(paramv), vb.Wirebase(argv), typ.Num(larg))
		} else {
			for j := typ.Num(0); j < larg; j++ {
				w := paramv.GetWire(j)
				assignWire(w, argv.GetWire(j))
				makeWireContainValue(w)
			}
			if unlockVar(argv) {
				pool.FreeIfNoRefs()
			}
		}
		if larg < lparam {
			writer.AddReplicate(W_0.Number, paramv.GetWire(larg).Number, typ.Num(lparam-larg))
		}
	}

	writer.AddFunctionCall(funcvar.FunctionNumber)

	//put results intocorv
	if funcvar.Returnv != nil {
		// get or create variable and add to scope
		var rvar vb.VarInterface
		counter := 0

		for true {
			rvar = fc[string(counter)+"-+r+"+id.Name]

			if rvar == nil {
				//create
				name := string(counter) + "-+r+" + id.Name
				rvar = vb.VarFromType(funcvar.Returnv.GetType(), name)
				rvar.FillInWires(&pool)
				rvar.Lock()
				fc[name] = rvar
				break
			} else if !rvar.GetWire(0).Locked {
				break
			}
			counter++
		}
		messyAssignAndCopy(funcvar.Returnv, rvar)
		return rvar
	}
	return nil
}

// outVariableExpression is used in case of variable declaration
func outVariableExpression(n *ast.VariableExpression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outVariableExpression of ", n.Name)
	}

	if v, ok := fc[n.Name]; ok {
		if v.IsInput() {
			// We don't want input value to be initialized, we need them
			// to keep UNKNOWN states.
			return nil
		}
		rv := outExpressionNode(n.Initializer, fc)
		if v.IsInt() && v.(vb.IntVariable).IsExt() {
			if !rv.IsInt() || !rv.(vb.IntVariable).IsExt() {
				fmt.Println("Error in outAssignNode: non ext right side assigned to ext right side")
			} else {
				v.(*vb.ExtInt).ChangeValue(rv.(*vb.ExtInt).Val())
				return v
			}
		}
		messyAssignAndCopy(rv, v)
		rv.Unlock()
	} else {
		fmt.Println("Error in outVariableExpression: unknown variable", n.Name)
	}
	return nil
}

func outIdentifier(n *ast.Identifier, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outIdentifier of ", n.Name)
	}
	if v, ok := fc[n.Name]; ok {
		return v
	} else if v, ok := context.FunctionContext[n.Name]; ok {
		return v
	}
	fmt.Println("Error in outIdentifier: unrecognized identifier ", n.Name)
	return nil
}

// outCallAndAssign is used when we call a function and directly assign the result
// Warning: this is not in Frigate, it might not work in several cases, in particular when used in conditions
func outCallAndAssign(n *ast.AssignExpression, fc vb.FunctionContext) vb.VarInterface {
	callExp := n.Right.(*ast.CallExpression)
	if debug {
		fmt.Println("\tStarting outCallAndAssign with callee", callExp.Callee.(*ast.Identifier).Name)
	}
	id, ok := callExp.Callee.(*ast.Identifier)
	if !ok {
		fmt.Println("Callee should be identifier, exiting now.")
		os.Exit(64)
	}
	funcvar := context.FunctionContext[id.Name].(*vb.FunctionVariable)

	//|algorithm:
	//--|copy parameters to function paramter slots
	//--|call function
	//--|copy returnv to CORV v and return

	for i, arg := range callExp.ArgumentList {
		argv := outExpressionNode(arg, fc)
		paramv := funcvar.Argsv[i]
		larg := argv.Size()
		lparam := paramv.Size()

		if argv.IsPerm() {
			writer.AddMassCopy(vb.Wirebase(paramv), vb.Wirebase(argv), typ.Num(larg))
		} else {
			for j := typ.Num(0); j < larg; j++ {
				w := paramv.GetWire(j)
				assignWire(w, argv.GetWire(j))
				makeWireContainValue(w)
			}
			if unlockVar(argv) {
				pool.FreeIfNoRefs()
			}
		}
		if larg < lparam {
			writer.AddReplicate(W_0.Number, paramv.GetWire(larg).Number, typ.Num(lparam-larg))
		}
	}

	writer.AddFunctionCall(funcvar.FunctionNumber)

	leftv := outExpressionNode(n.Left, fc)

	if funcvar.Returnv != nil {
		ifvar := fc["-+IFCOND+-"]
		if ifvar == nil {
			for i := typ.Num(0); i < leftv.Size(); i++ {
				w1 := leftv.GetWire(i)
				w2 := funcvar.Returnv.GetWire(i)

				if w1.Refs() > 0 && !(w2.Other == w1 && w1.Refs() == 1) {
					clearReffedWire(w1)
				}
				assignWire(w1, w2)
				makeWireContainValueNoONEZEROcopy(w1)
			}
		} else {
			cond := ifvar.GetWire(0)

			for i := typ.Num(0); i < leftv.Size(); i++ {
				w1 := leftv.GetWire(i)
				w2 := funcvar.Returnv.GetWire(i)
				if w1.Refs() > 0 {
					clearReffedWire(w1)
				}
				assignWireCond(w1, w2, cond)
				makeWireContainValueNoONEZEROcopy(w1)
			}
		}
	}
	return leftv
}

// outGetWireNode is used in case of call to built-in function GetWire
func outGetWireNode(left, index ast.Expression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outGetWireNode")
	}
	leftv := outExpressionNode(left, fc)
	indv := outExpressionNode(index, fc).(vb.IntVariable)
	if indv.Val() < 0 {
		fmt.Println("Received negative array index: ", indv.Val())
		os.Exit(64)
	}
	ind := typ.Num(indv.Val())

	if ind >= leftv.Size() {
		fmt.Println("Array index out of range for GetWire. Length is ", leftv.Size(), " and received index: ", indv.Val())
		os.Exit(64)
	}
	v := vb.NewBoolVariable("GENERATED_WIRE_VAR")
	v.W = leftv.GetWire(ind)

	unlockVar(leftv)
	unlockVar(indv)
	pool.FreeIfNoRefs()
	return v
}

// outSetWireNode is used in case of call to built-in function SetWire
func outSetWireNode(left, index, value ast.Expression, fc vb.FunctionContext) vb.VarInterface {
	if debug {
		fmt.Println("\tStarting outSetWireNode")
	}
	leftv := outExpressionNode(left, fc)
	indv := outExpressionNode(index, fc).(vb.IntVariable)
	valuev := outExpressionNode(value, fc)

	if indv.Val() < 0 || typ.Num(indv.Val()) >= leftv.Size() {
		fmt.Println("Array index out of range for GetWire. Length is ", leftv.Size(), " and received index: ", indv.Val())
		os.Exit(64)
	}
	if _, ok := leftv.(*vb.ExtInt); ok {
		fmt.Println("Error in outSetWireNode: cannot change wire of ext integer.")
	}
	w1 := leftv.GetWire(typ.Num(indv.Val()))
	w2 := valuev.GetWire(0)
	assignWire(w1, w2)
	makeWireContainValue(w1)

	unlockVar(indv)
	unlockVar(valuev)
	pool.FreeIfNoRefs()
	return nil
}
