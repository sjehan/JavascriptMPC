package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
	"os"
)

type ArrayVariable struct {
	Variable
	Av []VarInterface
}

func NewArrayVariable(t *typ.Type, name string) *ArrayVariable {
	if !t.IsArrayType() {
		fmt.Println("Error in NewArrayVariable: variable initialized with non array type.")
		os.Exit(64)
	}
	arv := ArrayVariable{
		Variable: Variable{
			Name: name,
			Type: t,
		},
		Av: make([]VarInterface, t.L),
	}
	for i := typ.Num(0); i < t.L; i++ {
		arv.Av[i] = VarFromType(t.SubType, name+"_item")
	}
	return &arv
}

func NewEmptyArray(t *typ.Type, name string) *ArrayVariable {
	if !t.IsArrayType() {
		fmt.Println("Error in NewArrayVariable: variable initialized with non array type.")
		os.Exit(64)
	}
	return &ArrayVariable{
		Variable: Variable{
			Name: name,
			Type: t,
		},
		Av: make([]VarInterface, t.L),
	}
}

func (arv ArrayVariable) GetWire(i typ.Num) *wr.Wire {
	var s typ.Num
	for _, v := range arv.Av {
		s = v.Size()
		if i < s {
			return v.GetWire(i)
		}
		i -= s
	}
	fmt.Println("Error in ArrayVariable's GetWire: i is too large")
	return nil
}

func (arv *ArrayVariable) Print(indent string) {
	for i, v := range arv.Av {
		if v != nil {
			v.Print(indent + "[" + string(i) + "]")
		}
	}
}

func (arv *ArrayVariable) SetPerm() {
	arv.isperm = true
	for _, v := range arv.Av {
		v.SetPerm()
	}
}

func (arv *ArrayVariable) AssignPermWires(l typ.Num) typ.Num {
	for i := 0; i < len(arv.Av); i++ {
		l = arv.Av[i].AssignPermWires(l)
	}
	return l
}

func (arv *ArrayVariable) FillInWires(pool *wr.WirePool) {
	for _, v := range arv.Av {
		v.FillInWires(pool)
	}
}

func (arv *ArrayVariable) IsArray() bool {
	return true
}

func (arv *ArrayVariable) Lock() {
	for _, v := range arv.Av {
		v.Lock()
	}
}

func (arv *ArrayVariable) Unlock() {
	for _, v := range arv.Av {
		v.Lock()
	}
}
