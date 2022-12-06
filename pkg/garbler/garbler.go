package garbler

import (
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
	"math/rand"
	"strings"
	"time"
)

var debug bool = false

var N uint8                 // global security parameter
var offsetR circ.GarbledKey // the global key offset

var gateIndex uint32 // globalIndex is the number of the command that we are garbling
var outIndex uint32  // outIndex gives the index of the next output

func SetParams(deb bool) {
	debug = deb
}

func GarbleCompiledCircuit(fileName string, debug bool, n uint8) {
	if !strings.HasSuffix(fileName, ".re") {
		fmt.Println("Warning: input file has no re extension.")
	}
	Cin := circ.RetrieveCircuit(fileName)
	tStart := time.Now()
	tableSet, enc, dec := Garble(Cin, 8)

	if debug {
		tableSet.Print("")
		enc.Print("")
		dec.Print("")
	}

	// We output the circuit to a .ts file
	outputFileName := strings.Replace(fileName, "re", "ts", 1)
	tableSet.SaveToFile(outputFileName)

	tEnd := time.Now()
	diff := tEnd.Sub(tStart)
	fmt.Println("Garbling achieved in ", diff)
}

// Garble is the main exported function of the package.
// It garbles a given circuit, producing the three usual outputs:
// - the garbled circuit itself,
// - an encoding function and
// - a decoding function.
func Garble(Cin circ.Circuit, n uint8) (circ.TableSet, circ.EncodingSet, circ.DecodingSet) {
	if debug {
		fmt.Println("\n\nEntering Garble")
	}
	// We initialize the seed for randomness
	rand.Seed(time.Now().UTC().UnixNano())
	gateIndex = 0
	outIndex = 0

	// We create the table set from the plain circuit, completed and returned at the end of the garbling
	var TS circ.TableSet = make([]circ.GarbledTable, Cin.NonXORgates)

	// We initialize the values useful for the garbling
	N = n
	offsetR = circ.RandomGarbledKey(n)

	// wireSet is used to know what is the base value of every wire actually used
	// at a certain time of the execution and thus computes hashes of gates efficiently.
	wireSet := make([]circ.GarbledValue, Cin.TotalWires)
	wireSet[0] = circ.GarbledValue{false, circ.NullKey(n)}

	// We create the sets of encoding and decoding keys
	enc := circ.NewEncodingSet(offsetR, Cin.Parties)
	dec := circ.NewDecodingSet(Cin.Parties)

	var com circ.Command
	chcom := make(chan circ.Command, 5)
	go Cin.Visit(chcom, Cin.Funcs)

	for k := uint32(0); k < Cin.XORgates+Cin.NonXORgates; k++ {
		com = <-chcom
		if debug {
			com.Print("")
		}

		switch com.Kind {

		case circ.EMPTY_COMMAND:
			fmt.Println("\tError: empty command found")

		case circ.INPUT:
			// Creation of a key for the given wire
			wireSet[com.To] = circ.RandomGarbledValue(N)
			enc.User[com.X] = append(enc.User[com.X], wireSet[com.To])

		case circ.MASS_INPUT:
			for j := typ.Num(0); j < com.Y; j++ {
				wireSet[com.To+j] = circ.RandomGarbledValue(N)
				enc.User[com.X] = append(enc.User[com.X], wireSet[com.To+j])
			}

		case circ.COPY: // We keep the same key when we copy
			wireSet[com.To] = wireSet[com.X].Copy() // Warning: check if there is a proper copy done here

		case circ.MASS_COPY:
			for j := typ.Num(0); j < com.Y; j++ {
				wireSet[com.To+j] = wireSet[com.X+j].Copy()
			}

		case circ.REPLICATE:
			for j := typ.Num(0); j < com.Y; j++ {
				wireSet[com.To+j] = wireSet[com.X].Copy()
			}

		case circ.OUTPUT:
			dec.User[com.To] = append(dec.User[com.To], outKey(wireSet[com.X]))
			outIndex += 1

		case circ.MASS_OUTPUT:
			for j := typ.Num(0); j < com.Y; j++ {
				dec.User[com.To] = append(dec.User[com.To], outKey(wireSet[com.X+j]))
				outIndex += 1
			}

		default:
			if com.IsGate() {
				if com.Kind == circ.GATE_6 {
					wireSet[com.To] = wireSet[com.X].XOR(wireSet[com.Y])
				} else {
					TS[gateIndex], wireSet[com.To] = tableFromWires(wireSet[com.X], wireSet[com.Y], com.Gate())
					gateIndex += 1
				}
			} else {
				fmt.Println("Error in garbleList: found unknown kind.")
			}
		}
	}

	return TS, enc, dec
}

// The Get method enables us to access to the value of a wire that we want
func getVal(gv circ.GarbledValue, a bool) circ.GarbledValue {
	if a {
		return circ.NewGarbledValue(!gv.P, gv.Key.XOR(offsetR))
	}
	return gv
}

// GetValue Returns the key corresponding to the permutation bit p
func getKey(gv circ.GarbledValue, b bool) circ.GarbledKey {
	if b {
		return gv.Key.XOR(offsetR)
	}
	return gv.Key
}

// outKey is used in case of an output command.
// The argument provided is a certain garbled value, which corresponds to the wire we want
// to output. Then outKey will compute the two boolean values of the decoding key which the
// receiver will need to decrypt the result.
func outKey(gv circ.GarbledValue) circ.DecodingKey {
	e0 := circ.HashOut(gv.Key, outIndex)
	e1 := !circ.HashOut(gv.Key.XOR(offsetR), outIndex)
	if gv.P {
		return [2]bool{e1, e0}
	}
	return [2]bool{e0, e1}
}

// tableFromWires creates a table from the given wires and operator
func tableFromWires(wx, wy circ.GarbledValue, op uint8) (circ.GarbledTable, circ.GarbledValue) {
	// We create the garbled table used for this gate
	var table circ.GarbledTable

	// We find the zero-value of the resulting wire
	gvto := hashGate(getKey(wx, wx.P), getKey(wy, wy.P))
	if boolsToInt(wx.P, wy.P)&op != 0 {
		gvto.P = !gvto.P
		gvto.Key = gvto.Key.XOR(offsetR)
	}
	// We assign this zero-value to the receiving wire.
	// This is the core of the row reduction optimisation: because the key of the output wire is determined
	// bu input wires and not randomly, there is one case during the evaluation where we don't need to
	// use the table to decrypt the output value, but only to get the hash from the input wires.

	// We add the values in the tables.
	// Because we use the row reduction optimisation, we only need three values in the table,
	// but then there is a shift of one: the value of index 0 in our table would normally be of
	// index 1 in a table without row reduction, and so on.
	var px, py bool
	for i := 1; i < 4; i++ {
		px = (i/2 == 1) != wx.P
		py = (i%2 == 1) != wy.P
		table[i-1] = getVal(gvto, boolsToInt(px, py)&op != 0).XOR(hashGate(getKey(wx, px), getKey(wy, py)))
	}
	return table, gvto
}

// boolsToInt converts a pair of boolean variables into an integer between 0 and 3
func boolsToInt(a, b bool) uint8 {
	var r uint8 = 1
	if a {
		r *= 4
	}
	if b {
		r *= 2
	}
	return r
}

// hashGate produces the hash value used in case of a gate
func hashGate(k1, k2 circ.GarbledKey) circ.GarbledValue {
	return circ.HashGate(k1, k2, gateIndex, N)
}
