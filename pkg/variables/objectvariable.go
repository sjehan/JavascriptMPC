package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
	"os"
)

type ObjectVariable struct {
	Variable
	Map map[string]VarInterface
}

func NewObjectVariable(t *typ.Type, name string) *ObjectVariable {
	if !t.IsObjType() {
		fmt.Println("Error in NewObjectVariable: variable initialized with non object type.")
		os.Exit(64)
	}
	ov := ObjectVariable{
		Variable: Variable{
			Name: name,
			Type: t,
		},
		Map: make(map[string]VarInterface),
	}
	for i, oit := range t.List {
		ov.Map[t.Keys[i]] = VarFromType(oit, t.Keys[i])
	}
	return &ov
}

func NewEmptyObject(t *typ.Type, name string) *ObjectVariable {
	if !t.IsObjType() {
		fmt.Println("Error in NewObjectVariable: variable initialized with non object type.")
		os.Exit(64)
	}
	return &ObjectVariable{
		Variable: Variable{
			Name: name,
			Type: t,
		},
		Map: make(map[string]VarInterface),
	}
}

func (obv ObjectVariable) GetWire(i typ.Num) *wr.Wire {
	var s typ.Num
	for j, oit := range obv.List {
		s = oit.Size()
		if i < s {
			return obv.Map[obv.Keys[j]].GetWire(i)
		}
		i -= s
	}
	fmt.Println("Error in ObjectVariable's GetWire: i is too large")
	return nil
}

func (obv ObjectVariable) Print(indent string) {
	indent += obv.GetName() + "."
	for s, v := range obv.Map {
		if v != nil {
			v.Print(indent + s)
		}
	}
}

func (obv *ObjectVariable) SetPerm() {
	obv.isperm = true
	for _, v := range obv.Map {
		v.SetPerm()
	}
}

func (obv *ObjectVariable) AssignPermWires(l typ.Num) typ.Num {
	for _, k := range obv.Keys {
		l = obv.Map[k].AssignPermWires(l)
	}
	return l
}

func (obv *ObjectVariable) FillInWires(pool *wr.WirePool) {
	for _, k := range obv.Keys {
		obv.Map[k].FillInWires(pool)
	}
}

func (obv *ObjectVariable) IsObject() bool {
	return true
}

func (obv *ObjectVariable) Wirebase() typ.Num {
	return obv.GetWire(0).Number
}

func (obv *ObjectVariable) Lock() {
	for _, oit := range obv.Map {
		oit.Lock()
	}
}

func (obv *ObjectVariable) Unlock() {
	for _, oit := range obv.Map {
		oit.Lock()
	}
}
