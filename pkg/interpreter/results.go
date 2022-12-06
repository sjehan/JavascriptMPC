package interpreter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
)

func SaveOutput(outp *circ.UserInOut, t *typ.Type, filepath string) {
	file, err := json.Marshal(GetGoValue(outp, t))
	if err != nil {
		fmt.Println("Error with Marshal function in SaveOutput")
		fmt.Println(err)
		os.Exit(64)
	}
	ioutil.WriteFile(filepath, file, 0644)
}

func GetGoValue(outp *circ.UserInOut, t *typ.Type) interface{} {
	var x interface{} = nil
	switch t.BaseType {
	case typ.BOOL:
		x = (*outp)[0]
	case typ.INT:
		x = GetGoInt(outp, t.L)
	case typ.UINT:
		x = GetGoUInt(outp, t.L)
	case typ.ARRAY:
		x = GetGoArray(outp, t.L, t.SubType)
	case typ.OBJECT:
		x = GetGoObj(outp, t)
	}
	return x
}

func GetGoInt(outp *circ.UserInOut, size typ.Num) interface{} {
	if size > 64 {
		return GetGoFloat(outp, size)
	} else {
		var x int64 = 0
		for i := typ.Num(0); i < size-1; i++ {
			if (*outp)[i] {
				x += 1 << i
			}
		}
		// The negative part
		if (*outp)[size-1] {
			x -= 1 << (size - 1)
		}
		return x
	}
}

func GetGoFloat(outp *circ.UserInOut, size typ.Num) interface{} {
	var x float64 = 0
	for i := typ.Num(0); i < size-1; i++ {
		if (*outp)[i] {
			tmp := 1 << i
			x += float64(tmp)
		}
	}
	// The negative part
	if (*outp)[size-1] {
		tmp := 1 << (size - 1)
		x -= float64(tmp)
	}
	return x
}

// UIntFromBuf prints a non negative integer from the buffer
func GetGoUInt(outp *circ.UserInOut, size typ.Num) interface{} {
	if size > 64 {
		return GetGoPosFloat(outp, size)
	} else {
		var x uint64 = 0
		for i := typ.Num(0); i < size; i++ {
			if (*outp)[i] {
				x += 1 << i
			}
		}
		return x
	}
}

func GetGoPosFloat(outp *circ.UserInOut, size typ.Num) interface{} {
	var x float64 = 0
	for i := typ.Num(0); i < size; i++ {
		if (*outp)[i] {
			tmp := 1 << i
			x += float64(tmp)
		}
	}
	return x
}

func GetGoArray(outp *circ.UserInOut, len typ.Num, item_t *typ.Type) interface{} {
	var x []interface{} = make([]interface{}, 0)
	item_len := item_t.Size()
	for i := typ.Num(0); i < len-1; i++ {
		x = append(x, GetGoValue(outp.SubUIO(i*item_len, item_len), item_t))
	}
	return x
}

func GetGoObj(outp *circ.UserInOut, t *typ.Type) interface{} {
	var x map[string]interface{} = make(map[string]interface{})
	var ind typ.Num = typ.Num(0)
	for i, st := range t.List {
		x[t.Keys[i]] = GetGoValue(outp.SubUIO(ind, st.Size()), st)
		ind += st.Size()
	}
	return x
}

/*
 * The following functions are used to display directly the result on the standard output
 */

// PrintResult is used to output the result of the interpreter to the standard
// output under the form specified in the entry file by the initializer of the
// variable out_* where * is the party number.
func PrintResult(outp *circ.UserInOut, t *typ.Type) {
	//printData(outp, t)
	fmt.Println(GetGoValue(outp, t))
}

// Prints an object of any type from the buffer
func printData(outp *circ.UserInOut, t *typ.Type) {
	switch t.BaseType {
	case typ.BOOL:
		printBool(outp)
	case typ.INT:
		printInt(outp, t.L)
	case typ.UINT:
		printUInt(outp, t.L)
	case typ.ARRAY:
		printArray(outp, t.L, t.SubType)
	case typ.OBJECT:
		printObj(outp, t)
	}
}

// BoolFromBuf prints a boolean from the buffer
func printBool(outp *circ.UserInOut) {
	fmt.Print((*outp)[0])
}

// IntFromBuf prints a (relative integer) from the buffer
func printInt(outp *circ.UserInOut, size typ.Num) {
	if size > 64 {
		printFloat(outp, size)
	} else {
		x := 0
		for i := typ.Num(0); i < size-1; i++ {
			if (*outp)[i] {
				x += 1 << i
			}
		}
		// The negative part
		if (*outp)[size-1] {
			x -= 1 << (size - 1)
		}
		fmt.Print(x)
	}
}

// UIntFromBuf prints a non negative integer from the buffer
func printUInt(outp *circ.UserInOut, size typ.Num) {
	x := 0
	for i := typ.Num(0); i < size; i++ {
		if (*outp)[i] {
			x += 1 << i
		}
	}
	fmt.Print(x)
}

// IntFromBuf prints a big number from the buffer
func printFloat(outp *circ.UserInOut, size typ.Num) {
	x := float64(0)
	for i := typ.Num(0); i < size-1; i++ {
		if (*outp)[i] {
			tmp := 1 << i
			x += float64(tmp)
		}
	}
	// The negative part
	if (*outp)[size-1] {
		tmp := 1 << (size - 1)
		x -= float64(tmp)
	}
	fmt.Print(x)
}

// ArrayFromBuf prints an array of a given type from the buffer
func printArray(outp *circ.UserInOut, len typ.Num, item_t *typ.Type) {
	fmt.Print("[")
	for i := typ.Num(0); i < len-1; i++ {
		printData(outp, item_t)
		fmt.Print(", ")
	}
	printData(outp, item_t)
	fmt.Println("]")
}

// ObjFromBuf prints an object of a given type from the buffer
func printObj(outp *circ.UserInOut, t *typ.Type) {
	fmt.Print("{ ")
	for i, st := range t.List {
		fmt.Print(t.Keys[i], ":")
		printData(outp, st)
		fmt.Print(", ")
	}
	fmt.Print("}")
}
