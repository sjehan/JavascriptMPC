package compiler

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	vb "ixxoprivacy/pkg/variables"
	wr "ixxoprivacy/pkg/wires"
	"strconv"
	"strings"

	"github.com/robertkrimen/otto/ast"
)

// Parse the name of an input or output variable in order to find the
// corresponding party
func getParty(varname string) typ.Num {
	ps := strings.Split(varname, "_")[1]
	p, err := strconv.ParseInt(ps, 10, 8)
	if err != nil {
		fmt.Println("Error in getParty")
		fmt.Println(err)
	}
	return typ.Num(p)
}

// makeONEandZERO initializes the two wires which are supposed to be
// set to values 0 and 1 respectively at all times.
func makeONEandZERO() {
	W_0 = new(wr.Wire)
	W_0.State = wr.ZERO
	W_0.Number = nextBaseWire
	writer.AddGate(0, nextBaseWire, 0, 0)
	nextBaseWire++

	W_1 = new(wr.Wire)
	W_1.State = wr.ONE
	W_1.Number = nextBaseWire
	writer.AddGate(15, nextBaseWire, 0, 0)
	nextBaseWire++
}

// findParameters analyses the AST to find :
// - the number of parties
// - the bit size to use for integers
func findParameters(decList []ast.Declaration) (intsize typ.Num, pnumb uint8) {
	for _, dec := range decList {
		// If it is a variable
		if vdec, ok := dec.(*ast.VariableDeclaration); ok {
			for _, v := range vdec.List {

				if v.Name == "$intsize" {
					vinit, ok := v.Initializer.(*ast.NumberLiteral)
					if !ok {
						fmt.Println("$intsize is no number")
					}
					intsize = typ.Num(vinit.Value.(int64))

				} else if v.Name == "$parties" {
					vinit, ok := v.Initializer.(*ast.NumberLiteral)
					if !ok {
						fmt.Println("$parties is no number")
					}
					pnumb = uint8(vinit.Value.(int64))
				}
			}
		}
	}
	return intsize, pnumb
}

// unlockVar will unlock the set of wires related to a given variable.
// It is useful when this is not a user-defined variable (tested in the if condition).
func unlockVar(v vb.VarInterface) bool {
	if v != nil && !v.IsPerm() && !v.IsConst() {
		v.Unlock()
		return true
	}
	return false
}

// unlockVar will lock the set of wires related to a given variable.
// It is useful when this is not a user-defined variable (tested in the if condition).
func lockVar(v vb.VarInterface) bool {
	if v != nil && !v.IsPerm() && !v.IsConst() {
		v.Lock()
		return true
	}
	return false
}

// messyAssignAndCopy will assign the value of a first variable to a second one
func messyAssignAndCopy(original, copy vb.VarInterface) {
	switch originalT := original.(type) {
	case *vb.BoolVariable:
		copyT := copy.(*vb.BoolVariable)
		assignWire(copyT.W, originalT.W)
		makeWireContainValue(copyT.W)

	case vb.IntVariable:
		copyT := copy.(*vb.RegularInt)
		intvsize := originalT.Size()
		for i, dw := range copyT.Wires {
			if typ.Num(i) < intvsize {
				assignWire(dw, originalT.GetWire(typ.Num(i)))
				makeWireContainValue(dw)
			} else {
				assignWire(dw, W_0)
			}
		}

	case *vb.ArrayVariable:
		copyT := copy.(*vb.ArrayVariable)
		for i, dv := range copyT.Av {
			messyAssignAndCopy(originalT.Av[i], dv)
		}

	case *vb.ObjectVariable:
		copyT := copy.(*vb.ObjectVariable)
		for k, v := range originalT.Map {
			messyAssignAndCopy(v, copyT.Map[k])
		}

	default:
		fmt.Println("undefined type in messy messyAssignAndCopy")
	}
}
