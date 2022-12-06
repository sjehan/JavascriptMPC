
![](RockEngine.png?raw=true)

Secured N-parties (N>=2) multi-party computation - using garbled circuits - in golang.
The input algorithms are in Javascript, making it easy to use in the browser, or in Node.js.

MPC can allow to solve problems such as:
1. Private set intersection (PSI) of two (or more) parties lists
2. Sign a message with a private key, without revealing the private key (shared on several parties)
3. Compute a function on a private input, without revealing the input (shared on several parties)
4. Match some orders from remote orderbooks, without revealing the orders (shared on several parties)
5. Do survey results aggregation, without revealing the answers (shared on several parties)
6. Vote on a secret ballot, without revealing the vote (shared on several parties)

This package requires you to have an understanding of the concepts of MPC, and garbled circuits.
It is not a turn-key solution, but rather an infrastructure for you to build your own MPC application.

To get technical support and the Enterprise version, contact us at: contact@ixxo.io

Enterprise version includes:
- A web-based UI for MPC
- A web-based UI for MPC with a trusted party
- concurrency support for MPC: sort gates so that when two cages can be computed at the same time, one follows the other in topoligical order
- better number encoding
- fast multiplication techniques
- circuit optimization techniques taken from CMBC-GC
- variable reuse
- Integration of FlexOR for performance improvement

We keep 80% of our code base open source on MPC, and 20% closed source. Help us continue improving MPC by contributing to the open source part of the code base or by buying the Enterprise version so that we keep pouring more resources into open source MPC.

## **Setup**
### Installation
Before using RockEngine, you need to install the following dependencies:
- Go (https://golang.org/doc/install)

---
```
go get github.com/robertkrimen/otto/ast
go get github.com/robertkrimen/otto/token
go get golang.org/x/crypto/sha3
go get github.com/mitchellh/cli
make privacy
```
---
## __Introduction to MPC__

### What is RockEngine?

RockEngine enables users to perform secure multi-party computation of their private data.
That is, given data *x* of user *A* and data *y* of user *B*, they will compute a value *f(x,y)* uniquely determined by their input, where the function *f* can be written as an algorithm under certain constraints. 

The computation is generally performed by a third party, the *server*, which does not learn any information about the input data. The server is assumed to be honest-but-curious, meaning that it can learn the output of the computation, but not the input data.

The computation could also be done by any of the user. In this case, the server is not needed.

The number of parties could be 2 or more.

### How does is work?

The basic principle of the computation can be decomposed as follows:

1. The function *f* is written as a JavaScript algorithm.

2. One of the users compile the JavaScript file into a logical circuit. This circuit has the .re extension in rockengine.

3. *A* uses this logical circuit to create a *garbled circuit*, which is equivalent but every input, operations and outputs are encrypted.
We denote this garbled circuit as *F*.

4. *A* sends *F* to *B*.

5. *A* encrypts its input *x*, whose encrypted form we denote *X*.
Hen he sends it to *B*.

6. *B* encrypts its own input using oblivious transfer with *A* in order to get the right encrypted value without revealing anything about it.
We denote by *Y* this encrypted input.

7. *B* computes *F(X,Y)*, i.e. he runs the garbled circuit on encrypted inputs.
Thus he gets a encrypted output *Z*.

8. *A* sends part of the decryption key *d* to *B* so that he can decrypt a certain part of *Z*. *B* sends to *A* the other part of *Z* which will also decrypt it. The two can share their information if they want.

### Tests

---
```go
// test0 is just a ratio computation of 2 private inputs
go run main.go build Tests/test0.js
```
---

---
```go
Compiled circuit saved to Tests/test0.re
Compilation achieved in  509.655µs
TotalWires 127
XORgates 141
NonXORgates 65
```
---

---
```go
go run main.go run Tests/test0.re Tests/entry0-0.json Tests/entry0-1.json
```
---

---
```go
Interpreted output to party 0
2
Interpreted output to party 1
<nil>
Interpretation achieved in  261.716µs
```
---

---
```go
go run main.go garble Tests/test0.re
// to garble in debug mode
go run main.go garble Tests/test0.re true
```
---

---
```go
// Multiply 2 private 64x64 matrixes
go run main.go build Tests/test7_matrix64.js
go run main.go run Tests/test7_matrix64.re Tests/entry7-0.json Tests/entry7-1.json
go run main.go garble Tests/test7_matrix64.re
```
---

### How is structured RockEngine?

RockEngine decomposes this process into three main parts:
- compilation
- garbling, i.e. the generation of a garbled circuit from a clear circuit,
- execution.

### The packages

At the moment RockEngine includes 7 core packages:

- __types__, a low-level package used to define basic structures and elements used at various stages, mainly during the compilation

- __circuit__, used at every stage of the algorithm, a package providing classes to describe a clear circuit as well as its garbled version.

- __wires__, a low-level package to define all the necessary classes to represent individual wires as used in the compilation.

- __variables__, used during the compilation, which contains functions used to deal with variables and their types.

- __compiler__, which contains the high-level function used for the compilation.

- __interpreter__, a package used to run entries on clear circuit. It is mostly used for debugging purposes.

- __garbler__, the package containing functions to perform the garbling part of the algorithm.

- __engine__, the package dedicated to the last part of the algorithm. It contains communication functions to transfer data and evaluate the circuit in a distributed way.

In addition to this we have:

- __builder__ which contains an executable to compile boolean circuits from JavaScript files.

- __runner__ which contains an executable to run clear circuits (locally, for testing).

- __garbler__ which contains en executable used to garble an already existing circuit.

- __Tests__, a folder used to store test JavaScript files.


### builder

Builder creates boolean circuits from JavaScript program.
Its relies mainly on the package **compile** which contains the functions `CircuitFromJS (path string) (pc.Circuit, error)` and `CircuitFromAST (prog *ast.Program) (pc.Circuit, error)`.

The output format has a *.re* extension.

The executable Builder is built on top of this package and enables a command-line compilation of a JavaScript code.

There is only one mandatory argument to use Builder, which is the path of the JavaScript file.

You can also use Builder with the following flags :
 * __-no_time__   do not print compile time
 * __-circ__      see output of the compiler (warning, this can be difficult to parse and understand, recommended only for debugging purposes)
 * __-ast__       prints the AST, useful for debugging purposes
 * __-cont__      prints the context (types and variables used), useful for debugging purposes
 * __-debug__     prints details of the compilation for debugging purposes


---
```
go run main.go build --help
go run main.go build Tests/test9_mult1024.js
```
---
when you are at the root of the RockEngine repository.

### Circuits

*circuits* is the package describing the complete structure of a circuit such as generated by the compiler and the structures used to complete it into a garbled circuit.

It contains the following files:
- **clear_structures.go**: includes definition of the *Circuit* structure and structures of its components, i.e. all things useful to contain the binary circuit as given by the compilator (see below).
- **garbled_structures.go**: includes definition of all objects which are used to described the garbled part of the circuit (see below).
- **hash.go**: provides an abstraction around the hashing function used for both the garbling of the gates and their evaluation, providing a common ground for packages *garble* and *execution*.
- **printutils.go** contains methods to output a text version of any object defined in this package to the standard output.

#### Description of a command

---
```
type Command struct {
	Kind CommandType
	X    typ.Num
	Y    typ.Num
	To   typ.Num
}
```
---
Every command describes a basic operation.
The field *Kind* which is a byte describes the nature of the command.
Depending of the value of this field, the command has the following effect:

+ `EMPTY_COMMAND`: used as an error, every other field is blank.

+ `COPY`: a first wire whose index is given by *X* will be copied to another wire whose index is *To*.

+ `FUNCTION_CALL`: the function number *X* is called. If the field *Y* is nonzero then the function is called *Y* times.

+ `INPUT`: the party number *X* sends an input to the wire of number *To*.

+ `OUTPUT`: the wire number *X* is sent to party *To* as an output of the circuit.

+ `MASS_COPY`: the wires from *X* to *X + Y - 1* are copied to the wires from *To* to *To + Y - 1*.

+ `MASS_INPUT`: the party *X* inputs wires from *To* to *To + Y - 1*.

+ `MASS_OUTPUT`: the wires from *X* to *X + Y - 1* are sent to party *To*.

+ `REPLICATE`: the wire *X* will be copied to wires from *To* to *To + Y - 1*.

+ `GATE_n` where `n` is a number between 0 and 15: describes a gate whose truth table is given by the binary representation of `n`.
	The gate takes wires `X` and `Y` (in that order) as entries and outputs the result to wire `To`.
	In particular `GATE_6` represents the XOR gate and is therefore sometimes used in different ways as part of the Free-XOR algorithm.


#### Clear structures

The following types are defined in the file *clear_structures.go* and are part of the binary circuit resulting of the compilation.

A `Function` is the representation in the circuit of a function in the JavaScript source code, including the body of the source code.
Thus, it is defined as a slice of commands and two integers: one to count the number of non-XOR gates which are used inside (which are the costly operations) and another to count the number of XOR gates (or other commands like copy or input).

A `Var` object represents a variable, it embeds a pointer to a `Type` as defined in the *types* package and includes a number which is the number of the first wire belonging to this variable.

Finally the `Circuit` structure represents a whole circuit.
It embeds a `Function` which is the main function, i.e. with commands which are directly in the body  of the code.
Its other fields are:
+ `Parties`: an integer to represent the number of people involved in the computation, usually two.

+ `IntSize`: the default size used for integers in the circuit.

+ `TotalWires`: the total number of wires required to run this circuit.

+ `Inputs` and `Inputs`: two slices of pointers to `Var` objects describing the input and output variables of the circuit.

+ `Funcs`: a slice of pointers to `Function` objects which represents the functions defined in the code.

#### Garbled structures

- `GarbledKey` which is the key of a wire, i.e. the part encoding its actual value, used for decryption of the table entries.
	The length in bytes of every key is the security parameter of the circuit.

- `GarbledValue` which is made of a boolean and a *GarbledKey*.
	The former is the permutation bit, used to select the entry for decryption.

- `GarbledTable` which represents a table used at each non-XOR gate.
	We use the reduced form and the tables are therefore made of three *GarbledValues*, and not four as in the classical form because the first value is always the zero key.

- `TableSet` which is a slice of `GarbledTable`s.
	A `TableSet` needs to be associated with a `Circuit` object to give a complete representation of a garbled circuit.
	This is the first output of the garbling algorithm.
	Indeed, a `TableSet` contains the garbled tables which are necessary for the evaluation of a circuit, but without any information on the wires which are meant to be used with those tables or the other commands in the circuit.

- `UserEncoder`: a slice of `GarbledValue` corresponding to one party.
	Every `GarbledValue` in it is the encoding of one bit of the party's input.

- `EncodingSet`: contains a slice of `UserEncoder`, one for each party, and the secret key of the circuit which is the common offset between the encoding of values *true* and *false* of each bit.
	This is the second output of the garbling algorithm and is used to provide the encoding of every input.

- `DecodingKey` which is made of two booleans and used to decode one output value.

- `UserDecoder`: a slice of `DecodingKey` which corresponds to a certain party and contains the information necessary to decrypt every output directed to this party.

- `DecodingSet`: a slice of `DecodingKey`, one for each party.
	This is the third output of the garbling algorithm and can be used to decrypt all the outputs of the circuit.

### Compiler

compiler is the main package of the compiler.
It contains the functions **CircuitFromAST** and **CircuitFromJS** which are called to create a circuit.
*CircuitFromJS* turns a JavaScript code into a circuit. It uses *CircuitFromAST* which creates the circuit directly from an AST whose format is given in **github.com/robertkrimen/otto/ast**.

The files included are the following:
+ __circuitgenerator.go__ the entry file with the main functions.
+ __utils.go__ with various functions.
+ __operators.go__ contains the functions from in charge of producing the gates to perform basic operations on numbers (`==`, `<`, `+`, `-`, `*`, `/`).
+ __gatesoperations.go__ : some low-level functions from to deal with gates.
+ __writers.go__ contains some methods to add gates to the circuit.
+ __expressionoutputs.go__ contains functions to produce output while receiving a node implementing the otto ast.Expression interface
+ __statementoutputs.go__ contains functions to produce output while receiving a node implementing the otto ast.Statement interface.
+ __wireutils.go__ contains functions to act on wires.

###  Engine

This package provides functions dedicated to the actual transfer and evaluation of the circuit.

*Execution* contains (or will will contain) the following files :

- **evaluation.go** which includes the functions `Evaluate` which is at the core of the algorithm and evaluates the actual results of a circuit using channels of data to garantee flexibility.

- **sender.go** which contains the entry functions to be used on the side of the user which is in charge to encrypt the circuit. The public functions are:
  + `ComputeCircuit` which takes as argument an already compiled circuit, an input and an identifier of the receiver. The function will perform the computation by garbling and sending the circuit to the receiver for evaluation.
  + `Compute`which performs a similar operation execept that the first argument provided is not the circuit itself but the path to a file with extension *.js* or *.freeg*. The circuit will then be compiled if necessary and then the computation will take place.

- **receiver.go** which contains the public functions to be used on the side of the receiver. Note that it uses the essential function `Evaluate` which is in a separate file.

- **oblivioustransfer.go** which contains functions needed to perform oblivious transfer used by both the sender and the receiver.

### Garbler

This package is the core package used by the application GPE.garble.
It contains functions which enable us to transform a plain circuit as defined in GPE.build into a garbled circuit.

It relies on the following internal packages:
- plainCircuit from GPE.build, which contains the description of clear circuit,
- garbledCircuit which contains the description of a garbled circuit as we want to produce.

The package possess two exported functions:
- `SetParams` which is called in GPE.garble to define some parameters to be used during the transformation.
- `Garble (Cin pc.Circuit, n uint8) (gc.Circuit, gc.EncodingFunction, gc.DecodingFunction)`, the main function. `Cin` is the clear circuit given as input, `n` is the security parameter. The function returns a classival tuple *(F,e,d)* where *F* is the garbled function (i.e. the proper garbled circuit), *e* is the encoding function used to get garbled inputs and *d* is the decoding function used to get clear outputs from garbled wires.


##### The security parameter

The value of `n` which is an integer defines the number of bytes on which the value of each wire is encoded.
The minimum value is 1, which corresponds to 8 bits.
As it already multiply by 8 the normal number of bits required we use it as a default value.

### Interpreter

*interpreter* is the package which enables the interpretation of clear boolean circuit from JavaScript inputs.
It also contains the functions used to extract data from json files and to convert output data back to usual variables.

It contains the following files:
- __interpreter.go__, which contains the core function *Interprete* and some auxiliary functions.
- __input_extract.go__ which contains the function *FindInput* used to extract data from the json entries, and other functions used in the first one.
- __results.go__ which contains functions used at the end of the interpretation to output the results to the standard output.

### types

*types* is the most low-level package of the RockEngine.
It defines the basic types (integers, arrays...) which are then used in other packages.

The files contained are the following:
+ __type.go__: defines the structure named `Type` and the basic functions related.
+ __typeutils.go__: defines functions `MaxType` and `CheckRecursiveObj`.
+ __printutils.go__: contains functions to print objects from package robertkrimen/otto.

The first important definition is that of `Num` as an equivalent of `uint32`.
This provides an abstract class to deal with everything which is meant to represent the number of a wire or the length in bits of a variable notably.
Thus, we could change it to any other unsigned integer type and this would still work in any other related package.
The second important definition is that of `Type` which is a structure.

---
```
type Type struct {
	BaseType VarType
	L        Num
	SubType  *Type
	List     []*Type
	Keys     []string
}
```
---
where all the fields are not always used and they don't always have the same meaning.

The field `BaseType`, which is a byte, can take the values `VOID`, `BOOL`, `INT`, `UINT`, `ARRAY`, `OBJECT` and `FUNCTION` which indicate the general category of the `Type` object.

The field `L` represent the length in bits for integer and unsigned integer and the length for an array.

The field `Keys` is used only for object types and contains the names of the sub-fields.

The field `List` can be used both for object types to contain the types of the sub-fields, or for function types to contain the types of the parameters.

### Variables

variables is an intermediate-level package in which we define classes to represent variables during the compilation.

I contains the following files:

+ __variable.go__ : content related to the general Variable class.
+ __boolvariable.go__: class and methods for booleans.
+ __intvariable.go__ : class and methods for integers.
+ __arrayvariable.go__ : class and methods for arrays.
+ __objectvariable.go__ : class and methods for objects.
+ __functionvariable.go__ : class and methods for functions.
+ __function_context.go__ : defines the FunctionContext type and functions on it.
+ __prog_context.go__ : defines the ProgramContext type and contains functions used to generate the context used by the compiler.
+ __typechecks.go__ : contains functions used to check that every operation is correct regarding the types of the variables used.

---
```
type Variable struct {
	*typ.Type
	Name     string
	isperm   bool
	isconst  bool
}
```
---
The field `isperm` takes value true when the variable is permanent, meaning that it is a variable defined by the user and that it has its reserved set of wire numbers.
The field `isconst` takes value true when the variable is constant, meaning that its value is known at compilation time and need not allocating any wire.

The more advanced kind of variables which embed this first type are:
- `BoolVariable` to represent booleans.
- `IntVariable` to represent signed integers, unsigned integers and booleans.
- `ArrayVariable` for arrays.
- `ObjectVariable` for objects.
- `FunctionVariable` for functions.

The following methods apply to all objects implemented `VarInterface`:

- `Size` which returns the size in bits of a variables, in practice the Size method for `Type` is called.
- `Print` which prints all useful infos to the standard output.
- `FillInWires` which is used to make every wire exists, i.e. it creates them when needed or take them from a `WirePool`
- `AssignPermWires` which will assign wire numbers to every wire of a variable, starting from the value given as an argument.
- `Lock`: every wire of the variable is now locked and in particular it cannot be freed if it is in a `WirePool`.
- `Unlock`: unlocks every wire of the variable.
- `GetWire`: returns the _i_-th wire of the variable where _i_ is the argument sent to the function.
- `IsInput`: based on the name of the variable it returns `true` if it is an input variable ans `false` otherwise.
- `IsOutput`: similar but for outputs.

The other methods are getters, setters or simple functions to assess the type of the variable.

#### Function context and program context

A object of type `FunctionContext` is used to describe the set of variables that are specifically used inside a certain function, that is: the parameters of the function, the variables that are defined in the body of the function and a special return variable generated when there is a return statement.

It is therefore defined as `map[string]VarInterface` where the strings used to access the variables are the name of those variables.Then, the type `ProgramContext` represents the whole set of variables used in the program.
It is defined as:

---
```
type ProgramContext struct {
	FunctionContext
	Funcs map[string]FunctionContext
}
```
---
where the embedded `FunctionContext` represents the main function of the program and the other ones in the `Funcs` field stand for the auxiliary functions defined in the code.
The variables which are defined in the main function can also be accessed in the other functions.

### Wires

Wires is a low-level package in which we define all the necessary classes to deal with wires during the computation.

I contains the following files:

+ __wire.go__: defines basic types, see below.
+ __shortcuts.go__: which is used when an operation between two wires is done but one has a constant value so we find a way not to generate a gate.
+ __shortcututils.go__: auxiliary functions used in the latter file.
+ __wirepool.go__: defines operations on pools of wires, see below.

---
```
type Wire struct {
	State     WireState
	Number    int
	Other     *Wire
	Locked    bool
	RefsToMe  WireSet
}
```
---
where `WireState` values are bytes used to represent a certain state of the wire and `WireSet` represents simply a set of wire adresses, defined as a slice of wire pointers.

The field *Number* is a crucial information that identifies uniquely every wire and which is used when the compiler writes gates in the circuit.

The state of the wire can be of six types:
- *ZERO*, represents a bit whose value is 0
- *ONE*, represents a bit whose value is 1
- *UNKNOWN*, represents wires that depend on input values such that their value cannot be computed at compile time
- *UNKNOWN_INVERT*, represents an unknown wire but at some point was inverted
- *UNKNOWN_OTHER*, wire whose values is pointer to another wire value
- *UNKNOWN_INVERT_OTHER*, wire whose values is pointer to the inversion of another wire value

If a wire *w1* points to another wire *w2*, the address of this wire is contained in the field *w1.Other*.
Moreover the address of *w1* will be contained in *w2.RefsToMe*.
Let *i* be the index of *&w1* in *w2.RefsToMe*.


#### Wirepools

The WirePool structure is used to stock wires used in the circuit.
It basically consists of two wire bins, the first is used wires and the second is free wires.
When a wire is needed in the circuit it is in the used bin and in the free bin otherwise.
Thus when a new wire is required we take it in the free bin if there are enough available.


Therefore it avoids creating new wires with new numbers for every operation and reduces the total number of wires in the circuit.

