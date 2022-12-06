package compiler

import (
	"fmt"
	"os"

	typ "ixxoprivacy/pkg/types"
	vb "ixxoprivacy/pkg/variables"
	wr "ixxoprivacy/pkg/wires"

	"github.com/robertkrimen/otto/ast"
)

// wiresToInt returns as an integer the value contained in a variable when a
// constant value (0 or 1) is assigned to every wire of this variable.
func wiresToInt(v vb.VarInterface, errorMessage string, errorNode ast.Node) int {
	var l int = 0

	if v != nil {
		for i := typ.Num(0); i < v.Size(); i++ {
			w := v.GetWire(i)
			if w.State == wr.ONE {
				l = l | (1 << uint(i))
			} else if w.State != wr.ZERO {
				fmt.Println("Error in wiresToInt: non 0/1 wire found: ", errorMessage)
				fmt.Println("Node location: ", errorNode.Idx0())
				os.Exit(64)
			}
		}
	} else {
		fmt.Println("Error in wiresToInt: nil variable provided")
	}
	return l
}

// clearWireForReuse reinitializes a wire to use it again when there is no reference
// to it. Otherwise it takes a new wire from the pool.
func clearWireForReuse(w *wr.Wire) *wr.Wire {
	if w.Refs() > 0 {
		return pool.GetWire()
	}
	w.State = wr.ZERO
	w.FreeRefs()
	return w
}

// makeWireContainValue removed dependencies of the given wire by creating
// the necessary gates so that the value of the wire is directly contained in it.
func makeWireContainValue(w *wr.Wire) {
	if w.State == wr.ONE {
		writer.AddCopy(w.Number, W_1.Number)
	} else if w.State == wr.ZERO {
		writer.AddCopy(w.Number, W_0.Number)
	}

	if w.Other == nil || w.State == wr.UNKNOWN {
		return
	}
	writer.AddCopy(w.Number, w.Other.Number)

	if w.State == wr.UNKNOWN_INVERT_OTHER_WIRE {
		writer.AddGate(6, w.Number, w.Number, W_1.Number)
	}
	w.State = wr.UNKNOWN
	w.Other.RemoveRef(w)
}

// makeWireContainValue removed dependencies of the given wire by creating
// the necessary gates but without considering cases when this values is
// constant equal to 0 or 1.
func makeWireContainValueNoONEZEROcopy(w *wr.Wire) {
	if w.Other == nil || w.State == wr.UNKNOWN {
		if w.State == wr.UNKNOWN_INVERT {
			writer.AddGate(6, w.Number, w.Number, W_1.Number)
			w.State = wr.UNKNOWN
		}
		return
	}
	writer.AddCopy(w.Number, w.Other.Number)

	if w.State == wr.UNKNOWN_INVERT_OTHER_WIRE {
		writer.AddGate(6, w.Number, w.Number, W_1.Number)
	}
	w.State = wr.UNKNOWN
	w.Other.RemoveRef(w)
}

// makeWireNotOther makes sure that the value of a given wire does not
// depend on the value of another wire.
func makeWireNotOther(w *wr.Wire) {
	if w.Other != nil {
		writer.AddCopy(w.Number, w.Other.Number)
		if w.State == wr.UNKNOWN_INVERT_OTHER_WIRE {
			w.State = wr.UNKNOWN_INVERT
		} else {
			w.State = wr.UNKNOWN
		}
	}
	switch w.State {
	case wr.UNKNOWN_INVERT, wr.UNKNOWN_INVERT_OTHER_WIRE:
		w.State = wr.UNKNOWN
		writer.AddGate(6, w.Number, w.Number, W_1.Number)
	case wr.UNKNOWN_OTHER_WIRE:
		w.State = wr.UNKNOWN
	}
}

// invertWire returns a new wire which is the inverted
// version of the given one
func invertWire(w2 *wr.Wire) *wr.Wire {
	var w1 *wr.Wire = pool.GetWire()

	switch w2.State {
	case wr.ONE:
		w1.State = wr.ZERO
	case wr.ZERO:
		w1.State = wr.ONE
	case wr.UNKNOWN:
		w1.State = wr.UNKNOWN_INVERT_OTHER_WIRE
		w2.AddRef(w1)
	case wr.UNKNOWN_OTHER_WIRE:
		w1.State = wr.UNKNOWN_INVERT_OTHER_WIRE
		w2.Other.AddRef(w1)
	case wr.UNKNOWN_INVERT:
		w1.State = wr.UNKNOWN_OTHER_WIRE
		w2.AddRef(w1)
	case wr.UNKNOWN_INVERT_OTHER_WIRE:
		w1.State = wr.UNKNOWN_OTHER_WIRE
		w2.Other.AddRef(w1)
	}
	return w1
}

// invertWireNoInvertOutput returns a new wire which is the inverted
// version of the given one, exepted when the result would be an
// unknown inverted wire. In that cas it adds a gate.
func invertWireNoInvertOutput(w2 *wr.Wire) *wr.Wire {
	var w1 *wr.Wire = pool.GetWire()

	switch w2.State {
	case wr.ONE:
		w1.State = wr.ZERO
	case wr.ZERO:
		w1.State = wr.ONE
	case wr.UNKNOWN:
		addGate(6, w2, W_1, w1)
	case wr.UNKNOWN_OTHER_WIRE:
		addGate(6, w2.Other, W_1, w1)
	case wr.UNKNOWN_INVERT:
		w1.State = wr.UNKNOWN_OTHER_WIRE
		w2.AddRef(w1)
	case wr.UNKNOWN_INVERT_OTHER_WIRE:
		w1.State = wr.UNKNOWN_OTHER_WIRE
		w2.Other.AddRef(w1)
	}
	return w1
}

// invertWireNoAllocUnlessNecessary returns a new wire which is the
// inverted version of the given one and modifies the original wire
// when it is possible in order not to allocate a new one
func invertWireNoAllocUnlessNecessary(w2 *wr.Wire) *wr.Wire {
	if w2.Refs() > 0 {
		return invertWire(w2)
	}
	switch w2.State {
	case wr.ONE:
		w2.State = wr.ZERO
	case wr.ZERO:
		w2.State = wr.ONE
	case wr.UNKNOWN:
		w2.State = wr.UNKNOWN_INVERT
	case wr.UNKNOWN_OTHER_WIRE:
		w2.State = wr.UNKNOWN_INVERT_OTHER_WIRE
	case wr.UNKNOWN_INVERT:
		w2.State = wr.UNKNOWN
	case wr.UNKNOWN_INVERT_OTHER_WIRE:
		w2.State = wr.UNKNOWN_OTHER_WIRE
	}
	return w2
}

// clearReffedWire clears a wire from the references to it and
// copy all of them to a new wire, it also produces a copy output
func clearReffedWire(w *wr.Wire) {
	if w.Refs() == 0 {
		return
	}
	var newwire *wr.Wire = pool.GetWire()
	writer.AddCopy(newwire.Number, w.Number)

	for i := w.Refs() - 1; i >= 0; i-- {
		temp := w.RefsToMe[i]
		w.RemoveRef(temp)
		newwire.AddRef(temp)
	}
	newwire.State = w.State

	if w.Refs() > 0 {
		fmt.Println("refs still > than 0")
	}
}

// assignWire assigns wires from w2 to w1 and deals with other references
func assignWire(w1 *wr.Wire, w2 *wr.Wire) {
	if w1 == w2 {
		return
	}
	if w1.Refs() == 1 && w2.Other == w1 {
		if w2.State == wr.UNKNOWN_OTHER_WIRE {
			w1.State = wr.UNKNOWN
			w1.RemoveRef(w2)
			w2.Other = nil
		} else if w2.State == wr.UNKNOWN_INVERT_OTHER_WIRE {
			w1.State = wr.UNKNOWN_INVERT
			w1.RemoveRef(w2)
			w2.Other = nil
		} else {
			fmt.Println("error in assign wire with refs, strange")
		}
		return
	}
	if w1.Refs() > 0 {
		clearReffedWire(w1)
	}

	// Clear w1
	if w1.State == wr.UNKNOWN_OTHER_WIRE || w1.State == wr.UNKNOWN_INVERT_OTHER_WIRE {
		w1.Other.RemoveRef(w1)
	}

	switch w2.State {
	case wr.ONE:
		w1.State = wr.ONE
	case wr.ZERO:
		w1.State = wr.ZERO
	case wr.UNKNOWN:
		w1.State = wr.UNKNOWN_OTHER_WIRE
		w2.AddRef(w1)
	case wr.UNKNOWN_OTHER_WIRE:
		w1.State = wr.UNKNOWN_OTHER_WIRE
		w2.Other.AddRef(w1)
	case wr.UNKNOWN_INVERT:
		w1.State = wr.UNKNOWN_INVERT_OTHER_WIRE
		w2.AddRef(w1)
	case wr.UNKNOWN_INVERT_OTHER_WIRE:
		w1.State = wr.UNKNOWN_INVERT_OTHER_WIRE
		w2.Other.AddRef(w1)
	}
}

// assignWireCond assigns w2 to w1 if w3 is true
func assignWireCond(w1 *wr.Wire, w2 *wr.Wire, w3 *wr.Wire) {
	if w1 == w2 {
		return
	}
	var xor1o *wr.Wire = outputGate(6, w2, w1)
	var and1o *wr.Wire = outputGate(8, xor1o, w3)

	if w1.Refs() > 0 {
		clearReffedWire(w1)
	} else if w1.Other != nil {
		makeWireContainValue(w1)
	}
	outputGateToDest(6, w1, and1o, w1)
}
