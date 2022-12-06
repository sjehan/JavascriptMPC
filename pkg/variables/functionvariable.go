package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"

	"github.com/robertkrimen/otto/ast"
)

type FunctionVariable struct {
	Variable
	Argsv          []VarInterface
	Returnv        VarInterface
	FunctionNumber typ.Num
	FunctionNode   *ast.FunctionLiteral
}

func NewFunctionVariable(f *ast.FunctionLiteral, fc FunctionContext) *FunctionVariable {
	rt := fc.GetReturnType(f.Body)

	fv := &FunctionVariable{
		Variable: Variable{
			Name: f.Name.Name,
			Type: typ.NewFunctionType(rt),
		},
		Returnv:      VarFromType(rt, "@return_var"),
		Argsv:        make([]VarInterface, 0),
		FunctionNode: f,
	}
	for _, param := range f.ParameterList.List {
		v := fc[param.Name]
		fv.AddType(v.GetType())
		fv.Argsv = append(fv.Argsv, v)
	}
	return fv
}

func (fv FunctionVariable) GetWire(i typ.Num) *wr.Wire {
	if fv.Returnv != nil {
		if i < fv.Returnv.Size() {
			return fv.Returnv.GetWire(i)
		} else {
			i -= fv.Returnv.Size()
		}
	}
	for _, v := range fv.Argsv {
		if i < v.Size() {
			return v.GetWire(i)
		} else {
			i -= v.Size()
		}
	}
	fmt.Println("Error in FunctionVariable's GetWire: index given is too large.")
	return nil
}

func (fv *FunctionVariable) AssignPermWires(l typ.Num) typ.Num {
	if fv.Returnv != nil {
		l = fv.Returnv.AssignPermWires(l)
	}
	for _, v := range fv.Argsv {
		l = v.AssignPermWires(l)
	}
	return l
}

func (fv *FunctionVariable) FillInWires(pool *wr.WirePool) {
	if fv.Returnv != nil {
		fv.Returnv.FillInWires(pool)
	}
	for _, v := range fv.Argsv {
		v.FillInWires(pool)
	}
}

func (fv *FunctionVariable) IsFunction() bool {
	return true
}

func (fv *FunctionVariable) Wirebase() typ.Num {
	return fv.GetWire(0).Number
}
