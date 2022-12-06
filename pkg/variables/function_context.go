package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"

	"github.com/robertkrimen/otto/ast"
	tk "github.com/robertkrimen/otto/token"
)

type FunctionContext map[string]VarInterface

// NewFunctionContext returns a new FunctionContext variable
func NewFunctionContext() FunctionContext {
	return make(map[string]VarInterface)
}

// Print sends relevant information about a given FunctionContext to the standard output
func (fc FunctionContext) Print(indent string) {
	for s, v := range fc {
		switch v.(type) {
		case *BoolVariable:
			fmt.Print(indent, "BoolVariable     ", s)
		case *RegularInt:
			fmt.Print(indent, "RegularInt       ", s)
		case *ExtInt:
			fmt.Print(indent, "ExtInt           ", s)
		case *ArrayVariable:
			fmt.Print(indent, "ArrayVariable    ", s)
		case *ObjectVariable:
			fmt.Print(indent, "ObjectVariable   ", s)
		case *FunctionVariable:
			fmt.Print(indent, "FunctionVariable ", s)
		}
		fmt.Print(", ")
		v.GetType().Print("")
		fmt.Println()
	}
	fmt.Println(indent)
}

// addVarDeclaration is a method used to add a new variables to a function context, creating
// it based on a variable declaration node in the abstract syntax tree
func (fc *FunctionContext) addVarDeclaration(vdec *ast.VariableDeclaration) {
	for _, v := range vdec.List {
		// v is a VariableExpression node
		if _, ok := (*fc)[v.Name]; ok {
			fmt.Println("Error in addVarDeclaration: Variable ", v.Name, " is defined twice.")
		} else {
			if v.Initializer == nil {
				fmt.Println("Error in addVarDeclaration: initialization of ", v.Name, "is required.")
			}
			(*fc)[v.Name] = VarFromType(fc.GetNodeType(v.Initializer), v.Name)
		}
	}
}

// CheckForRecTypes checks id there are recursive types defined
// inside the given FunctionContext
func (fc FunctionContext) CheckForRecTypes() {
	for _, v := range fc {
		if v != nil && v.IsObject() {
			arr := make([]*typ.Type, 0)
			typ.CheckRecursiveObj(v.GetType(), arr)
		}
	}
}

// GetNodeType analyses a node of the abstract syntax tree and identify the type of the
// variable which is emitted as a result of this nod
func (fc FunctionContext) GetNodeType(n ast.Node) *typ.Type {
	switch n2 := n.(type) {

	case *ast.BinaryExpression:
		switch n2.Operator {
		case tk.OR, tk.AND, tk.EXCLUSIVE_OR, tk.SHIFT_LEFT, tk.SHIFT_RIGHT, tk.UNSIGNED_SHIFT_RIGHT, tk.AND_NOT:
			return fc.GetNodeType(n2.Left)
		case tk.PLUS, tk.MINUS, tk.MULTIPLY, tk.SLASH, tk.REMAINDER:
			return fc.GetNodeType(n2.Left)
		case tk.LESS, tk.GREATER, tk.LESS_OR_EQUAL, tk.GREATER_OR_EQUAL:
			return GetBoolt()
		case tk.EQUAL, tk.NOT_EQUAL:
			return GetBoolt()
		case tk.LOGICAL_AND, tk.LOGICAL_OR:
			return GetBoolt()
		}

	case *ast.UnaryExpression:
		switch n2.Operator {
		case tk.NOT:
			return typ.BoolType
		default:
			return fc.GetNodeType(n2.Operand)
		}

	case *ast.ArrayLiteral:
		if n2.Value != nil && len(n2.Value) != 0 {
			return typ.NewArrayType(typ.Num(len(n2.Value)), fc.GetNodeType(n2.Value[0]))
		} else {
			return GetVoidType()
		}

	case *ast.BooleanLiteral:
		return GetBoolt()

	case *ast.NumberLiteral:
		return GetIntt()

	case *ast.ObjectLiteral:
		ot := typ.NewObjType()
		for _, prop := range n2.Value {
			ot.AddKeyType(prop.Key, fc.GetNodeType(prop.Value))
		}
		return ot

	case *ast.AssignExpression:
		return fc.GetNodeType(n2.Right)

	case *ast.BracketExpression:
		t := fc.GetNodeType(n2.Left)
		if t.IsArrayType() {
			return t.SubType
		} else {
			return GetVoidType()
		}

	case *ast.CallExpression:
		return fc.GetNodeType(n2.Callee).SubType

	case *ast.DotExpression:
		t := fc.GetNodeType(n2.Left)
		if t.IsObjType() {
			for i, ot := range t.List {
				if t.Keys[i] == n2.Identifier.Name {
					return ot
				}
			}
		}
		return GetVoidType()

	case *ast.Identifier:
		if v, ok := fc[n2.Name]; ok {
			return v.GetType()
		} else if t, ok := ReservedFunc[n2.Name]; ok {
			return t
		}
		return PC.FunctionContext[n2.Name].GetType()
		// TODO: add type conversions
	}
	return GetVoidType()
}

// GetReturnType explores the body of a function in order to find a return
// statement to determine the type of the function
type returnVisitor struct {
	T  *typ.Type
	FC FunctionContext
}

func (rv *returnVisitor) Enter(n ast.Node) ast.Visitor {
	return rv
}
func (rv *returnVisitor) Exit(n ast.Node) {
	if rs, ok := n.(*ast.ReturnStatement); ok {
		rv.T = rv.FC.GetNodeType(rs.Argument)
	}
}

func (fc FunctionContext) GetReturnType(st ast.Node) *typ.Type {
	rv := returnVisitor{GetVoidType(), fc}
	ast.Walk(&rv, st)
	return rv.T
}

// GetParams explores the body of a program in order to find a call to the function of
// name fname and then determine what are the types of arguments given to this function.
type paramVisitor struct {
	FoundCall bool
	FName     string
	FC        *FunctionContext
	IDs       []*ast.Identifier
}

func (pv *paramVisitor) Enter(n ast.Node) ast.Visitor {
	if pv.FoundCall {
		return pv
	}
	if ce, ok := n.(*ast.CallExpression); ok {
		if id, ok := ce.Callee.(*ast.Identifier); ok && id.Name == pv.FName {
			pv.FoundCall = true
			for i, exp := range ce.ArgumentList {
				name := pv.IDs[i].Name
				(*pv.FC)[name] = VarFromType(PC.GetNodeType(exp), name)
			}
			return nil
		}
	}
	return pv
}
func (pv *paramVisitor) Exit(n ast.Node) {}

func (fc *FunctionContext) GetParams(fname string, prog *ast.Program, ids []*ast.Identifier) {
	pv := paramVisitor{false, fname, fc, ids}
	ast.Walk(&pv, prog)

	if !pv.FoundCall {
		fmt.Println("Error: no call found for function", fname)
	}
}
