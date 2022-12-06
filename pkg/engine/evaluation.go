package engine

import (
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
	"os"
	"sync"
)

var wg sync.WaitGroup

// Evaluate is the function at the core of the evaluation of a circuit.
// It takes as argument the circuit to evaluate and two channels to receive inputs and send outputs.
// This implementation enables the function to be independent to a large extent of other parts of the code.
func Evaluate(C circ.Circuit, chtab chan circ.GarbledTable, chin []chan circ.GarbledValue, chout []chan circ.DecodingKey) {
	if n == 0 {
		fmt.Println("Execution package not initialized")
		os.Exit(64)
	}

	var wireSet []circ.GarbledValue = make([]circ.GarbledValue, C.TotalWires)
	wireSet[0] = circ.GarbledValue{false, circ.NullKey(n)}

	var gateIndex uint32 = 0
	var outIndex uint32 = 0

	var gt circ.GarbledTable
	var wa, wb circ.GarbledValue

	var com circ.Command
	chcom := make(chan circ.Command, 5)
	go C.Visit(chcom, C.Funcs)

	for k := uint32(0); k < C.XORgates+C.NonXORgates; k++ {
		com = <-chcom
		switch com.Kind {

		case circ.INPUT:
			wireSet[com.To] = <-chin[com.X]

		case circ.MASS_INPUT:
			for j := typ.Num(0); j < com.Y; j++ {
				wireSet[com.To+j] = <-chin[com.X]
			}

		case circ.COPY:
			wireSet[com.To] = wireSet[com.X]

		case circ.MASS_COPY:
			for j := typ.Num(0); j < com.Y; j++ {
				wireSet[com.To+j] = wireSet[com.X+j]
			}

		case circ.REPLICATE:
			for j := typ.Num(0); j < com.Y; j++ {
				wireSet[com.To+j] = wireSet[com.X]
			}

		case circ.OUTPUT:
			wa = wireSet[com.X]
			chout[com.To] <- circ.DecodingKey{wa.P, circ.HashOut(wa.Key, outIndex)}
			outIndex += 1

		case circ.MASS_OUTPUT:
			for j := typ.Num(0); j < com.Y; j++ {
				wa = wireSet[com.X+j]
				chout[com.To] <- circ.DecodingKey{wa.P, circ.HashOut(wa.Key, outIndex)}
				outIndex += 1
			}

		default:
			if com.IsGate() {
				if com.Kind == circ.GATE_6 {
					wireSet[com.To] = wireSet[com.X].XOR(wireSet[com.Y])
				} else {
					gt = <-chtab
					wa = wireSet[com.X]
					wb = wireSet[com.Y]
					wireSet[com.To] = circ.HashGate(wa.Key, wb.Key, gateIndex, n).XOR(gt.GetValue(wa.P, wb.P))
					gateIndex += 1
				}
			} else {
				fmt.Println("Error in garbleList: found unknown kind.")
			}
		}
	}
}

// TabSender sends progressively all table from a TableSet object to a given channel
func TabSender(TS circ.TableSet, chtab chan<- circ.GarbledTable) {
	defer wg.Done()
	for _, tab := range TS {
		chtab <- tab
	}
}

// InputSender sends to a channel the input values of one party
func InputSender(inp []circ.GarbledValue, chin chan<- circ.GarbledValue) {
	defer wg.Done()
	for _, v := range inp {
		chin <- v
	}
}

// OutputReceiver receives from a channel all output values to a certain party, decodes it and returns the clear value
func OutputReceiver(udec circ.UserDecoder, chout <-chan circ.DecodingKey, dest *circ.UserInOut) {
	defer wg.Done()
	outputs := make([]circ.DecodingKey, len(udec), len(udec))
	for i := 0; i < len(udec); i++ {
		outputs[i] = <-chout
	}
	*dest = udec.Decode(outputs)
}
