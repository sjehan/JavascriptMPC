package variables

import (
	"fmt"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
	"os"
	"strconv"
	str "strings"

	"github.com/robertkrimen/otto/ast"
)

type ProgramContext struct {
	FunctionContext
	Funcs map[string]FunctionContext
}

/*           Variables used in the compilation               */
/*************************************************************/

var PC ProgramContext

var ZeroExt *ExtInt
var OneExt *ExtInt

var intt *typ.Type
var uintt *typ.Type

var FalseV *BoolVariable = NewBoolVariable("false")
var TrueV *BoolVariable = NewBoolVariable("true")

var rotateLeftt *typ.Type
var getWiret *typ.Type
var setWiret *typ.Type

var ReservedFunc map[string]*typ.Type

/*                   Getters                                 */
/*************************************************************/

func GetVoidType() *typ.Type {
	return typ.VoidType
}
func GetIntt() *typ.Type {
	return intt
}
func GetUIntt() *typ.Type {
	return uintt
}
func GetBoolt() *typ.Type {
	return typ.BoolType
}

// IsConversion assess if a word represents a conversion function and if yes
// it returns the type of this function
func IsConversion(a string) (bool, *typ.Type) {
	if str.HasPrefix(a, "int") {
		b := str.TrimPrefix(a, "int")
		if b == "" {
			ft := typ.NewFunctionType(intt)
			ft.AddType(intt)
			return true, ft
		}
		if s, err := strconv.ParseUint(b, 10, 32); err == nil {
			t := typ.NewIntType(typ.Num(s))
			ft := typ.NewFunctionType(t)
			ft.AddType(intt)
			return true, ft
		}
	} else if str.HasPrefix(a, "uint") {
		b := str.TrimPrefix(a, "uint")
		if b == "" {
			ft := typ.NewFunctionType(uintt)
			ft.AddType(intt)
			return true, ft
		}
		if s, err := strconv.ParseUint(b, 10, 32); err == nil {
			t := typ.NewUIntType(typ.Num(s))
			ft := typ.NewFunctionType(t)
			ft.AddType(intt)
			return true, ft
		}
	}
	return false, nil
}

/*      Functions and methods on ProgramContext              */
/*************************************************************/

// NewProgramContext returns a new ProgramContext variable
func NewProgramContext() ProgramContext {
	return ProgramContext{
		FunctionContext: NewFunctionContext(),
		Funcs:           make(map[string]FunctionContext),
	}
}

// Prints all the content of the given ProgramContext
func (pc ProgramContext) Print(indent string) {
	fmt.Println(indent, "Program context:")
	indent = indent + "\t"
	fmt.Println(indent, "Main function:")
	pc.FunctionContext.Print(indent + "|\t")
	for k, f := range pc.Funcs {
		fmt.Println(indent, "Function", k)
		f.Print(indent + "|\t")
	}
}

// GenerateContext is called by OutputCircuit to create the ProgramContext which will be used in the compilation
func GenerateContext(prog *ast.Program, intsize typ.Num, w0, w1 *wr.Wire) ProgramContext {
	PC = NewProgramContext()
	intt = typ.NewIntType(intsize)
	uintt = typ.NewUIntType(intsize)

	FalseV.W = w0
	TrueV.W = w1
	ZeroExt = SimpleExtInt(0)
	OneExt = SimpleExtInt(1)

	rotateLeftt = typ.NewFunctionType(intt)
	rotateLeftt.AddType(intt)
	rotateLeftt.AddType(intt)

	getWiret = typ.NewFunctionType(typ.BoolType)
	getWiret.AddType(intt)
	getWiret.AddType(intt)

	setWiret = typ.NewFunctionType(GetVoidType())
	setWiret.AddType(intt)
	setWiret.AddType(intt)
	setWiret.AddType(typ.BoolType)

	ReservedFunc = map[string]*typ.Type{
		"RotateLeft": rotateLeftt,
		"GetWire":    getWiret,
		"SetWire":    setWiret,
	}

	// First we find all variables declarations in the body
	for _, dec := range prog.DeclarationList {
		if d, ok := dec.(*ast.VariableDeclaration); ok {
			PC.addVarDeclaration(d)
		}
	}

	// Then we find functions declarations
	for _, dec := range prog.DeclarationList {
		if d, ok := dec.(*ast.FunctionDeclaration); ok {
			f := d.Function
			fc := NewFunctionContext()

			// We find the parameter types of the function and put it in the FunctionContext
			fc.GetParams(f.Name.Name, prog, f.ParameterList.List)

			// We find the types of all other variables in the function to complete the FunctionContext
			for _, fdec := range f.DeclarationList {
				fd, ok := fdec.(*ast.VariableDeclaration)
				if !ok {
					fmt.Println("Error in GenerateContext: only variables should be declared inside functions.")
					os.Exit(64)
				}
				fc.addVarDeclaration(fd)
			}
			fc.CheckForRecTypes()
			PC.Funcs[f.Name.Name] = fc

			fv := NewFunctionVariable(f, fc)
			fc[fv.Returnv.GetName()] = fv.Returnv
			PC.FunctionContext[f.Name.Name] = fv
		}
	}
	PC.CheckForRecTypes()
	CheckProgram(prog)
	return PC
}
