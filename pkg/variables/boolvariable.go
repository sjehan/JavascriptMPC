package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
)

type BoolVariable struct {
	Variable
	W *wr.Wire
}

func NewBoolVariable(name string) *BoolVariable {
	return &BoolVariable{
		Variable: Variable{
			Name: name,
			Type: GetBoolt(),
		},
	}
}

func (bv *BoolVariable) IsBool() bool {
	return true
}

func (bv *BoolVariable) GetWire(i typ.Num) *wr.Wire {
	if i != 0 {
		fmt.Println("Error in BoolVariable's GetWire: i should be zero")
		return nil
	}
	return bv.W
}

func (bv *BoolVariable) Print(indent string) {
	fmt.Print(indent, "BoolVariable", bv.GetName(), bv.W.Number)
}

func (bv *BoolVariable) AssignPermWires(l typ.Num) typ.Num {
	bv.W.Number = l
	l++
	return l
}

func (bv *BoolVariable) FillInWires(pool *wr.WirePool) {
	if pool == nil {
		bv.W = new(wr.Wire)
	} else {
		bv.W = pool.GetWire()
	}
}

func (bv *BoolVariable) Unlock() {
	bv.W.Locked = false
}

func (bv *BoolVariable) Lock() {
	bv.W.Locked = true
}
