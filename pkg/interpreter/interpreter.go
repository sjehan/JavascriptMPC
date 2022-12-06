package interpreter

import (
	"fmt"

	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
)

var seeDetails bool = false

// Interprete is the entry point of the interpreter package.
// It runs all commands contained in a circuit in order.
func Interprete(C circ.Circuit, origInputs []*circ.UserInOut) []*circ.UserInOut {
	outputs := make([]*circ.UserInOut, len(origInputs)) // The output buffers, one for each receiving party
	wires := make(map[typ.Num]bool)                     // Set of booleans representing the wires used during the execution

	inputs := make([]*circ.UserInOut, len(origInputs))
	for i, inp := range origInputs {
		inputs[i] = inp.Copy()
	}

	var com circ.Command
	chcom := make(chan circ.Command, 5)
	go C.Visit(chcom, C.Funcs)

	for k := uint32(0); k < C.XORgates+C.NonXORgates; k++ {
		com = <-chcom
		if seeDetails {
			com.Print("")
		}

		switch com.Kind {
		case circ.EMPTY_COMMAND:
			fmt.Println("Error: empty command found")

		case circ.COPY:
			wires[com.To] = wires[com.X]

		case circ.MASS_COPY:
			for i := typ.Num(0); i < com.Y; i++ {
				wires[com.To+i] = wires[com.X+i]
			}

		case circ.INPUT:
			wires[com.To] = inputs[com.X].Pop()

		case circ.MASS_INPUT:
			for i := typ.Num(0); i < com.Y; i++ {
				wires[com.To+i] = inputs[com.X].Pop()
			}

		case circ.OUTPUT:
			if outputs[com.To] == nil {
				outputs[com.To] = circ.NewUIO()
			}
			outputs[com.To].Add(wires[com.X])

		case circ.MASS_OUTPUT:
			if outputs[com.To] == nil {
				outputs[com.To] = circ.NewUIO()
			}
			for i := typ.Num(0); i < com.Y; i++ {
				outputs[com.To].Add(wires[com.X+i])
			}

		case circ.REPLICATE:
			for i := typ.Num(0); i < com.Y; i++ {
				wires[com.To+i] = wires[com.X]
			}

		default:
			if com.IsGate() {
				wires[com.To] = 1<<(2*conv(wires[com.X])+conv(wires[com.Y]))&com.Gate() != 0
			} else {
				fmt.Println("Error: unrecognized command type")
			}
		}
	}

	return outputs
}

// conv is an auxiliary function to convert from bool to uint
func conv(b bool) uint {
	if b {
		return 1
	}
	return 0
}
