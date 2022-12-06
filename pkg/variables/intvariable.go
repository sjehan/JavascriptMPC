package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
)

type IntVariable interface {
	VarInterface
	IsExt() bool
	Val() int
	WSet() wr.WireSet
}

type RegularInt struct {
	Variable
	Wires wr.WireSet
}

type ExtInt struct {
	RegularInt
	value int
}

/******** Methods and functions for RegularInt *************/

func NewIntVariable(t *typ.Type, name string) *RegularInt {
	if !t.IsIntType() && !t.IsUIntType() {
		fmt.Println("Error in NewIntVariable: variable", name, "initialized with non integer type, received:")
		t.Print("")
	}
	return &RegularInt{
		Variable: Variable{
			Name: name,
			Type: t,
		},
	}
}

func (iv *RegularInt) IsInt() bool {
	return true
}

func (iv *RegularInt) GetWire(i typ.Num) *wr.Wire {
	if i >= typ.Num(len(iv.Wires)) {
		fmt.Println("Error in IntVariable's GetWire: i is too large")
		return nil
	}
	return iv.Wires[i]
}

func (iv *RegularInt) Print(indent string) {
	base := Wirebase(iv)
	fmt.Println(indent, "RegularInt", iv.GetName(), " [", base, ",", base+iv.Size()-1, "]")
}

func (iv *RegularInt) AssignPermWires(l typ.Num) typ.Num {
	for _, w := range iv.Wires {
		w.Number = l
		l++
	}
	return l
}

func (iv *RegularInt) FillInWires(pool *wr.WirePool) {
	if pool == nil {
		iv.Wires = wr.NewWireSet(iv.Size())
	} else {
		iv.Wires = pool.GetWires(iv.Type.Size())
	}
}

func (iv *RegularInt) IsExt() bool {
	return false
}

func (iv *RegularInt) Val() int {
	x := 0
	for i := typ.Num(0); i < iv.Size()-1; i++ {
		if iv.Wires[i].State == wr.ONE {
			x += 1 << i
		} else if iv.Wires[i].State != wr.ZERO {
			fmt.Println("Warning: Val method used with non determined value.")
		}
	}
	if iv.Wires[iv.Size()-1].State == wr.ONE {
		if iv.Type.IsUIntType() {
			x += 1<<iv.Size() - 1
		} else {
			// The negative part
			x -= 1<<iv.Size() - 1
		}
	}
	return x
}

func (iv *RegularInt) Unlock() {
	for _, w := range iv.Wires {
		w.Locked = false
	}
}

func (iv *RegularInt) Lock() {
	for _, w := range iv.Wires {
		w.Locked = true
	}
}

func (iv *RegularInt) WSet() wr.WireSet {
	return iv.Wires
}

/*        Methods and functions for ExtInt         */
/***************************************************/

func NewExtInt(t *typ.Type, name string, val int) *ExtInt {
	if !t.IsIntType() && !t.IsUIntType() {
		fmt.Println("Error in NewIntVariable: variable initialized with non integer type.")
	}
	ei := ExtInt{
		RegularInt: RegularInt{
			Variable: Variable{
				Name:    name,
				Type:    t,
				isconst: true,
			},
		},
	}
	ei.ChangeValue(val)
	return &ei
}

func SimpleExtInt(val int) *ExtInt {
	return NewExtInt(GetIntt(), "", val)
}

func (ev *ExtInt) IsExt() bool {
	return true
}

func (ev *ExtInt) ChangeValue(val int) {
	ev.value = val
	if ev.value >= 0 {
		ev.Wires = wr.IntToWireSet(ev.value, FalseV.W, TrueV.W)
		for len(ev.Wires) < int(ev.Size()) {
			ev.Wires = append(ev.Wires, FalseV.W)
		}
	} else {
		ev.Wires = wr.IntToWireSet(1<<uint(ev.Size())-ev.value, FalseV.W, TrueV.W)
	}
}

func (ev *ExtInt) Print(indent string) {
	fmt.Println(indent, "ExtInt", ev.GetName(), " = ", ev.value, " ; size: ", len(ev.Wires))
}

func (ev *ExtInt) Val() int {
	return ev.value
}
