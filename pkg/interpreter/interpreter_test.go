package interpreter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/robertkrimen/otto/ast"
)

var bl0 ast.BooleanLiteral = ast.BooleanLiteral{
	Literal: "false",
	Value:   false,
}

var bl1 ast.BooleanLiteral = ast.BooleanLiteral{
	Literal: "true",
	Value:   true,
}

var nl1 ast.NumberLiteral = ast.NumberLiteral{
	Literal: "17",
	Value:   int64(17),
}

var nl2 ast.NumberLiteral = ast.NumberLiteral{
	Literal: "127",
	Value:   int64(127),
}

var nl3 ast.NumberLiteral = ast.NumberLiteral{
	Literal: "132",
	Value:   int64(132),
}

var al ast.ArrayLiteral = ast.ArrayLiteral{
	Value: []ast.Expression{&nl1, &nl2},
}

var ol ast.ObjectLiteral = ast.ObjectLiteral{
	Value: []ast.Property{
		ast.Property{Value: &bl0},
		ast.Property{Value: &nl3},
		ast.Property{Value: &al},
	},
}

func TestBooleanToBuf(t *testing.T) {
	fmt.Println("Starting TestBooleanToBuf")
	buf := new(bytes.Buffer)
	expressionToBuf(&bl1, buf)
	PrintBuf(buf)
	fmt.Println()
}

func TestNumberToBuf(t *testing.T) {
	intSize = 8
	fmt.Println("Starting TestNumberToBuf")
	buf := new(bytes.Buffer)
	expressionToBuf(&nl1, buf)
	PrintBuf(buf)
	fmt.Println()
}

func TestArrayToBuf(t *testing.T) {
	intSize = 8
	fmt.Println("Starting TestArrayToBuf")
	buf := new(bytes.Buffer)
	expressionToBuf(&al, buf)
	PrintBuf(buf)
	fmt.Println()
}

func TestObjectToBuf(t *testing.T) {
	intSize = 8
	fmt.Println("Starting TestObjectToBuf")
	buf := new(bytes.Buffer)
	expressionToBuf(&ol, buf)
	PrintBuf(buf)
	fmt.Println()
}
