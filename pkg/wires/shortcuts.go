package wires

import "fmt"

/*
Tables
0 - 0000 (0)
1 - 0001 (nor)
2 - 0010
3 - 0011 (invert passthrough a)
4 - 0100
5 - 0101 (invert passthorugh b)
6 - 0110 (xor)
7 - 0111 (nand)
8 - 1000 (and)
9 - 1001 (reverse xor)
10 - 1010 (passthrough b)
11 - 1011
12 - 1100 (passthrough a)
13 - 1101
14 - 1110 (or)
15 - 1111 (1)
*/

// This function inverts the table for one of the two entries
// aOrb - true is b, false is a
// Let f be the boolean function whose mapping is given by the table variable.
// Let us suppose that aOrb is false, i.e. we want to invert the table for the
// first variable.
// Then the output is the table of a new boolean function f' such that,
// for all a, b in {0,1}, f'(a,b) = f(1 - a,b).
func InvertTable(aOrb bool, table uint8) uint8 {
	if aOrb {
		table = ((table << 1) & 10) | ((table >> 1) & 5) //((table <<2)&12) | ((table >>2)&3)
	} else {
		table = ((table << 2) & 12) | ((table >> 2) & 3)
	}
	return table
}

// This function computes the operation relative to a table with two input wires.
// The table variable represents a gate, a and b are the two input wires and
// the goal is to change the wire dest in order to make it be the result of
// the gate.
func ShortCut(a *Wire, b *Wire, table uint8, dest *Wire) bool {
	if dest.Other != nil && dest != a && dest != b {
		fmt.Println("shoft circuit other != NIL", dest.Number, " -> ", dest.Other.Number)
	}
	if len(dest.RefsToMe) != 0 {
		fmt.Println("shoft circuit refs != 0, length is ", len(dest.RefsToMe))
	}

	// if  gate is constant we evaluate
	if (a.State == ONE || a.State == ZERO) && (b.State == ONE || b.State == ZERO) {
		var wa uint8 = 0
		if a.State == ONE {
			wa = 1
		}
		var wb uint8 = 0
		if b.State == ONE {
			wb = 1
		}
		var entry uint8 = (wa << 1) | wb

		res := (table >> entry) & 1
		dest.State = ZERO
		if res == 1 {
			dest.State = ONE
		}
		return true
	}

	// Trivial short circuits
	if table == 0 {
		dest.State = ZERO
		return true
	}
	if table == 15 {
		dest.State = ONE
		return true
	}

	if a.State == UNKNOWN_INVERT {
		table = InvertTable(false, table)
	}
	if b.State == UNKNOWN_INVERT {
		table = InvertTable(true, table)
	}

	// invert passthrough a
	if table == 3 {
		return invertPassthrough(a, dest)
	}

	// invert passthrough b
	if table == 5 {
		return invertPassthrough(b, dest)
	}

	// passthrough b
	if table == 10 {
		return passthrough(b, dest)
	}

	// passthrough a
	if table == 12 {
		return passthrough(a, dest)
	}

	//slightly less trivial short circuits | if  one OR value is one
	if table == 14 && (a.State == ONE || b.State == ONE) {
		dest.State = ONE
		return true
	}

	// if  one AND value is zero
	if table == 8 && (a.State == ZERO || b.State == ZERO) {
		dest.State = ZERO
		return true
	}

	//one NOR value is one
	if table == 1 && (a.State == ONE || b.State == ONE) {
		dest.State = ZERO
		return true
	}

	//one NAND value is zero
	if table == 7 && (a.State == ZERO || b.State == ZERO) {
		dest.State = ONE
		return true
	}

	if a.State == ONE {
		option0 := (table>>2)&1 == 1
		option1 := (table>>3)&1 == 1
		return oneIsConst(b, option0, option1, dest)

	} else if a.State == ZERO {
		option0 := (table>>0)&1 == 1
		option1 := (table>>1)&1 == 1
		return oneIsConst(b, option0, option1, dest)

	} else if b.State == ONE {
		option0 := (table>>1)&1 == 1
		option1 := (table>>3)&1 == 1
		return oneIsConst(a, option0, option1, dest)

	} else if b.State == ZERO {
		option0 := (table>>0)&1 == 1
		option1 := (table>>2)&1 == 1
		return oneIsConst(a, option0, option1, dest)
	}

	// tables 0,3,5,10,12,15 already done at this point in time)
	// 2,4,11,13 cannot be optimized in general (i think)
	// leaving 1 6 7 8 9 14

	if a == b {
		switch table {
		case 6, 2, 4: // put 0
			dest.State = ZERO

		case 9, 11, 13: // put 1
			dest.State = ONE

		case 14, 8: // put a
			if dest == a {
				switch dest.State {
				case UNKNOWN, UNKNOWN_INVERT:
					dest.State = UNKNOWN
				case UNKNOWN_OTHER_WIRE:
					dest.State = UNKNOWN_OTHER_WIRE
				case UNKNOWN_INVERT_OTHER_WIRE:
					dest.State = UNKNOWN_INVERT_OTHER_WIRE
				}
				return true
			}

			switch a.State {
			case UNKNOWN, UNKNOWN_INVERT:
				dest.State = UNKNOWN_OTHER_WIRE
				dest.Other = a
			default:
				dest.State = a.State
				dest.Other = a.Other
			}
			dest.Other.AddRef(dest)

		case 1, 7: // put invert a
			if dest == a {
				switch dest.State {
				case UNKNOWN, UNKNOWN_INVERT:
					dest.State = UNKNOWN_INVERT
				case UNKNOWN_OTHER_WIRE:
					dest.State = UNKNOWN_INVERT_OTHER_WIRE
				case UNKNOWN_INVERT_OTHER_WIRE:
					dest.State = UNKNOWN_OTHER_WIRE
				}
				return true
			}

			switch a.State {
			case UNKNOWN, UNKNOWN_INVERT:
				dest.State = UNKNOWN_INVERT_OTHER_WIRE
				dest.Other = a
			case UNKNOWN_OTHER_WIRE:
				dest.State = UNKNOWN_INVERT_OTHER_WIRE
				dest.Other = a.Other
			case UNKNOWN_INVERT_OTHER_WIRE:
				dest.State = UNKNOWN_OTHER_WIRE
				dest.Other = a.Other
			}

			dest.Other.AddRef(dest)
		}
		return true
	}
	return false
}

// This function computes the operation relative to a table with two input
// wires, but do not enable inverted ouputs.
// The table variable represents a gate, a and b are the two input wires and
// the goal is to change the wire dest in order to make it be the result of
// the gate.
func ShortCutNoInvertOutput(a *Wire, b *Wire, table uint8, dest *Wire) bool {
	if dest.Other != nil {
		fmt.Println("shoft circuit other != NIL")
		fmt.Println("is ", dest.Other)
	}
	if len(dest.RefsToMe) != 0 {
		fmt.Println("shoft circuit refs != 0")
		fmt.Println("is ", len(dest.RefsToMe))
	}

	// if  gate is constant we evaluate
	if (a.State == ONE || a.State == ZERO) && (b.State == ONE || b.State == ZERO) {
		var wa uint8 = 0
		if a.State == ONE {
			wa = 1
		}
		var wb uint8 = 0
		if b.State == ONE {
			wb = 1
		}
		var entry uint8 = (wa << 1) | wb

		res := (table >> entry) & 1
		dest.State = ZERO
		if res == 1 {
			dest.State = ONE
		}
		return true
	}

	// Trivial short circuits
	if table == 0 {
		dest.State = ZERO
		return true
	}
	if table == 15 {
		dest.State = ONE
		return true
	}

	if a.State == UNKNOWN_INVERT {
		table = InvertTable(false, table)
	}
	if b.State == UNKNOWN_INVERT {
		table = InvertTable(true, table)
	}

	// invert passthrough a
	if table == 3 {
		return invertPassthroughNoInvertOutput(a, dest)
	}

	// invert passthrough b
	if table == 5 {
		return invertPassthroughNoInvertOutput(b, dest)
	}

	// passthrough b
	if table == 10 {
		return passthroughNoInvertOutput(b, dest)
	}

	// passthrough a
	if table == 12 {
		return passthroughNoInvertOutput(a, dest)
	}

	// slightly less trivial short circuits | if  one OR value is one
	if table == 14 && (a.State == ONE || b.State == ONE) {
		dest.State = ONE
		return true
	}

	// if  one AND value is zero
	if table == 8 && (a.State == ZERO || b.State == ZERO) {
		dest.State = ZERO
		return true
	}

	// one NOR value is one
	if table == 1 && (a.State == ONE || b.State == ONE) {
		dest.State = ZERO
		return true
	}

	// one NAND value is zero
	if table == 7 && (a.State == ZERO || b.State == ZERO) {
		dest.State = ONE
		return true
	}

	if a.State == ONE {
		option0 := (table>>2)&1 == 1
		option1 := (table>>3)&1 == 1
		return oneIsConstNoInvertOutput(b, option0, option1, dest)
	}
	if a.State == ZERO {
		option0 := (table>>0)&1 == 1
		option1 := (table>>1)&1 == 1
		return oneIsConstNoInvertOutput(b, option0, option1, dest)

	}
	if b.State == ONE {
		option0 := (table>>1)&1 == 1
		option1 := (table>>3)&1 == 1
		return oneIsConstNoInvertOutput(a, option0, option1, dest)
	}
	if b.State == ZERO {
		option0 := (table>>0)&1 == 1
		option1 := (table>>2)&1 == 1
		return oneIsConstNoInvertOutput(a, option0, option1, dest)
	}

	//tables 0,3,5,10,12,15 already done at this point in time)
	//2,4,11,13 cannot be optimized in general (i think)
	//leaving 1 6 7 8 9 14

	if a == b {
		switch table {
		case 6, 2, 4:
			dest.State = ZERO

		case 9, 11, 13:
			dest.State = ONE

		case 14, 8: // put a
			if dest == a {
				switch dest.State {
				case UNKNOWN, UNKNOWN_INVERT:
					dest.State = UNKNOWN
				case UNKNOWN_OTHER_WIRE:
					dest.State = UNKNOWN_OTHER_WIRE
				case UNKNOWN_INVERT_OTHER_WIRE:
					return false
				}
				return true
			}
			switch a.State {
			case UNKNOWN:
				dest.State = UNKNOWN_OTHER_WIRE
				dest.Other = a
			case UNKNOWN_INVERT:
				dest.State = UNKNOWN_OTHER_WIRE
				dest.Other = a
			default:
				dest.State = a.State
				dest.Other = a.Other
			}

		case 1, 7: // put invert a
			if dest == a {
				switch dest.State {
				case UNKNOWN, UNKNOWN_INVERT, UNKNOWN_OTHER_WIRE:
					return false
				case UNKNOWN_INVERT_OTHER_WIRE:
					dest.State = UNKNOWN_OTHER_WIRE
				}
				return true
			}

			switch a.State {
			case UNKNOWN, UNKNOWN_INVERT, UNKNOWN_OTHER_WIRE:
				dest.State = UNKNOWN_INVERT_OTHER_WIRE
				return false
			case UNKNOWN_INVERT_OTHER_WIRE:
				dest.State = UNKNOWN_OTHER_WIRE
				dest.Other = a.Other
			}
			dest.Other.AddRef(dest)
		}
		return true
	}
	return false
}

// Auxiliary function for ShortCut
func invertPassthrough(w *Wire, dest *Wire) bool {
	if w.State == ONE {
		dest.State = ZERO
		return true
	}
	if w.State == ZERO {
		dest.State = ONE
		return true
	}

	if dest == w {
		switch w.State {
		case UNKNOWN:
			dest.State = UNKNOWN_INVERT
		case UNKNOWN_INVERT:
			dest.State = UNKNOWN_INVERT
		case UNKNOWN_OTHER_WIRE:
			dest.State = UNKNOWN_INVERT_OTHER_WIRE
		case UNKNOWN_INVERT_OTHER_WIRE:
			dest.State = UNKNOWN_OTHER_WIRE
		}
		return true
	}

	switch w.State {
	case UNKNOWN_OTHER_WIRE:
		dest.Other = w.Other
		dest.State = UNKNOWN_INVERT_OTHER_WIRE
	case UNKNOWN_INVERT_OTHER_WIRE:
		dest.Other = w.Other
		dest.State = UNKNOWN_OTHER_WIRE
	default:
		dest.Other = w
		dest.State = UNKNOWN_INVERT_OTHER_WIRE
	}
	dest.Other.AddRef(dest)
	return true
}

// Auxiliary function for ShortCut
func passthrough(w *Wire, dest *Wire) bool {
	if w.State == ONE {
		dest.State = ONE
		return true
	}
	if w.State == ZERO {
		dest.State = ZERO
		return true
	}

	if dest == w {
		switch dest.State {
		case UNKNOWN:
			dest.State = UNKNOWN
		case UNKNOWN_INVERT:
			dest.State = UNKNOWN
		case UNKNOWN_OTHER_WIRE:
			dest.State = UNKNOWN_OTHER_WIRE
		case UNKNOWN_INVERT_OTHER_WIRE:
			dest.State = UNKNOWN_INVERT_OTHER_WIRE
		}
		return true
	}

	switch w.State {
	case UNKNOWN_OTHER_WIRE:
		dest.Other = w.Other
		dest.State = UNKNOWN_OTHER_WIRE
	case UNKNOWN_INVERT_OTHER_WIRE:
		dest.Other = w.Other
		dest.State = UNKNOWN_INVERT_OTHER_WIRE
	default:
		dest.Other = w
		dest.State = UNKNOWN_OTHER_WIRE
	}
	dest.Other.AddRef(dest)
	return true
}

// Auxiliary function for ShortCut when a or b is const and w2 is the other one
func oneIsConst(w2 *Wire, option0, option1 bool, dest *Wire) bool {
	if option0 && option1 {
		dest.State = ONE
		return true
	}

	if !option0 && !option1 {
		dest.State = ZERO
		return true
	}

	if !option0 && option1 {
		if dest == w2 {
			if dest.State == UNKNOWN || dest.State == UNKNOWN_INVERT {
				dest.State = UNKNOWN
			}
			return true
		}

		switch w2.State {
		case UNKNOWN_OTHER_WIRE:
			dest.Other = w2.Other
			dest.State = UNKNOWN_OTHER_WIRE
		case UNKNOWN_INVERT_OTHER_WIRE:
			dest.Other = w2.Other
			dest.State = UNKNOWN_INVERT_OTHER_WIRE
		default:
			dest.Other = w2
			dest.State = UNKNOWN_OTHER_WIRE
		}
		dest.Other.AddRef(dest)
		return true
	}

	if option0 && !option1 {
		if dest == w2 {
			switch dest.State {
			case UNKNOWN:
				dest.State = UNKNOWN_INVERT
			case UNKNOWN_INVERT:
				dest.State = UNKNOWN_INVERT
			case UNKNOWN_OTHER_WIRE:
				dest.State = UNKNOWN_INVERT_OTHER_WIRE
			case UNKNOWN_INVERT_OTHER_WIRE:
				dest.State = UNKNOWN_OTHER_WIRE
			}
			return true
		}

		switch w2.State {
		case UNKNOWN_OTHER_WIRE:
			dest.Other = w2.Other
			dest.State = UNKNOWN_INVERT_OTHER_WIRE
		case UNKNOWN_INVERT_OTHER_WIRE:
			dest.Other = w2.Other
			dest.State = UNKNOWN_OTHER_WIRE
		default:
			dest.Other = w2
			dest.State = UNKNOWN_INVERT_OTHER_WIRE
		}
		dest.Other.AddRef(dest)
		return true
	}

	return false
}

// Auxiliary function for ShortCutNoInvertOutput
func invertPassthroughNoInvertOutput(w *Wire, dest *Wire) bool {
	if w.State == ONE {
		dest.State = ZERO
		return true
	}
	if w.State == ZERO {
		dest.State = ONE
		return true
	}

	if dest == w {
		switch w.State {
		case UNKNOWN:
			return false
		case UNKNOWN_INVERT:
			return false
		case UNKNOWN_OTHER_WIRE:
			return false
		case UNKNOWN_INVERT_OTHER_WIRE:
			dest.State = UNKNOWN_OTHER_WIRE
		}
		return true
	}

	if w.State == UNKNOWN_OTHER_WIRE {
		return false
	}
	if w.State == UNKNOWN_INVERT_OTHER_WIRE {
		dest.Other = w.Other
		dest.State = UNKNOWN_OTHER_WIRE
	}
	dest.Other.AddRef(dest)
	return true
}

// Auxiliary function for ShortCutNoInvertOutput
func passthroughNoInvertOutput(w *Wire, dest *Wire) bool {
	if w.State == ONE {
		dest.State = ONE
		return true
	}
	if w.State == ZERO {
		dest.State = ZERO
		return true
	}

	if dest == w {
		switch dest.State {
		case UNKNOWN:
			dest.State = UNKNOWN
		case UNKNOWN_INVERT:
			dest.State = UNKNOWN
		case UNKNOWN_OTHER_WIRE:
			dest.State = UNKNOWN_OTHER_WIRE
		case UNKNOWN_INVERT_OTHER_WIRE:
			return false
		}
		return true
	}

	switch w.State {
	case UNKNOWN_INVERT_OTHER_WIRE:
		return false
	case UNKNOWN_OTHER_WIRE:
		dest.Other = w.Other
		dest.State = UNKNOWN_OTHER_WIRE
	default:
		dest.Other = w
		dest.State = UNKNOWN_OTHER_WIRE
	}
	dest.Other.AddRef(dest)
	return true
}

// Auxiliary function for ShortCutNoInvertOutput when a or b is const and w2 is the other one
func oneIsConstNoInvertOutput(w2 *Wire, option0, option1 bool, dest *Wire) bool {
	if option0 && option1 {
		dest.State = ONE
		return true
	}

	if !option0 && !option1 {
		dest.State = ZERO
		return true
	}

	if !option0 && option1 {
		if dest == w2 {
			if w2.State == UNKNOWN || w2.State == UNKNOWN_INVERT {
				dest.State = UNKNOWN
			}
			return true
		}

		if w2.State == UNKNOWN_INVERT_OTHER_WIRE {
			return false
		}

		dest.Other = w2
		dest.State = UNKNOWN_OTHER_WIRE

		if w2.State == UNKNOWN_OTHER_WIRE {
			dest.Other = w2.Other
		}
		dest.Other.AddRef(dest)
		return true
	}

	if option0 && !option1 {
		if dest == w2 {
			switch w2.State {
			case UNKNOWN:
				return false
			case UNKNOWN_INVERT:
				return false
			case UNKNOWN_OTHER_WIRE:
				return false
			case UNKNOWN_INVERT_OTHER_WIRE:
				dest.State = UNKNOWN_OTHER_WIRE
			}
			return true
		}

		switch w2.State {
		case UNKNOWN_OTHER_WIRE:
			return false
		case UNKNOWN_INVERT_OTHER_WIRE:
			dest.Other = w2.Other
			dest.State = UNKNOWN_OTHER_WIRE
		default:
			dest.Other = w2
			dest.State = UNKNOWN_INVERT_OTHER_WIRE
		}
		dest.Other.AddRef(dest)
		return true
	}

	return false
}
