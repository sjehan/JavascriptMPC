package compiler

import (
	"fmt"

	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
)

// A FuncWriter variable is an object which progressively writes new gates into a function
// of the circuit. Its main interest over the solution which would be to simply write gates
// as they come is that it will simplify consecutive gates when it can, thus reducing the
// size of the circuit.
type FuncWriter struct {
	f    *circ.Function
	prev circ.Command
}

// The variable nullComm is used to initialize FuncWriter's
var nullComm circ.Command = circ.Command{circ.EMPTY_COMMAND, 0, 0, 0}

// StartFuncWriter creates a new FuncWriter variable corresponding to a given circuit function
func StartFuncWriter(f *circ.Function) FuncWriter {
	if debug {
		fmt.Println("Starting function writer")
	}
	return FuncWriter{f, nullComm}
}

// AddPrev pushes the last added command to the circuit function and replace the prev field with
// a new command
func (fw *FuncWriter) AddPrev(newComm circ.Command) {
	if fw.prev.Kind != circ.EMPTY_COMMAND {
		if fw.prev.Kind == circ.FUNCTION_CALL {
			fpushed := circuit.Funcs[fw.prev.X]
			fw.f.PushFunctionCall(fw.prev, fpushed.XORgates, fpushed.NonXORgates)
		} else {
			fw.f.PushNonFunctionCall(fw.prev)
		}
	}
	fw.prev = newComm
}

// ChangeFunction is used to keep the same FuncWriter object but change the underlying function
func (fw *FuncWriter) ChangeFunction(newF *circ.Function) {
	fw.AddPrev(nullComm)
	fw.f = newF
	if debug {
		fmt.Println("Changing function")
	}
}

// GetFunction returns the circuit function of the FuncWriter object
func (fw FuncWriter) GetFunction() *circ.Function {
	return fw.f
}

// AddGate adds a gate to the circuit whose input wires are given by x and y,
// the destination wire is d and the operator is table.
func (fw *FuncWriter) AddGate(table uint8, d, x, y typ.Num) {
	fw.AddPrev(circ.Command{circ.CommandType(circ.GATE_0 + table), x, y, d})
	if debug {
		fmt.Printf("Gate: %d(%d, %d) -> %d\n", table, x, y, d)
	}
}

// AddCopy adds a copy command to the circuit.
// The wire "to" takes the value of wire "from".
func (fw *FuncWriter) AddCopy(to, from typ.Num) {
	if fw.prev.Kind == circ.COPY && fw.prev.X == from-1 && fw.prev.To == to-1 {
		fw.prev.Kind = circ.MASS_COPY
		fw.prev.Y = 2
	} else if fw.prev.Kind == circ.MASS_COPY && fw.prev.X+fw.prev.Y == from && fw.prev.To+fw.prev.Y == to {
		fw.prev.Y++
	} else if fw.prev.Kind == circ.REPLICATE && fw.prev.X == from && fw.prev.To+fw.prev.Y == to {
		fw.prev.Y++
	} else {
		fw.AddPrev(circ.Command{circ.COPY, from, 0, to})
	}
	if debug {
		fmt.Printf("Copy: %d -> %d\n", from, to)
	}
}

// AddMassCopy adds a mass copy command to the circuit, which is a compact way to write
// a copy of set of consecutive wires onto another set of consecutive wires.
func (fw *FuncWriter) AddMassCopy(to, from, len typ.Num) {
	if fw.prev.Kind == circ.COPY && fw.prev.X == from-1 && fw.prev.To == to-1 {
		fw.prev.Kind = circ.MASS_COPY
		fw.prev.Y = len + 1
	} else if fw.prev.Kind == circ.MASS_COPY && fw.prev.X+fw.prev.Y == from && fw.prev.To+fw.prev.Y == to {
		fw.prev.Y += len
	} else {
		fw.AddPrev(circ.Command{circ.MASS_COPY, from, len, to})
	}
	if debug {
		fmt.Printf("Mass Copy: (%d, %d) -> (%d, %d)\n", from, from+len, to, to+len)
	}
}

// AddFunctionCall adds to the circuit a function call command.
func (fw *FuncWriter) AddFunctionCall(fid typ.Num) {
	fw.AddPrev(circ.Command{circ.FUNCTION_CALL, fid, 0, 0})
	if debug {
		fmt.Printf("Call: %d\n", fid)
	}
}

// AddProcCall adds to the circuit a procedure command, the set of commands
// of this procedure is contained in the function referenced and argument len is the number
// of iterations of this procedure.
func (fw *FuncWriter) AddProcCall(fid, itr typ.Num) {
	fw.AddPrev(circ.Command{circ.FUNCTION_CALL, fid, itr, 0})
	if debug {
		fmt.Printf("Procedure × %d\n", itr)
	}
}

// AddIn adds to the circuit an input command.
// The next input of participant "party" is assigned to wire "wire".
func (fw *FuncWriter) AddIn(wire, party typ.Num) {
	if fw.prev.Kind == circ.INPUT && fw.prev.X == party && fw.prev.To == wire-1 {
		fw.prev.Kind = circ.MASS_INPUT
		fw.prev.Y = 2
	} else if fw.prev.Kind == circ.MASS_INPUT && fw.prev.X == party && fw.prev.To+fw.prev.Y == wire {
		fw.prev.Y++
	} else {
		fw.AddPrev(circ.Command{circ.INPUT, party, 0, wire})
	}
	if debug {
		fmt.Printf("Input: %d from %d\n", wire, party)
	}
}

// AddMassIn adds to the circuit a mass input commands, which is a compact way to
// assign inputs from the same party to a set of consecutive wires.
func (fw *FuncWriter) AddMassIn(wire, len, party typ.Num) {
	if fw.prev.Kind == circ.INPUT && fw.prev.X == party && fw.prev.To == wire-1 {
		fw.prev.Kind = circ.MASS_INPUT
		fw.prev.Y = len + 1
	} else if fw.prev.Kind == circ.MASS_INPUT && fw.prev.X == party && fw.prev.To+fw.prev.Y == wire {
		fw.prev.Y += len
	} else {
		fw.AddPrev(circ.Command{circ.MASS_INPUT, party, len, wire})
	}
	if debug {
		fmt.Printf("Mass Input: (%d, %d) from %d\n", wire, wire+len, party)
	}
}

// AddOut adds to the circuit an output command.
// The value of wire "wire" is directed to participant "party".
func (fw *FuncWriter) AddOut(wire, party typ.Num) {
	if fw.prev.Kind == circ.OUTPUT && fw.prev.X == wire-1 && fw.prev.To == party {
		fw.prev.Kind = circ.MASS_OUTPUT
		fw.prev.Y = 2
	} else if fw.prev.Kind == circ.MASS_OUTPUT && fw.prev.X+fw.prev.Y == wire && fw.prev.To == party {
		fw.prev.Y++
	} else {
		fw.AddPrev(circ.Command{circ.OUTPUT, wire, 0, party})
	}
	if debug {
		fmt.Printf("Input: %d to %d\n", wire, party)
	}
}

// AddMassOut adds to the circuit a mass output commands, which is a compact way to
// assign outputs from a set of consecutive wires to a single party.
func (fw *FuncWriter) AddMassOut(wire, len, party typ.Num) {
	if fw.prev.Kind == circ.OUTPUT && fw.prev.X == wire-1 && fw.prev.To == party {
		fw.prev.Kind = circ.MASS_OUTPUT
		fw.prev.Y = len + 1
	} else if fw.prev.Kind == circ.MASS_OUTPUT && fw.prev.X+fw.prev.Y == wire && fw.prev.To == party {
		fw.prev.Y += len
	} else {
		fw.AddPrev(circ.Command{circ.MASS_OUTPUT, wire, len, party})
	}
	if debug {
		fmt.Printf("Mass Output: (%d, %d) to %d\n", wire, wire+len, party)
	}
}

// AddReplicate adds a replicate commmand to the circuit. It is a compact way
// to represent a series of copy commands from a single wire to a set of consecutive wires.
func (fw *FuncWriter) AddReplicate(to, from, len typ.Num) {
	if fw.prev.Kind == circ.COPY && fw.prev.X == from && fw.prev.To == to-1 {
		fw.prev.Kind = circ.REPLICATE
		fw.prev.Y = len + 1
	} else if fw.prev.Kind == circ.REPLICATE && fw.prev.X == from && fw.prev.To+fw.prev.Y == to {
		fw.prev.Y += len
	} else {
		fw.AddPrev(circ.Command{circ.REPLICATE, from, len, to})
	}
	if debug {
		fmt.Printf("Replicate: %d -> (%d, %d)\n", from, to, to+len)
	}
}
