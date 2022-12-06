package wires

import (
	"fmt"

	typ "ixxoprivacy/pkg/types"
)

/*                         Classes                                    */
/**********************************************************************/

type WirePoolNode struct {
	MapFree map[typ.Num]WireSet
	MapUsed map[typ.Num]WireSet
}

type WirePool struct {
	NextNumber typ.Num
	NodeMap    map[typ.Num]*WirePoolNode
}

/*                         Methods for WirePools                      */
/**********************************************************************/

func NewWirePoolNode() *WirePoolNode {
	wpn := WirePoolNode{
		MapUsed: make(map[typ.Num]WireSet),
		MapFree: make(map[typ.Num]WireSet),
	}
	return &wpn
}

// NewWirePool creates a new WirePool with one WirePoolLLHeadNode
// in its WireSetMap map.
func NewWirePool(nextWire typ.Num) WirePool {
	ws := make(map[typ.Num]*WirePoolNode)
	ws[1] = NewWirePoolNode()
	return WirePool{NodeMap: ws, NextNumber: nextWire}
}

// FreeIfNoRefs free all wires from the given wirepool which are not used
// anymore in the sense that they is not locked or used as reference by
// any other wire.
func (wp *WirePool) FreeIfNoRefs() {
	for _, wpn := range wp.NodeMap {
		for n, ws := range wpn.MapUsed {
			if ws.ReadyToFree() {
				for _, w := range ws {
					w.State = ZERO
					if tmp := w.Other; tmp != nil {
						tmp.RemoveRef(w)
						wp.FreeWire(tmp)
					}
				}
				wpn.MapFree[n] = ws
				delete(wpn.MapUsed, n)
			}
		}
	}
}

// GetWires returns a currently unused set of wires or creates a new set
// of wires which can then be used.
func (wp *WirePool) GetWires(length typ.Num) WireSet {
	var wpn *WirePoolNode = wp.NodeMap[length]
	if wpn == nil {
		wpn = NewWirePoolNode()
		wp.NodeMap[length] = wpn
	}

	if len(wpn.MapFree) != 0 {
		for n, ws := range wpn.MapFree {
			delete(wpn.MapFree, n)
			wpn.MapUsed[n] = ws
			return ws
		}
	}
	ws := NewWireSet(length)
	n := wp.NextNumber
	for j := typ.Num(0); j < length; j++ {
		ws[j] = new(Wire)
		ws[j].Number = wp.NextNumber
		wp.NextNumber++
	}
	wpn.MapUsed[n] = ws
	return ws
}

// GetWire returns a currently unused wire or creates a new
// wire which can then be used
func (wp *WirePool) GetWire() *Wire {
	return wp.GetWires(1)[0]
}

// FreeWire will release a wire, enabling us to use it again later
func (wp *WirePool) FreeWire(w *Wire) {
	if w.Refs() == 0 && !w.Locked {
		wpn1 := wp.NodeMap[1]
		n := w.Number
		if ws, ok := wpn1.MapUsed[n]; ok {
			delete(wpn1.MapUsed, n)
			w.FreeRefs()
			w.State = ZERO
			wpn1.MapFree[n] = ws
		}
	}
}

// FreeSet will release a set of wire that is no more useful,
// so that it can be used again later
func (wp *WirePool) FreeSet(ws WireSet) {
	if ws.ReadyToFree() {
		wpn := wp.NodeMap[typ.Num(len(ws))]
		if wpn != nil {
			n := ws[0].Number
			if ws, ok := wpn.MapUsed[n]; ok {
				delete(wpn.MapUsed, n)
				for _, w := range ws {
					w.State = ZERO
					if tmp := w.Other; tmp != nil {
						tmp.RemoveRef(w)
						wp.FreeWire(tmp)
					}
				}
				wpn.MapFree[n] = ws
			}
		}
	}
}

// FreeSinglesIfNoRefs goes through all the wires used and free them
// if they have no other wire pointing to them
func (wp *WirePool) FreeSinglesIfNoRefs() {
	wpn1 := wp.NodeMap[1]
	if wpn1 != nil {
		for n, ws := range wpn1.MapUsed {
			if ws[0].Refs() == 0 && !ws[0].Locked {
				ws[0].State = ZERO
				if tmp := ws[0].Other; tmp != nil {
					tmp.RemoveRef(ws[0])
					wp.FreeWire(tmp)
				}
				wpn1.MapFree[n] = ws
				delete(wpn1.MapUsed, n)
			}
		}
	}
}

// PrintUsedPoolState is used to check that there are no more used wires
// in the pool after the end of each function output.
// It prints every used wire in the given pool.
func (wp *WirePool) PrintUsedPoolState() {
	for length, wpn := range wp.NodeMap {
		for _, ws := range wpn.MapUsed {
			fmt.Println("\nWire set size: ", length)
			for _, w := range ws {
				w.Print("\t")
			}
		}
	}
}

// PrintFreePoolState prints all free wires in the given pool.
func (wp *WirePool) PrintFreePoolState() {
	fmt.Println("---- Free wires ----")
	for length, wpn := range wp.NodeMap {
		for _, ws := range wpn.MapFree {
			fmt.Println("\nWire set size: ", length)
			for _, w := range ws {
				w.Print("\t")
			}
		}
	}
}
