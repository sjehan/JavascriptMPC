package compiler

import (
	"fmt"
	"os"
	"strings"

	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
	vb "ixxoprivacy/pkg/variables"
	wr "ixxoprivacy/pkg/wires"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
)

var circuit circ.Circuit      // the circuit we work on
var writer FuncWriter         // the intermediary object that we use to push commands to the circuit
var pool wr.WirePool          // the pool all of wires which are not user-defined to fixed variables
var context vb.ProgramContext // contains information about the variables, the functions and their types

var W_1 *wr.Wire
var W_0 *wr.Wire

var nextBaseWire typ.Num = 0 // variable to keep count of the wire number to use next

/* Some parameters which can be defined using SetParamsCG */
var printAST bool = false
var printCont bool = false
var printIOTypeWires bool = true
var debug bool = false

// This exported function enables caller to tweak parameters of the compiler
// in order to print some of the content used or produced.
func SetParamsCG(bast, bcont, bdebug bool) {
	printAST = bast
	printCont = bcont
	debug = bdebug
}

// CircuitFromJS returns a boolean circuit from a JavaScript file whose path
// is given in argument
func CircuitFromJS(path string) (circ.Circuit, error) {
	src, err := os.Open(path)
	if err != nil {
		fmt.Println("Error : could not open file")
		fmt.Println(err)
		os.Exit(64)
	}
	program, err := parser.ParseFile(nil, "", src, 0)
	src.Close()
	if err != nil {
		fmt.Println("Error : could not parse file :\n")
		fmt.Println(err)
		os.Exit(64)
	}
	return CircuitFromAST(program)
}

// CircuitFromAST returns a boolean circuit from an abstract syntax tree
// whith the format used in the otto package
func CircuitFromAST(prog *ast.Program) (circ.Circuit, error) {
	if printAST {
		typ.PrintAST(prog, false)
	}
	if debug {
		fmt.Println("Starting to initialize")
	}

	circuit = circ.NewCircuit(findParameters(prog.DeclarationList))
	writer = StartFuncWriter(&circuit.Function)
	makeONEandZERO()

	context = vb.GenerateContext(prog, circuit.IntSize, W_0, W_1)
	if printCont {
		context.Print("")
	}
	fNames := make([]string, 0) // fNames contains the names of the functions using the same index as the one in circuit

	if debug {
		fmt.Println("Starting with variable set up")
	}

	// We initialize permanent wires for all variables
	for name, v := range context.FunctionContext {

		if fv, ok := v.(*vb.FunctionVariable); !ok {
			v.FillInWires(nil)
			if !strings.HasPrefix(v.GetName(), "$") {
				v.SetPerm()
				nextBaseWire = v.AssignPermWires(nextBaseWire)
				for i := typ.Num(0); i < v.Size(); i++ {
					v.GetWire(i).State = wr.UNKNOWN
				}
			}

			if v.IsInput() {
				circuit.Inputs[getParty(name)] = vb.CircVar(v)

			} else if v.IsOutput() {
				circuit.Outputs[getParty(name)] = vb.CircVar(v)
			}

		} else {
			fv.FunctionNumber = typ.Num(len(circuit.Funcs))
			circuit.Funcs = append(circuit.Funcs, circ.NewFunctionPt())
			fNames = append(fNames, fv.GetName())

			for _, v := range context.Funcs[fv.GetName()] {
				v.FillInWires(nil)
				if strings.HasPrefix(v.GetName(), "$") {
					v.SetConst()
				} else {
					v.SetPerm()
					nextBaseWire = v.AssignPermWires(nextBaseWire)
				}
			}

			// setting arguments wires to unknown
			for _, av := range fv.Argsv {
				for i := typ.Num(0); i < av.Size(); i++ {
					av.GetWire(i).State = wr.UNKNOWN
				}
			}
			// setting return wires to unknown if they exist
			if fv.Returnv != nil {
				for i := typ.Num(0); i < fv.Returnv.Size(); i++ {
					fv.Returnv.GetWire(i).State = wr.UNKNOWN
				}
			}
		}
	}

	// TODO: sort functions
	pool = wr.NewWirePool(nextBaseWire)

	/*
	 * We start writing gates from that point
	 */

	// We write input gates
	for party, v := range circuit.Inputs {
		if v.Type.Size() == 1 {
			writer.AddIn(v.Wirebase, typ.Num(party))
		} else {
			writer.AddMassIn(v.Wirebase, typ.Num(v.Type.Size()), typ.Num(party))
		}
	}

	if debug {
		fmt.Println("\nStarting with functions\n")
	}

	// We output the auxiliary functions
	for i, f := range circuit.Funcs {
		writer.ChangeFunction(f)
		name := fNames[i]
		fv := context.FunctionContext[name].(*vb.FunctionVariable)
		fc := context.Funcs[name]

		outFunctionLiteral(fv.FunctionNode, fc)

		nextBaseWire = pool.NextNumber
		pool = wr.NewWirePool(nextBaseWire)

		for _, v := range context.FunctionContext {
			if _, ok := v.(*vb.FunctionVariable); !ok && !strings.HasPrefix(v.GetName(), "$") {
				for i := typ.Num(0); i < v.Size(); i++ {
					v.GetWire(i).State = wr.UNKNOWN
				}
			}
		}
	}

	// Output of the main body
	if debug {
		fmt.Println("\nStarting with main\n")
	}
	writer.ChangeFunction(&circuit.Function)

	mainNode := &ast.BlockStatement{List: prog.Body}
	outStatementNode(mainNode, context.FunctionContext)

	// We write output gates
	for party, v := range circuit.Outputs {
		if v != nil {
			if v.Type.Size() == 1 {
				writer.AddOut(v.Wirebase, typ.Num(party))
			} else {
				writer.AddMassOut(v.Wirebase, typ.Num(v.Type.Size()), typ.Num(party))
			}
		}
	}

	writer.AddPrev(nullComm)
	nextBaseWire = pool.NextNumber
	circuit.TotalWires = nextBaseWire

	return circuit, nil
}
