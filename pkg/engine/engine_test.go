package engine

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	circ "ixxoprivacy/pkg/circuit"
	compiler "ixxoprivacy/pkg/compiler"
	garble "ixxoprivacy/pkg/garbler"
	ip "ixxoprivacy/pkg/interpreter"
)

var debug bool = false

func basicTest(testName string) {
	fmt.Println("\t Running ", testName)
	security := uint8(4)
	Init(security)

	var testNumber rune = []rune(testName)[4]
	var entryRoot string = "../Tests/entry" + string(testNumber) + "-"
	testName = "../Tests/" + testName
	var reName string = testName + ".re"

	// Compilation of the circuit
	tStart := time.Now()
	C1, err := compiler.CircuitFromJS(testName + ".js")
	if err != nil {
		fmt.Println("Compilation error:")
		fmt.Println(err)
	}
	diff := time.Now().Sub(tStart)
	fmt.Println("\t Compilation done in", diff)

	C1.SaveToFile(reName)
	C2 := circ.RetrieveCircuit(reName)
	if debug {
		C2.Print("\t\t")
	}

	// Garling of the circuit
	tStart = time.Now()
	TS, enc, dec := garble.Garble(C2, security)
	diff = time.Now().Sub(tStart)
	fmt.Println("\t Garbling done in", diff)

	if debug { // Testing if enc and dec seem ok
		fmt.Println("\t Encoder:")
		enc.Print("\t\t")
		fmt.Println("\n\t Decoder:")
		dec.Print("\t\t")
	}

	// Retrieving inputs
	tStart = time.Now()
	inputFiles := make([]string, 0)
	for i := 0; i < int(C2.Parties); i++ {
		inputFiles = append(inputFiles, entryRoot+strconv.Itoa(i)+".json")
	}
	inputs := ip.GetAllInputs(C2.Inputs, inputFiles)
	diff = time.Now().Sub(tStart)
	fmt.Println("\t Getting inputs done in", diff)

	if debug { // Testing if inputs were well retrieved
		fmt.Println("\t Inputs:")
		for i := uint8(0); i < C2.Parties; i++ {
			inputs[i].Print("\t\t")
			fmt.Println()
		}
	}

	// Doing the interpretation in the clear way
	tStart = time.Now()
	ioutputs := ip.Interprete(C2, inputs)
	diff = time.Now().Sub(tStart)
	fmt.Println("\t Clear interpretation done in", diff)

	if debug {
		fmt.Println("\t Clear outputs:")
		for i := uint8(0); i < C2.Parties; i++ {
			if ioutputs[i] != nil {
				ioutputs[i].Print("\t\t")
				fmt.Println()
			}
		}
	}

	// We create the channels to be used
	tStart = time.Now()
	chtab := make(chan circ.GarbledTable, 5)
	chin := make([]chan circ.GarbledValue, 0)
	chout := make([]chan circ.DecodingKey, 0)
	outputs := make([]*circ.UserInOut, 0)
	for i := uint8(0); i < C2.Parties; i++ {
		chin = append(chin, make(chan circ.GarbledValue, 5))
		chout = append(chout, make(chan circ.DecodingKey, 5))
		outputs = append(outputs, new(circ.UserInOut))
	}

	// We send those channels to specific functions and evaluate the circuit
	wg.Add(1 + 2*int(C2.Parties))
	go TabSender(TS, chtab)
	for i := uint8(0); i < C2.Parties; i++ {
		go InputSender(enc.User[i].Encode(enc.SecretKey, inputs[i]), chin[i])
		go OutputReceiver(dec.User[i], chout[i], outputs[i])
	}
	Evaluate(C2, chtab, chin, chout)
	wg.Wait()
	diff = time.Now().Sub(tStart)
	fmt.Println("\t Evaluation done in", diff)

	if debug {
		for party, outp := range outputs {
			if len(*outp) != 0 {
				fmt.Println("\t Wire output to party", party)
				outp.Print("\t\t")
				fmt.Println()
			}
		}
	}

	// We print the outputs if they are incorrect
	for party, out := range outputs {
		if ioutputs[party] != nil && !ioutputs[party].Equals(out) {
			fmt.Println("Difference in results for party", party)
			fmt.Print("\t Clear result\n\t\t")
			ip.PrintResult(ioutputs[party], C2.Outputs[party].Type)
			fmt.Print("\t Garbled result\n\t\t")
			ip.PrintResult(out, C2.Outputs[party].Type)
		} else {
			if !C2.Outputs[party].IsVoid() {
				path := "../Tests/result" + string(testNumber) + "-" + strconv.Itoa(party) + ".json"
				ip.SaveOutput(out, C2.Outputs[party].Type, path)
			}
		}
	}
}

func TestGarbledValue(t *testing.T) {
	fmt.Println("Starting TestBattery")
	basicTest("test0")
	fmt.Println()
	basicTest("test2")
	fmt.Println()
	basicTest("test3")
	fmt.Println()
	basicTest("test5_matrix4")
	fmt.Println()
	basicTest("test6_matrix16")
	fmt.Println()
	basicTest("test8_mult256")
}

func mTestOperations(t *testing.T) {
	Init(2)
	var x *big.Int = new(big.Int)
	x = x.Rand(RandGen, order)
	fmt.Println(x)

	O := BaseExp(big.NewInt(0))
	O.Print("0 = ")

	g := BaseExp(big.NewInt(1))
	g.Print("g = ")

	e1 := BaseExp(big.NewInt(63))
	e1.Print("e1 = ")

	e2 := Mult(e1, e1)
	e2.Print("e2 = ")

	e3 := Invert(e1)
	e3.Print("e3 = ")
	if !cur.IsOnCurve(e3.X, e3.Y) {
		fmt.Println("/!\\")
	}

	e4 := Mult(e2, e3)
	e4.Print("e4 = ")
}

func mTestOT(t *testing.T) {
	fmt.Println("Starting TestOT")
	var n uint8 = 4
	var m0 circ.GarbledValue = circ.RandomGarbledValue(n)
	var m1 circ.GarbledValue = circ.RandomGarbledValue(n)

	Init(n)

	sd := NewSender()
	rc := NewReceiver()

	sdata := sd.Step0()
	rdata := rc.Step1(sdata, true)
	v0, v1 := sd.Step2(rdata, m0, m1)
	m := rc.Step3(v0, v1)

	fmt.Println("Initial messages:")
	m0.Print("\t")
	m1.Print("\t")

	fmt.Println("Crypted messages:")
	v0.Print("\t")
	v1.Print("\t")

	fmt.Println("Decrypted message:")
	m.Print("\t")
}
