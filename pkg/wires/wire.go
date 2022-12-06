package wires

import (
	"fmt"

	typ "ixxoprivacy/pkg/types"
)

type WireState byte

const (
	ZERO = iota
	ONE
	UNKNOWN
	UNKNOWN_INVERT
	UNKNOWN_INVERT_OTHER_WIRE
	UNKNOWN_OTHER_WIRE
)

type Wire struct {
	State    WireState
	Number   typ.Num
	Other    *Wire
	Locked   bool
	RefsToMe WireSet
}

type WireSet []*Wire

/*                  Functions on states                 */
/*------------------------------------------------------*/

// ToString converts a WireState (byte) to a string
func (x WireState) ToString() string {
	switch x {
	case ZERO:
		return "ZERO"
	case ONE:
		return "ONE"
	case UNKNOWN:
		return "UNKNOWN"
	case UNKNOWN_INVERT:
		return "UNKNOWN_INVERT"
	case UNKNOWN_OTHER_WIRE:
		return "UNKNOWN_OTHER_WIRE"
	case UNKNOWN_INVERT_OTHER_WIRE:
		return "UNKNOWN_INVERT_OTHER_WIRE"
	default:
		return "Not A State!!!!!"
	}
}

// IntToState converts an int to a WireState
// It is similar to using byte(⋅)
func IntToState(i int) WireState {
	switch i {
	case 0:
		return ZERO
	case 1:
		return ONE
	case 2:
		return UNKNOWN
	case 3:
		return UNKNOWN_INVERT
	case 4:
		return UNKNOWN_INVERT_OTHER_WIRE
	default:
		return UNKNOWN_OTHER_WIRE
	}
}

/*                  Functions on wires                  */
/*------------------------------------------------------*/

// Refs returns the number of references to the given wire
func (w *Wire) Refs() int {
	return len(w.RefsToMe)
}

// AddRef is used so that the given wire is the new reference of
// an other wire w2.
func (w *Wire) AddRef(w2 *Wire) {
	w2.Other = w
	w.RefsToMe = append(w.RefsToMe, w2)
}

func (w *Wire) findRef(w2 *Wire) int {
	for i, r := range w.RefsToMe {
		if r == w2 {
			return i
		}
	}
	fmt.Println("Error in findRef: reference not found.")
	return 0
}

// RemoveRef is to delete a reference from a wire w2 to the given wire.
func (w *Wire) RemoveRef(w2 *Wire) {
	if w2.Other != w {
		fmt.Println("Other wire's other is not this\n")
	}
	i := w.findRef(w2)
	l := len(w.RefsToMe)
	w.RefsToMe[i] = w.RefsToMe[l-1]
	w.RefsToMe = w.RefsToMe[:l-1]
	w2.Other = nil
}

// FreeRefs is used when the given wire references another one
// and we want to delete this dependency.
func (w *Wire) FreeRefs() {
	if w.Other != nil {
		w.Other.RemoveRef(w)
	}
}

// Print is used mainly for debugging purposes and prints every information
// about the given wire to the standard output.
func (w *Wire) Print(indent string) {
	fmt.Println(indent, "Wire: ", w.Number)
	indent += "\t"
	fmt.Println(indent, "State: ", w.State.ToString())
	if w.Other != nil {
		fmt.Println(indent, " ->", w.Other.Number)
	}
	fmt.Println(indent, "Locked: ", w.Locked)
	fmt.Println(indent, "Refs to me: ", w.Refs())
}

/*                         Methods for WireSets                       */
/**********************************************************************/

// NewWireSet returns a new WireSet variable of the given size with defined wires
func EmptyWireSet(size typ.Num) WireSet {
	return make([]*Wire, size)
}

// NewWireSet returns a new WireSet variable of the given size with defined wires
func NewWireSet(size typ.Num) WireSet {
	ws := make([]*Wire, size)
	for i := typ.Num(0); i < size; i++ {
		ws[i] = new(Wire)
	}
	return ws
}

// ReadyToFree is used in FreeIfNoRefs to assess if a set can be freed from the wirepool
func (ws WireSet) ReadyToFree() bool {
	for _, w := range ws {
		if w.Refs() > 0 || w.Locked {
			return false
		}
	}
	return true
}

// PopBack is used to simplify the action of retrieving the last wire
// of the WireSet while removing it from the WireSet.
func (ws *WireSet) PopBack() *Wire {
	tmp := (*ws)[len(*ws)-1]
	*ws = (*ws)[:len(*ws)-1]
	return tmp
}

// PushBack is used to add a wire at the end of a WireSet
func (ws *WireSet) PushBack(w *Wire) {
	*ws = append(*ws, w)
}

// PushBack is used to add a wire at the end of a WireSet
func (ws WireSet) Extend(l typ.Num, w0 *Wire) WireSet {
	for len(ws) < int(l) {
		ws = append(ws, w0)
	}
	return ws
}

// PrintWireSet prints the lists of states of wires of a wire vector
func (ws WireSet) PrintWireSet() {
	for i := len(ws) - 1; i >= 0; i-- {
		if ws[i].State == ONE {
			fmt.Printf("1")
		} else if ws[i].State == ZERO {
			fmt.Printf("0")
		} else {
			fmt.Printf("-")
		}
	}
}

// Considering the set of wires as the representation of an unsigned integer,
// PrintWireSetValue prints the corresponding value
func (ws WireSet) PrintWireSetValue() {
	var val uint64 = 0
	for i := len(ws) - 1; i >= 0; i-- {
		val = val << 1
		if ws[i] == nil {
			fmt.Println("Error: nil wire in PrintWireSetValue")
		} else if ws[i].State == ONE {
			val = val | 0x1
		}
	}
	fmt.Print(val)
}

// IntToWireSet takes an integer as an input and wires representing 0 and 1, and returns
// a wire set corresponding to this integer
func IntToWireSet(v int, w0, w1 *Wire) WireSet {
	if v == 0 {
		return []*Wire{w0}
	}
	ws := NewWireSet(0)
	for v != 0 {
		switch v & 1 {
		case 0:
			ws = append(ws, w0)
		case 1:
			ws = append(ws, w1)
		}
		v = v >> 1
	}
	return ws
}
