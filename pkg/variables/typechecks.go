package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"

	"github.com/robertkrimen/otto/ast"
	tk "github.com/robertkrimen/otto/token"
)

func CheckProgram(prog *ast.Program) error {
	for _, st := range prog.Body {
		PC.CheckNode(st)
	}
	return nil
}

func (fc FunctionContext) CheckNode(n ast.Node) *typ.Type {
	switch n2 := n.(type) {

	case *ast.BinaryExpression:
		switch n2.Operator {
		case tk.OR, tk.AND, tk.EXCLUSIVE_OR:
			return fc.checkSize(n2.Left, n2.Right, n2.Operator)
		case tk.PLUS, tk.MINUS, tk.MULTIPLY, tk.SLASH, tk.REMAINDER:
			return fc.checkBinaryIntOp(n2.Left, n2.Right, n2.Operator)
		case tk.LESS, tk.GREATER, tk.LESS_OR_EQUAL, tk.GREATER_OR_EQUAL:
			fc.checkBinaryIntOp(n2.Left, n2.Right, n2.Operator)
			return GetBoolt()
		case tk.EQUAL, tk.NOT_EQUAL:
			fc.checkBinarySame(n2.Left, n2.Right, n2.Operator)
			return GetBoolt()
		case tk.SHIFT_LEFT, tk.SHIFT_RIGHT:
			return fc.checkShift(n2.Left, n2.Right, n2.Operator)
		case tk.LOGICAL_AND, tk.LOGICAL_OR:
			fc.checkBool(n2.Left)
			fc.checkBool(n2.Right)
			return GetBoolt()
		}

	case *ast.UnaryExpression:
		switch n2.Operator {
		case tk.NOT:
			return fc.checkBool(n2.Operand)
		default:
			return fc.checkNumber(n2.Operand, n2.Operator)
		}

	case *ast.ArrayLiteral:
		if n2.Value == nil || len(n2.Value) == 0 {
			return GetVoidType()
		} else {
			t := fc.CheckNode(n2.Value[0])
			for i := 1; i < len(n2.Value); i++ {
				fc.checkArrayItem(n2.Value[i], t)
			}
			return typ.NewArrayType(typ.Num(len(n2.Value)), t)
		}

	case *ast.BooleanLiteral:
		return GetBoolt()

	case *ast.NumberLiteral:
		return GetIntt()

	case *ast.ObjectLiteral:
		ot := typ.NewObjType()
		for _, prop := range n2.Value {
			ot.AddKeyType(prop.Key, fc.CheckNode(prop.Value))
		}
		return ot

	case *ast.AssignExpression:
		return fc.checkBinarySame(n2.Left, n2.Right, n2.Operator)

	case *ast.BracketExpression:
		return fc.checkArray(n2)

	case *ast.CallExpression:
		return fc.checkFunctionCall(n2)

	case *ast.ConditionalExpression:
		fc.checkBool(n2.Test)
		fc.CheckNode(n2.Consequent)
		fc.CheckNode(n2.Alternate)
		return GetVoidType()

	case *ast.DotExpression:
		return fc.checkDot(n2)

	case *ast.FunctionLiteral:
		v, ok := fc[n2.Name.Name]
		if !ok {
			return GetVoidType()
		}
		return v.GetType()

	case *ast.Identifier:
		if v, ok := fc[n2.Name]; ok {
			return v.GetType()
		} else if ft, ok := ReservedFunc[n2.Name]; ok {
			return ft
		} else if v, ok := PC.FunctionContext[n2.Name]; ok {
			return v.GetType()
		} else {
			fmt.Println("Error: Unrecognized identifier", n2.Name)
			return GetVoidType()
		}
		// TODO: add type conversions

	case *ast.SequenceExpression:
		for _, exp := range n2.Sequence {
			fc.CheckNode(exp)
		}
		return GetVoidType()

	case *ast.VariableExpression:
		return fc.checkDeclarationVar(n2)

	case *ast.BlockStatement:
		rett := GetVoidType()
		for _, s := range n2.List {
			_, ok := s.(*ast.ReturnStatement)
			tmpt := fc.CheckNode(s)
			if ok {
				if rett != GetVoidType() && rett != tmpt {
					fmt.Println("Error: Multiple return types in block.")
				}
				rett = tmpt
			}
		}
		return rett

	case *ast.ExpressionStatement:
		if n2.Expression == nil {
			return GetVoidType()
		} else {
			return fc.CheckNode(n2.Expression)
		}

	case *ast.ForStatement:
		return fc.checkFor(n2)

	case *ast.FunctionStatement:
		return fc.CheckNode(n2.Function)

	case *ast.IfStatement:
		return fc.checkIf(n2)

	case *ast.ReturnStatement:
		return fc.checkReturn(n2)

	case *ast.VariableStatement:
		for _, exp := range n2.List {
			fc.CheckNode(exp)
		}
		return GetVoidType()
	}
	fmt.Println("Error: node type not recognized")
	return nil
}

func (fc FunctionContext) checkArrayItem(n ast.Expression, t *typ.Type) {
	if !t.Equals(fc.CheckNode(n)) {
		fmt.Println("Error: Array Initialization is not all the same, position ", n.Idx0())
	}
}

func (fc FunctionContext) checkBinaryIntOp(left, right ast.Expression, op tk.Token) *typ.Type {
	leftt := fc.CheckNode(left)
	rightt := fc.CheckNode(right)

	if !leftt.IsIntType() || !leftt.IsIntType() {
		fmt.Println("Error: received type incompatible with integer operation:", typ.Token2string[op])
		leftt.Print("\t")
		fmt.Println()
	}
	if !rightt.IsIntType() || !rightt.IsIntType() {
		fmt.Println("Error: received type incompatible with integer operation:", typ.Token2string[op])
		leftt.Print("\t")
		fmt.Println()
	}

	if leftt.IsUIntType() && rightt.IsIntType() {
		return rightt
	}
	return leftt
}

func (fc FunctionContext) checkBinarySame(left, right ast.Node, op tk.Token) *typ.Type {
	leftt := fc.CheckNode(left)
	rightt := fc.CheckNode(right)

	if !leftt.Equals(rightt) {
		fmt.Println("Error in checkBinarySame: left side and right side are not the same type in operation", typ.Token2string[op])
		leftt.Print("\t")
		rightt.Print("\t")
		fmt.Println()
	}
	return leftt
}

func (fc FunctionContext) checkSize(left, right ast.Expression, op tk.Token) *typ.Type {
	leftt := fc.CheckNode(left)
	if leftt == nil {
		fmt.Println("leftt is nil")
	}
	rightt := fc.CheckNode(right)
	if rightt == nil {
		fmt.Println("rightt is nil")
	}

	if leftt.Size() != rightt.Size() {
		fmt.Println("Error: operation", typ.Token2string[op], "requires both sides to be of same size. Received:")
		fmt.Println("Types:")
		leftt.Print("\t")
		rightt.Print("\t")
		fmt.Println()
	}
	if !leftt.IsIntType() || !leftt.IsIntType() {
		fmt.Println("Error: received type incompatible with integer operation:", typ.Token2string[op])
		leftt.Print("\t")
		fmt.Println()
	}
	if !rightt.IsIntType() || !rightt.IsIntType() {
		fmt.Println("Error: received type incompatible with integer operation:", typ.Token2string[op])
		leftt.Print("\t")
		fmt.Println()
	}
	return leftt
}

func (fc FunctionContext) checkShift(left, right ast.Expression, op tk.Token) *typ.Type {
	leftt := fc.CheckNode(left)
	rightt := fc.CheckNode(right)

	if rightt.IsIntType() && rightt.IsUIntType() {
		fmt.Println("Error: Operator " + typ.Token2string[op] + " cannot be applied, type must be number, recieved:")
		rightt.Print("")
		fmt.Println()
	}
	if leftt.IsIntType() && leftt.IsUIntType() {
		fmt.Println("Error: Operator " + typ.Token2string[op] + " cannot be applied, type must be number, received:")
		leftt.Print("")
		fmt.Println()
	}
	return rightt
}

func (fc FunctionContext) checkNumber(operand ast.Expression, op tk.Token) *typ.Type {
	t := fc.CheckNode(operand)

	if !t.IsIntType() && !t.IsUIntType() {
		fmt.Print("Error: number expected in operation, received")
		t.Print("")
		return GetVoidType()
	}
	return t
}

func (fc FunctionContext) checkFunctionCall(cExp *ast.CallExpression) *typ.Type {
	t := fc.CheckNode(cExp.Callee)
	if !t.IsFunctionType() {
		fmt.Println("Error: Callee is not a function.")
		return GetVoidType()
	}
	if len(cExp.ArgumentList) != len(t.List) {
		fmt.Println("Error: number of arguments is wrong, function should have", len(t.List), "arguments, received", len(cExp.ArgumentList))
		return GetVoidType()
	}

	for i, argExp := range cExp.ArgumentList {
		if !fc.CheckNode(argExp).Equals(t.List[i]) {
			fmt.Println("Error: Function call's parameter types do not match.")
		}
	}
	return t.SubType
}

func (fc FunctionContext) checkDeclarationVar(vexp *ast.VariableExpression) *typ.Type {
	v, ok := fc[vexp.Name]
	if !ok {
		fmt.Println("Error: variable " + vexp.Name + " has no defined type.")
		return GetVoidType()
	}
	t := v.GetType()
	t2 := fc.CheckNode(vexp.Initializer)
	if t2 != GetVoidType() && !t.Equals(t2) {
		fmt.Println("Error: Initializer type not coherent with variable type for " + vexp.Name)
	}
	return t
}

func (fc FunctionContext) checkBool(bn ast.Node) *typ.Type {
	t := fc.CheckNode(bn)
	if !t.IsBoolType() {
		fmt.Print("Error: Wrong type, expected a single wire (bool) type. Got ")
		t.Print("")
	}
	return GetBoolt()
}

func (fc FunctionContext) checkIf(ifn *ast.IfStatement) *typ.Type {
	t := fc.CheckNode(ifn.Test)
	if !t.IsBoolType() {
		fmt.Print("Error: If statement condition is of wrong type, expected a single wire (bool) type. Got ")
		t.Print("")
	}
	if ifn.Consequent != nil {
		fc.CheckNode(ifn.Consequent)
	}
	if ifn.Alternate != nil {
		fc.CheckNode(ifn.Alternate)
	}
	return GetVoidType()
}

func (fc FunctionContext) checkFor(forn *ast.ForStatement) *typ.Type {
	fc.CheckNode(forn.Initializer)

	t := fc.CheckNode(forn.Test)
	if !t.IsBoolType() {
		fmt.Println("Error: For statement condition is of wrong type, expected a single wire (bool) type. Got ")
		t.Print("")
	}
	fc.CheckNode(forn.Update)
	fc.CheckNode(forn.Body)
	return GetVoidType()
}

func (fc FunctionContext) checkDot(dotn *ast.DotExpression) *typ.Type {
	t := fc.CheckNode(dotn.Left)

	if !t.IsObjType() {
		fmt.Println("Error: Dot operator requires ObjType_t, got ")
		t.Print("")
		return GetVoidType()
	}
	for i, k := range t.Keys {
		if k == dotn.Identifier.Name {
			return t.List[i]
		}
	}
	fmt.Println("Error in checkDot: right side is not a part of type ")
	t.Print("")
	return GetVoidType()
}

func (fc FunctionContext) checkArray(aan *ast.BracketExpression) *typ.Type {
	leftt := fc.CheckNode(aan.Left)
	indext := fc.CheckNode(aan.Member)

	if !indext.IsUIntType() && !indext.IsIntType() {
		fmt.Println("Error: In array operator, indexing type is no integer, received ")
		indext.Print("")
	}
	if !leftt.IsArrayType() {
		fmt.Println("Array operator cannot be applied to type ")
		leftt.Print("")
		return GetVoidType()
	}
	return leftt.SubType
}

func (fc FunctionContext) checkReturn(rst *ast.ReturnStatement) *typ.Type {
	return fc.CheckNode(rst.Argument)
	// TODO: add checking like in Frigate if proves necessary
}
