package compiler

import (
	"fmt"
	"os"

	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
	vb "ixxoprivacy/pkg/variables"
	wr "ixxoprivacy/pkg/wires"

	"github.com/robertkrimen/otto/ast"
	tk "github.com/robertkrimen/otto/token"
)

// outStatementNode computes the part of the circuit coming from a part of the code which
// identifies as a segment in the otto package.
func outStatementNode(n ast.Statement, fc vb.FunctionContext) {
	if debug {
		fmt.Println("Starting outStatementNode")
	}
	switch st := n.(type) {
	case *ast.BlockStatement:
		for _, val := range st.List {
			outStatementNode(val, fc)
		}

	case *ast.ExpressionStatement:
		outExpressionNode(st.Expression, fc)

	case *ast.ForInStatement:

	case *ast.ForStatement:
		outForNode(st, fc)

	case *ast.FunctionStatement:
		if debug {
			fmt.Println("FunctionStatement node, not to be treated now.")
		}

	case *ast.IfStatement:
		outIfNode(st, fc)

	case *ast.ReturnStatement:
		outReturnNode(st, fc)

	case *ast.VariableStatement:
		for _, v := range st.List {
			outExpressionNode(v, fc)
		}
	}
}

// outReturnNode deals with return statements.
// If the returned value is the result of an operation it performs it.
// Then it writes the result on the dedicated wires.
func outReturnNode(n *ast.ReturnStatement, fc vb.FunctionContext) {
	if debug {
		fmt.Println("Starting outReturnNode")
	}
	if n.Argument != nil {
		returnv := fc["@return_var"]
		if exp, ok := n.Argument.(*ast.BinaryExpression); ok {
			switch exp.Operator {

			case tk.OR:

				leftv := outExpressionNode(exp.Left, fc)
				rightv := outExpressionNode(exp.Right, fc)

				for i := typ.Num(0); i < leftv.Size(); i++ {
					w := returnv.GetWire(i)
					assignWire(w, outputGate(14, leftv.GetWire(i), rightv.GetWire(i)))
					makeWireContainValue(w)
				}
				leftv.Unlock()
				rightv.Unlock()
				return

			case tk.AND:

				leftv := outExpressionNode(exp.Left, fc)
				rightv := outExpressionNode(exp.Right, fc)

				for i := typ.Num(0); i < leftv.Size(); i++ {
					w := returnv.GetWire(i)
					assignWire(returnv.GetWire(i), outputGate(8, leftv.GetWire(i), rightv.GetWire(i)))
					makeWireContainValue(w)
				}
				leftv.Unlock()
				rightv.Unlock()
				return

			case tk.EXCLUSIVE_OR:

				leftv := outExpressionNode(exp.Left, fc)
				rightv := outExpressionNode(exp.Right, fc)

				for i := typ.Num(0); i < leftv.Size(); i++ {
					w := returnv.GetWire(i)
					assignWire(returnv.GetWire(i), outputGate(6, leftv.GetWire(i), rightv.GetWire(i)))
					makeWireContainValue(w)
				}
				leftv.Unlock()
				rightv.Unlock()
				return

			case tk.PLUS:

				_, leftv, rightv := auxIntegersOperands(exp, fc)
				outputAddition(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires)
				leftv.Unlock()
				rightv.Unlock()
				for _, w := range returnv.(*vb.RegularInt).Wires {
					makeWireContainValue(w)
				}
				return

			case tk.MINUS:

				_, leftv, rightv := auxIntegersOperands(exp, fc)
				outputSubtract(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires)
				leftv.Unlock()
				rightv.Unlock()
				for _, w := range returnv.(*vb.RegularInt).Wires {
					makeWireContainValue(w)
				}
				return

			case tk.MULTIPLY:

				t, leftv, rightv := auxIntegersOperands(exp, fc)
				if t.IsIntType() {
					outputMultSigned(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires)
				} else if t.IsUIntType() {
					outputMultUnsigned(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires)
				}
				leftv.Unlock()
				rightv.Unlock()
				for _, w := range returnv.(*vb.RegularInt).Wires {
					makeWireContainValue(w)
				}
				return

			case tk.SLASH:

				t, leftv, rightv := auxIntegersOperands(exp, fc)
				if t.IsIntType() {
					outputDivideSigned(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires, false)
				} else if t.IsUIntType() {
					outputDivideUnsigned(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires, false)
				}
				leftv.Unlock()
				rightv.Unlock()
				for _, w := range returnv.(*vb.RegularInt).Wires {
					makeWireContainValue(w)
				}
				return

			case tk.REMAINDER:

				t, leftv, rightv := auxIntegersOperands(exp, fc)
				if t.IsIntType() {
					outputDivideSigned(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires, true)
				} else if t.IsUIntType() {
					outputDivideUnsigned(leftv.WSet(), rightv.WSet(), returnv.(*vb.RegularInt).Wires, true)
				}
				leftv.Unlock()
				rightv.Unlock()
				for _, w := range returnv.(*vb.RegularInt).Wires {
					makeWireContainValue(w)
				}
				return

			case tk.LOGICAL_AND:

				_, leftv, rightv := auxIntegersOperands(exp, fc)
				assignWire(returnv.GetWire(0), outputGate(8, leftv.GetWire(0), rightv.GetWire(0)))
				makeWireContainValue(returnv.GetWire(0))
				leftv.Unlock()
				rightv.Unlock()
				return

			case tk.LOGICAL_OR:

				_, leftv, rightv := auxIntegersOperands(exp, fc)
				assignWire(returnv.GetWire(0), outputGate(14, leftv.GetWire(0), rightv.GetWire(0)))
				makeWireContainValue(returnv.GetWire(0))
				leftv.Unlock()
				rightv.Unlock()
				return
			}
		}

		rv := outExpressionNode(n.Argument, fc)
		if ivar, ok := rv.(vb.IntVariable); ok {
			j := typ.Num(0)
			for ; j < ivar.Size() && j < returnv.Size(); j++ {
				w := returnv.GetWire(j)
				assignWire(w, ivar.GetWire(j))
				makeWireContainValue(w)
			}
			for ; j < returnv.Size(); j++ {
				w := returnv.GetWire(j)
				assignWire(w, W_0)
				makeWireContainValue(w)
			}
		} else {
			for j := typ.Num(0); j < rv.Size(); j++ {
				w := returnv.GetWire(j)
				assignWire(w, rv.GetWire(j))
				makeWireContainValue(w)
			}
		}
		unlockVar(rv)
	}
}

// outIfNode deals with the conditional if statements.
// The result of the condition is stored in a wire named cond and then all
// operations in the body of the statement are contioned to this wire.
// Similarly all operations in the "else" part are conditioned to the inverse
// of this wire.
func outIfNode(n *ast.IfStatement, fc vb.FunctionContext) {
	if debug {
		fmt.Println("Starting outIfNode")
	}
	condv := outExpressionNode(n.Test, fc)
	cond := condv.GetWire(0)

	if cond.State == wr.ONE {
		if n.Consequent != nil {
			outStatementNode(n.Consequent, fc)
		}
	} else if cond.State == wr.ZERO {
		if n.Alternate != nil {
			outStatementNode(n.Alternate, fc)
		}
	} else {
		var_x := fc["-+IFCOND+-"]
		iv := vb.NewBoolVariable("-+IFCOND+-")
		var ififcond *wr.Wire
		var prevcond *wr.Wire

		if var_x != nil {
			prevcond = var_x.GetWire(0)
			ififcond = outputGate(8, prevcond, cond)
			ififcond.Locked = true
			iv.W = ififcond
		} else {
			iv.W = cond
			cond.Locked = true
		}
		fc["-+IFCOND+-"] = iv

		if n.Consequent != nil {
			outStatementNode(n.Consequent, fc)
		}
		if n.Alternate != nil {
			cond = invertWire(cond)
			cond.Locked = true
			if var_x == nil {
				iv.W = cond
			} else {
				ififcond = outputGate(8, prevcond, cond)
				ififcond.Locked = true
				iv.W = ififcond
			}
			outStatementNode(n.Alternate, fc)
		}

		if ififcond != nil {
			ififcond.Locked = false
		}
		cond.Locked = false
		if var_x != nil {
			fc["-+IFCOND+-"] = var_x
		} else {
			delete(fc, "-+IFCOND+-")
		}
	}
	unlockVar(condv)
	pool.FreeIfNoRefs()
}

// outIfNode deals with the for loop statements.
// The iteration of the loop depends on a condition which must be a known value,
// i.e. not depend on inputs, so that the total number of iterations is fixed.
func outForNode(n *ast.ForStatement, fc vb.FunctionContext) {
	if debug {
		fmt.Println("Starting outForNode")
	}

	outExpressionNode(n.Initializer, fc)
	condv := outExpressionNode(n.Test, fc)
	cond := condv.GetWire(0)

	if cond.State != wr.ZERO && cond.State != wr.ONE {
		fmt.Println("Conditional Expression in for loop cannot be based on input values when it is checked!")
		os.Exit(64)
	}

	isproc := isProc(n)
	var itr uint32
	var upperFunc *circ.Function
	if isproc {
		if debug {
			fmt.Println("Proc found")
		}
		upperFunc = writer.GetFunction()
		writer.ChangeFunction(circ.NewFunctionPt())
	}

	for cond.State == wr.ONE {
		if !isproc || itr == 0 {
			outStatementNode(n.Body, fc)
		}
		itr++
		outExpressionNode(n.Update, fc)
		condv = outExpressionNode(n.Test, fc)
		cond = condv.GetWire(0)
		if cond.State != wr.ZERO && cond.State != wr.ONE {
			fmt.Println("Conditional Expression in for loop cannot be based on input values when it is check!")
			pool.PrintUsedPoolState()
			os.Exit(64)
		}
	}
	unlockVar(condv)
	if isproc {
		procID := len(circuit.Funcs)
		circuit.Funcs = append(circuit.Funcs, writer.GetFunction())
		writer.ChangeFunction(upperFunc)
		writer.AddProcCall(typ.Num(procID), typ.Num(itr))
	}
	pool.FreeIfNoRefs()
}

// isProc will assess if a given for statement qualifies as a procedure, which
// means that the operations performed are the same at every iteration.
// The condition tested is that the iterative index does not appear in the body.
// But other conditions may have to be verified as well.
type updateVisitor []string

func (uv *updateVisitor) Enter(n ast.Node) ast.Visitor {
	return uv
}
func (uv *updateVisitor) Exit(n ast.Node) {
	if id, ok := n.(*ast.Identifier); ok {
		*uv = append(*uv, id.Name)
	}
}

type procVisitor struct {
	UV   *updateVisitor
	Proc bool
}

func (pv *procVisitor) Enter(n ast.Node) ast.Visitor {
	return pv
}
func (pv *procVisitor) Exit(n ast.Node) {
	if id, ok := n.(*ast.Identifier); ok {
		for _, name := range *pv.UV {
			if name == id.Name {
				pv.Proc = false
			}
		}
	}
}

func isProc(n *ast.ForStatement) bool {
	var ids updateVisitor = make([]string, 0)
	ast.Walk(&ids, n.Update)

	var result procVisitor = procVisitor{&ids, true}
	ast.Walk(&result, n.Body)

	return result.Proc
}
