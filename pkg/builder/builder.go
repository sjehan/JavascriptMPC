package builder

import (
	"flag"
	"fmt"
	compiler "ixxoprivacy/pkg/compiler"
	"strings"
	"time"
)

/*
 * GPE.build is an important executable since it is the one which allows to compile a circuit
 * from a JavaScript file. It relies directly on the package compiler.
 *
 * The executable takes only one compulsory argument which is the path to the JavaScript file.
 * The optional arguments (or flags) are described below.
 */

/*
 * Flags:
 * -ast - print the ast, for debugging purposes
 * -cont - print the TypeMap and VariableContext for debugging purposes
 * -circ  - outputs the circuit file into an inlined plain text format
 * -no_time - print compile time
 * -debug - see output (for debugging purposes)
 * -nowarn  - do not show warnings
 *
 * The other argument provided is the path to the source file.
 */

var printAST bool = false
var printCont bool = false
var printCicrcuit bool = false
var printCompileTime bool = true
var debug bool = false
var gatelistfilename string

func BuildCircuit(fileName string) {
	/*
	 * First we process args
	 */
	flag.BoolVar(&printAST, "ast", false, "print AST before compilation")
	flag.BoolVar(&printCont, "cont", false, "print context during compilation")
	flag.BoolVar(&printCicrcuit, "circ", false, "prints the resulting circuit into a txt file")
	flag.BoolVar(&debug, "debug", false, "prints details of the compilation for debugging purposes")

	noTimer := flag.Bool("no_time", false, "print compile time")

	flag.Parse()

	compiler.SetParamsCG(printAST, printCont, debug)

	if *noTimer {
		printCompileTime = false
	}

	tStart := time.Now()

	// Compilation of the program into a boolean circuit
	circuit, err := compiler.CircuitFromJS(fileName)
	if err != nil {
		fmt.Println("Compilation error:")
		fmt.Println(err)
	}

	// We output the circuit to a .re file
	outputFileName := strings.Replace(fileName, "js", "re", 1)
	circuit.SaveToFile(outputFileName)
	fmt.Println("Compiled circuit saved to", outputFileName)
	tEnd := time.Now()
	diff := tEnd.Sub(tStart)
	if printCompileTime {
		fmt.Println("Compilation achieved in ", diff)
		fmt.Println("TotalWires", circuit.TotalWires)
		fmt.Println("XORgates", circuit.XORgates)
		fmt.Println("NonXORgates", circuit.NonXORgates)
	}

	if printCicrcuit {
		circuit.Print("")
	}
}
