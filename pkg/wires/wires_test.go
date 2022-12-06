package wires

import (
	"fmt"
	"testing"

	typ "ixxoprivacy/pkg/types"
)

// Vérifie que l'algorithme de création de clés est bien fonctionnel
func mTestCreateWire(t *testing.T) {
	fmt.Printf("\nStarting test 1\n")
	w := new(Wire)
	w.Print("")

	fmt.Println()
	w1 := new(Wire)
	w2 := new(Wire)
	w3 := new(Wire)
	w.AddRef(w1)
	w.AddRef(w2)
	w.AddRef(w3)
	fmt.Println("refs to me: ", w.Refs())

	w.RemoveRef(w1)
	fmt.Println("refs to me: ", w.Refs())
	w2.FreeRefs()
	fmt.Println("refs to me: ", w.Refs())
}

func mTestStateEnum(t *testing.T) {
	fmt.Printf("\nStarting test 2\n")
	var st WireState = UNKNOWN_INVERT

	stString := st.ToString()
	if stString != "UNKNOWN_INVERT" {
		t.Errorf("Conversion 1 failed, got %v", stString)
	}
	if IntToState(3) != st {
		t.Errorf("Conversion 2 failed, got %d", IntToState(3))
	}
}

// Vérifie que l'algorithme de création de clés est bien fonctionnel
func mTestInvert(t *testing.T) {
	fmt.Printf("\nStarting inversion test:\n")
	var i uint8
	fmt.Println("Changing a")
	for i = 0; i < 16; i++ {
		fmt.Printf("%b -> %b\n", i, InvertTable(false, i))
	}
	fmt.Println("\nChanging b")
	for i = 0; i < 16; i++ {
		fmt.Printf("%b -> %b\n", i, InvertTable(true, i))
	}
}

/*                       Tests on shortcuts                       */
/******************************************************************/

func mTestShortCut(t *testing.T) {
	a := new(Wire)
	b := new(Wire)
	dest := new(Wire)

	a.State = ONE
	b.State = ONE
	var table uint8 = 14

	ok := ShortCut(a, b, table, dest)
	if !ok {
		t.Errorf("Failure\n")
	} else {
		fmt.Println("Success, got ", dest.State.ToString())
	}

	a.State = ZERO
	b.State = UNKNOWN
	table = 14
	dest = new(Wire)
	ok = ShortCut(a, b, table, dest)
	if !ok {
		t.Errorf("Failure\n")
	} else {
		fmt.Println("Success, got ", dest.State.ToString())
	}

	a.State = ZERO
	b.State = UNKNOWN
	table = 8
	dest = new(Wire)
	ok = ShortCut(a, b, table, dest)
	if !ok {
		t.Errorf("Failure\n")
	} else {
		fmt.Println("Success, got ", dest.State.ToString())
	}

	a.State = ZERO
	b.State = UNKNOWN
	table = 5
	dest = new(Wire)
	ok = ShortCut(a, b, table, dest)
	if !ok {
		t.Errorf("Failure\n")
	} else {
		fmt.Println("Success, got ", dest.State.ToString())
	}
}

/*                       Tests on wiresets                        */
/******************************************************************/

func TestWSet(t *testing.T) {
	fmt.Printf("\nStarting TestWSet:\n")
	var n typ.Num = 8
	ws := NewWireSet(n)
	for _, w := range ws {
		w.Print("")
	}
}
