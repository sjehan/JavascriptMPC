package circuit

import (
	"encoding/gob"
	"fmt"
	typ "ixxoprivacy/pkg/types"
	"os"
	"testing"
)

var c1 Command = Command{GATE_10, 1, 2, 3}
var c2 Command = Command{COPY, 1, 0, 3}

var k1 GarbledKey = []byte{128}
var gv1 GarbledValue = NewGarbledValue(true, k1)
var k2 GarbledKey = []byte{10}
var gv2 GarbledValue = NewGarbledValue(false, k2)
var k3 GarbledKey = []byte{133}
var gv3 GarbledValue = NewGarbledValue(false, k3)

var dk1 DecodingKey = [2]bool{true, true}
var dk2 DecodingKey = [2]bool{false, true}
var dk3 DecodingKey = [2]bool{true, false}

var v1 Var = Var{typ.NewBoolType(), 0}
var v2 Var = Var{typ.NewIntType(8), 1}

func mTestGarbledValue(t *testing.T) {
	fmt.Println("Starting TestGarbledValue")
	gvXOR := gv1.XOR(gv2)

	gv1.Print("")
	gv2.Print("")
	gvXOR.Print("")
	fmt.Println()

	var tab GarbledTable = [3]GarbledValue{gv1, gv2, gv3}
	tab.Print("\t")
}

func mTestEandD(t *testing.T) {
	fmt.Println("\nStarting TestEandD")
	r := k3
	enc := NewEncodingSet(r)
	dec := NewDecodingSet()

	enc.User[0] = append(enc.User[0], gv1)
	enc.User[1] = append(enc.User[1], gv2)
	enc.User[1] = append(enc.User[1], gv3)
	enc.Print("")

	dec.User[0] = append(dec.User[0], dk1)
	dec.User[1] = append(dec.User[1], dk2)
	dec.User[1] = append(dec.User[1], dk3)
	dec.Print("")
}

func mTestRandom(t *testing.T) {
	fmt.Println("\nStarting TestRandom")
	RandomGarbledKey(2).Print("\t")
	RandomGarbledKey(4).Print("\t")
	RandomGarbledKey(8).Print("\t")
	fmt.Println()

	RandomGarbledValue(1).Print("\t")
	RandomGarbledValue(2).Print("\t")
	RandomGarbledValue(3).Print("\t")
}

func mTestHash(t *testing.T) {
	fmt.Println("\nStarting TestHash")
	HashGate(k1, k2, 11, 1).Print("\t")
	HashGate(k1, k2, 11, 3).Print("\t")
	fmt.Println()

	HashGate(k1, k2, 10, 1).Print("\t")
	HashGate(k1, k2, 10, 3).Print("\t")
}

func mTestVisit0(t *testing.T) {
	fmt.Println("\nStarting TestVisit0")
	C := RetrieveCircuit("../Tests/test0.freeg")

	chcom := make(chan Command, 5)
	go C.Visit(chcom, C.Funcs)
	var com Command

	for i := uint32(0); i < C.XORgates+C.NonXORgates; i++ {
		com = <-chcom
		fmt.Printf("%d on %d", i+1, C.XORgates+C.NonXORgates)
		com.Print("\t")
	}
}

func mTestFuncion(t *testing.T) {
	fmt.Println("\nStarting TestPush")
	f := NewFunctionPt()
	f.PushNonFunctionCall(c1)
	f.PushNonFunctionCall(c2)
	f.Print("")
}

func TestFile(t *testing.T) {
	fmt.Println("\nStarting TestFile")
	f := NewFunctionPt()
	f.PushNonFunctionCall(c1)
	f.PushNonFunctionCall(c2)

	C := NewCircuit(8, 2)
	C.PushNonFunctionCall(c1)
	C.PushNonFunctionCall(c2)

	C.Funcs = append(C.Funcs, f)
	C.Funcs = append(C.Funcs, f)
	C.Print("")

	path := "testFile"
	C.SaveToFile(path)
	Cbis := RetrieveCircuit(path)
	Cbis.Print("")
}

func mTestSandR(t *testing.T) {
	fmt.Println("\nStarting TestSandR")

	f := NewFunctionPt()
	f.PushNonFunctionCall(c1)
	f.PushNonFunctionCall(c2)
	path := "testFile"
	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error in SaveToFile:")
		fmt.Println(err)
		os.Exit(64)
	}
	encoder := gob.NewEncoder(outputFile)
	encoder.Encode(f)
	outputFile.Close()

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error: could not open input file")
		fmt.Println(err)
		os.Exit(64)
	}
	decoder := gob.NewDecoder(file)
	var ff Function
	err = decoder.Decode(&ff)
	if err != nil {
		fmt.Println("Error: could not decode var.")
		fmt.Println(err)
		os.Exit(64)
	}
	file.Close()

	ff.Print("")
}
