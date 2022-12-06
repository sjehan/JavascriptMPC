package circuit

import (
	"encoding/gob"
	"fmt"
	typ "ixxoprivacy/pkg/types"
	"os"
)

// To each different kind of commands used in circuits correspond to a
// certain value of type CommandType.
type CommandType byte

const (
	EMPTY_COMMAND = iota
	COPY
	FUNCTION_CALL
	INPUT
	OUTPUT
	MASS_COPY
	MASS_INPUT
	MASS_OUTPUT
	REPLICATE

	GATE_0 // ZERO gate
	GATE_1 // NOR
	GATE_2
	GATE_3 // invert passthrough a
	GATE_4
	GATE_5  // invert passthrough b
	GATE_6  // XOR
	GATE_7  // NAND
	GATE_8  // AND
	GATE_9  // reverse XOR
	GATE_10 // passthrough b
	GATE_11
	GATE_12 // passthrough a
	GATE_13
	GATE_14 // OR
	GATE_15 // ONE gate
)

// The basic structure of a command, which is a unit of computational
// work to be executed during the evaluation.
type Command struct {
	Kind CommandType
	X    typ.Num
	Y    typ.Num
	To   typ.Num
}

// The Function type is used to describe a recursive composant of the
// code, which often corresponds to a function as we know it which was
// defined in the code by the user.
type Function struct {
	XORgates    uint32
	NonXORgates uint32
	Commands    []Command
}

// The Var type is used to store informations about input and output
// variables in the circuit for it to be easy to understand by an
// external reader. It has no practical utility during the garbling
// and evaluation.
type Var struct {
	*typ.Type
	Wirebase typ.Num
}

// This structure describes a whole circuit.
// To read it one has to start by reading the commands contained
// in the attribute Main.
type Circuit struct {
	Function
	Parties    uint8
	IntSize    typ.Num
	TotalWires typ.Num
	Inputs     []*Var
	Outputs    []*Var
	Funcs      []*Function
}

// UserInOut objects are used to deal with clear inputs and outputs
type UserInOut []bool

/*         Methods and functions on commands         */
/*****************************************************/

// IsGate returns true if the command given is a gate and false otherwise
func (c *Command) IsGate() bool {
	return c.Kind >= GATE_0 && c.Kind <= GATE_15
}

// Gate returns the number-operator corresponding to the given command
// when it is a gate
func (c *Command) Gate() byte {
	if c.Kind < GATE_0 || c.Kind > GATE_15 {
		fmt.Println("Error in GateNum: given command is not a gate")
		os.Exit(64)
	}
	return byte(c.Kind - GATE_0)
}

/*     Methods and functions on Functions      */
/***********************************************/

// NewFunction returns a new Function variable
func NewFunction() Function {
	return Function{
		Commands: make([]Command, 0),
	}
}

// NewFunctionPt returns a pointer to a new Function variable
func NewFunctionPt() *Function {
	f := new(Function)
	f.Commands = make([]Command, 0)
	return f
}

// PushNonFunctionCall adds a command which is not of type FUNCTION_CALL to the given function
// The methods takes care of increasing the count of XOR or non-XOR gates
func (f *Function) PushNonFunctionCall(com Command) {
	if com.Kind == FUNCTION_CALL || com.Kind == EMPTY_COMMAND || com.Kind > GATE_15 {
		fmt.Println("Error in PushNonFunctionCall: invalid command kind, received", com.Kind)
		os.Exit(64)
	} else if com.IsGate() && com.Kind != GATE_6 {
		f.NonXORgates++
	} else {
		f.XORgates++
	}
	f.Commands = append(f.Commands, com)
}

// PushFunctionCall adds a command of type FUNCTION_CALL to the given function.
// It uses the count of XOR and non-XOR gates provided to update the count of the function manipulated.
func (f *Function) PushFunctionCall(com Command, xor, nxor uint32) {
	if com.Kind != FUNCTION_CALL {
		fmt.Println("Error in PushFunctionCall: invalid command kind, received", com.Kind)
		os.Exit(64)
	} else if com.Y > 0 {
		f.XORgates += uint32(com.Y) * xor
		f.NonXORgates += uint32(com.Y) * nxor
	} else {
		f.XORgates += xor
		f.NonXORgates += nxor
	}
	f.Commands = append(f.Commands, com)
}

// Visit goes through all commands starting from a given function and transmits those
// commands through a channel.
func (f Function) Visit(chcom chan<- Command, funcs []*Function) {
	for _, com := range f.Commands {
		if com.Kind == FUNCTION_CALL {
			funcs[com.X].Visit(chcom, funcs)
			for i := uint32(1); i < uint32(com.Y); i++ {
				funcs[com.X].Visit(chcom, funcs)
			}
		} else {
			chcom <- com
		}
	}
}

/*         Methods and functions on Circuits        */
/****************************************************/

// NewCircuit returns a new circuit variable
func NewCircuit(intSize typ.Num, parties uint8) Circuit {
	return Circuit{
		Function: NewFunction(),
		Parties:  parties,
		IntSize:  intSize,
		Funcs:    make([]*Function, 0),
		Inputs:   make([]*Var, parties),
		Outputs:  make([]*Var, parties),
	}
}

// SaveToFile saves a circuit into a file whose path is given in
// argument and using standard gobs encoding
func (C *Circuit) SaveToFile(path string) {
	for i, in := range C.Inputs {
		if in == nil {
			C.Inputs[i] = &Var{typ.VoidType, 0}
		}
	}
	for i, out := range C.Outputs {
		if out == nil {
			C.Outputs[i] = &Var{typ.VoidType, 0}
		}
	}
	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error in SaveToFile: file creation failed")
		fmt.Println(err)
		os.Exit(64)
	}
	encoder := gob.NewEncoder(outputFile)
	err = encoder.Encode(C)
	if err != nil {
		fmt.Println("Error in SaveToFile: encoding failed")
		fmt.Println(err)
		os.Exit(64)
	}
	outputFile.Close()
}

// RetrieveFromFile is used to get the circuit from a file generated
// with method SaveToFile
func RetrieveCircuit(path string) Circuit {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error: could not open input file")
		fmt.Println(err)
		os.Exit(64)
	}
	decoder := gob.NewDecoder(file)
	var C Circuit
	err = decoder.Decode(&C)
	if err != nil {
		fmt.Println("Error: could not decode circuit.")
		fmt.Println(err)
		os.Exit(64)
	}
	file.Close()
	return C
}

/*              Methods and functions on UserInOut               */
/*****************************************************************/

// NewUIO creates a new UserInOut object
func NewUIO() *UserInOut {
	var uio UserInOut = make([]bool, 0)
	return &uio
}

// Add is a method to add a boolean value to a UserInOut object
func (uio *UserInOut) Add(b bool) {
	*uio = append(*uio, b)
}

// Pop returns the first value in a UserInOut object and removes it
func (uio *UserInOut) Pop() (b bool) {
	b, *uio = (*uio)[0], (*uio)[1:]
	return b
}

// Copy creates a new UserInOut object identical to the one given
func (uio *UserInOut) Copy() (newuio *UserInOut) {
	newuio = NewUIO()
	for _, b := range *uio {
		newuio.Add(b)
	}
	return newuio
}

func (uio UserInOut) SubUIO(start, len typ.Num) (subuio *UserInOut) {
	subuio = NewUIO()
	for i := start; i < start+len; i++ {
		subuio.Add(uio[i])
	}
	return subuio
}

// Equals tests the equality of two UserInOut objects
// It returns true iff the two have the same length and all of there components are equal
func (uio1 UserInOut) Equals(uio2 *UserInOut) bool {
	if len(uio1) != len(*uio2) {
		return false
	}
	for i, b := range *uio2 {
		if uio1[i] != b {
			return false
		}
	}
	return true
}
