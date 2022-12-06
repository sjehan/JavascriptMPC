package runner

import (
	"flag"
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	garbler "ixxoprivacy/pkg/garbler"
	ip "ixxoprivacy/pkg/interpreter"
	"os"
	"strings"
	"time"
)

/*
 * This package is meant to test the validity of compiled circuit.
 * It provides a way to execute the circuit, the same way we would execute a
 * programm resulting of a usual compiler. It relies mostly on the package
 * interpreter. The name is a reference to the executable Battleship which
 * accompanied the compiler Frigate which was itself a major inspiration
 * source for RockEngine.
 *
 * It takes as arguments a circuit with a .re extension and a JavaScript file
 * containing the data to use as inputs.
 */

/*
 * Flags:
 * -see_out - see output wires
 * -see_in - see input wires
 * -circ  - outputs the circuit file into an inlined plain text format
 *
 * The other arguments provided are the path to the source file and to the input files.
 */

var seeOutput bool = false
var seeInput bool = false
var printCircuit bool = false

var printTables bool = false
var debug bool = false
var printTime bool = true

var intSize uint16

func RunCircuit(circuitfilename string, inputFiles []string) {
	/*
	 * First we process args
	 */

	flag.BoolVar(&seeOutput, "see_out", false, "see output wires")
	flag.BoolVar(&seeInput, "see_in", false, "see input wires")
	flag.BoolVar(&printCircuit, "circ", false, "outputs the circuit file into an inlined plain text format")
	flag.Parse()

	// Timer to measure the running time
	tStart := time.Now()

	if !strings.HasSuffix(circuitfilename, ".re") {
		fmt.Println("Warning: input file has no re extension.")
	}

	var st string
	for i := 1; i < len(inputFiles); i++ {
		st = inputFiles[i]
		if !strings.HasSuffix(st, ".json") {
			fmt.Println("Warning: entry file", st, "has no json extension.")
		}
	}

	// Decoding of the circuit
	circuit := circ.RetrieveCircuit(circuitfilename)
	if printCircuit {
		circuit.Print("")
	}
	if circuit.Parties != uint8(len(inputFiles)) {
		fmt.Println("Error: number of argument doesn't match number of parties of the circuit.")
		os.Exit(64)
	}

	// We find the input given in the entry file
	inputs := ip.GetAllInputs(circuit.Inputs, inputFiles)
	if seeInput {
		for party, inp := range inputs {
			fmt.Println("Input of party ", party)
			inp.Print("\t")
			fmt.Println()
		}
	}

	// We run the interpreter with the given inputs
	// It returns a map of bytes buffers. Each buffer is for a certain party.
	output := ip.Interprete(circuit, inputs)

	if seeOutput {
		for party, outp := range output {
			if len(*outp) != 0 {
				fmt.Println("Wire output to party", party)
				outp.Print("\t")
			}
		}
	}

	for party, out := range output {
		fmt.Println("Interpreted output to party", party)
		ip.PrintResult(out, circuit.Outputs[party].Type)
	}

	tEnd := time.Now()
	diff := tEnd.Sub(tStart)
	fmt.Println("Interpretation achieved in ", diff)
}

func GarbleCircuit(circuitFileName string) {
	/*
	 * First we process args
	 */
	tStart := time.Now()

	flag.BoolVar(&printTables, "tab", false, "outputs the circuit file into an inlined plain text format")
	flag.BoolVar(&debug, "debug", false, "prints extra information, to be used for debugging purposes")

	flag.Parse()

	// Decoding of the circuit
	circuit := circ.RetrieveCircuit(circuitFileName)
	if debug {
		circuit.Print("")
		garbler.SetParams(true)
	}

	// We run the interpreter with the given inputs
	// It returns a map of bytes buffers. Each buffer is for a certain party.
	tableSet, enc, dec := garbler.Garble(circuit, 8)

	if printTables {
		tableSet.Print("")
		enc.Print("")
		dec.Print("")
	}

	// We output the circuit to a .ts file
	outputFileName := strings.Replace(circuitFileName, "re", "ts", 1)
	tableSet.SaveToFile(outputFileName)

	tEnd := time.Now()
	diff := tEnd.Sub(tStart)
	if printTime {
		fmt.Println("Garbling achieved in ", diff)
	}
}
