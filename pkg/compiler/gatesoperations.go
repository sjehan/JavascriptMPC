package compiler

import (
	wr "ixxoprivacy/pkg/wires"
)

// We use AddGate when there are unknown values a priori and it is
// necessary to add a gate to the circuit
func addGate(table uint8, a *wr.Wire, b *wr.Wire, dest *wr.Wire) *wr.Wire {
	awirenum := a.Number
	bwirenum := b.Number

	switch a.State {
	case wr.ONE:
		awirenum = W_1.Number
	case wr.ZERO:
		awirenum = W_0.Number
	case wr.UNKNOWN_INVERT_OTHER_WIRE:
		awirenum = a.Other.Number
		table = wr.InvertTable(false, table)
	case wr.UNKNOWN_OTHER_WIRE:
		awirenum = a.Other.Number
	case wr.UNKNOWN_INVERT:
		table = wr.InvertTable(false, table)
	}

	switch b.State {
	case wr.ONE:
		bwirenum = W_1.Number
	case wr.ZERO:
		bwirenum = W_0.Number
	case wr.UNKNOWN_INVERT_OTHER_WIRE:
		bwirenum = b.Other.Number
		table = wr.InvertTable(true, table)
	case wr.UNKNOWN_OTHER_WIRE:
		bwirenum = b.Other.Number
	case wr.UNKNOWN_INVERT:
		table = wr.InvertTable(true, table)
	}

	// If this the program gets here then the value must be unknown.
	// If it was known in any way (or a reference to another wire) it
	// would have been done in the short circuit function
	dest.State = wr.UNKNOWN
	writer.AddGate(table, dest.Number, awirenum, bwirenum)
	return dest
}

// outputGate produces a gate if it is strictly necessary, with regards to the values
// of the wires, and returns the destination wire
func outputGate(table uint8, a *wr.Wire, b *wr.Wire) *wr.Wire {
	var dest *wr.Wire = pool.GetWire()
	if wr.ShortCut(a, b, table, dest) {
		return dest
	}
	return addGate(table, a, b, dest)
}

// outputGateToDest produces a gate if it is strictly necessary, with regards to the values
// of the wires, and writes the result in the wire provided
func outputGateToDest(table uint8, a *wr.Wire, b *wr.Wire, dest *wr.Wire) {
	if wr.ShortCut(a, b, table, dest) {
		return
	}
	addGate(table, a, b, dest)
}

// outputGateToDest produces a gate if it is necessary, with regards to the values of
// the wires and the constraint that the destination wire should not be an inverted value.
// Then it returns the destination wire.
func outputGateNoInvertOutput(table uint8, a *wr.Wire, b *wr.Wire) *wr.Wire {
	var dest *wr.Wire = pool.GetWire()
	if wr.ShortCutNoInvertOutput(a, b, table, dest) {
		return dest
	}
	return addGate(table, a, b, dest)
}

// outputGateNoInvertOutputToDest produces a gate if it is necessary, with regards to the values
// of the wires and the constraint that the destination wire should not be an inverted value.
// Then it writes the result in the wire provided.
func outputGateNoInvertOutputToDest(table uint8, a *wr.Wire, b *wr.Wire, dest *wr.Wire) {
	if wr.ShortCutNoInvertOutput(a, b, table, dest) {
		return
	}
	addGate(table, a, b, dest)
}
