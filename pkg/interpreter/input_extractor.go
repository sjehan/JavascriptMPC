package interpreter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
)

// GetAllInputs takes as entries the whole list of input types and the name of
// json files in which to find them. It then calls GetInput to get every one of them.
func GetAllInputs(inputs []*circ.Var, fileNames []string) []*circ.UserInOut {
	B := make([]*circ.UserInOut, 0)
	for i, v := range inputs {
		B = append(B, GetInput(fileNames[i], v.Type))
	}
	return B
}

// FindInput visits variable declarations of a program parsed by the otto parser
// in order to find the initial value of a variable whose name is provided.
// This value is then converted into bytes representing booleans and written
// in a buffer passed by address to FindInput.
func GetInput(fileName string, t *typ.Type) *circ.UserInOut {
	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var data interface{}
	err = json.Unmarshal(raw, &data)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	inp := circ.NewUIO()
	dataToBuf(data, t, inp)
	return inp
}

// dataToBuf takes an generic data and break it down into bits
// in its auxiliary function to send it to the buffer buf
func dataToBuf(data interface{}, t *typ.Type, inp *circ.UserInOut) {
	switch t.BaseType {
	case typ.VOID:
		fmt.Println("Error, input with type VoidType")
	case typ.BOOL:
		booleanToBuf(data, inp)
	case typ.INT, typ.UINT:
		numberToBuf(data, t, inp)
	case typ.ARRAY:
		arrayToBuf(data, t, inp)
	case typ.OBJECT:
		objectToBuf(data, t, inp)
	default:
		fmt.Println("Error in dataToBuf: base type unrecognized, value received is", t.BaseType)
	}
}

// booleanToBuf sends a boolean encoded value as bits to the buffer buf
func booleanToBuf(data interface{}, inp *circ.UserInOut) {
	b, ok := data.(bool)
	if !ok {
		fmt.Println("Error in booleanToBuf, data does not fit as bool")
	}
	inp.Add(b)
}

// positiveToBuf sends an encoded integer value as bits to the buffer buf
func numberToBuf(data interface{}, t *typ.Type, inp *circ.UserInOut) {
	fval, ok := data.(float64)
	if !ok {
		fmt.Println("Error in numberToBuf: data of uncompatible type")
	}

	val := int64(fval)
	if val >= 0 {
		intToBuf(val, t.L, inp)
	} else {
		if t.BaseType == typ.UINT {
			fmt.Println("Error in numberToBuf: positive expected and received negative number")
		}
		intToBuf((1<<t.L)-val, t.L, inp)
	}
}

// positiveToBuf sends an encoded integer value as bits to the buffer buf
func intToBuf(val int64, size typ.Num, inp *circ.UserInOut) {
	for i := typ.Num(0); i < size; i++ {
		inp.Add(val&(1<<i) != 0)
	}
}

// arrayToBuf sends an array encoded as bits to the buffer buf
func arrayToBuf(data interface{}, t *typ.Type, inp *circ.UserInOut) {
	arr, ok := data.([]interface{})
	if !ok {
		fmt.Println("Error in arrayToBuf: data provided is no array")
		os.Exit(64)
	}
	if len(arr) != int(t.L) {
		fmt.Println("Error in arrayToBuf: the array has improper length")
		os.Exit(64)
	}
	for _, val := range arr {
		dataToBuf(val, t.SubType, inp)
	}
}

// objectToBuf sends an object encoded as bits to the buffer buf
func objectToBuf(data interface{}, t *typ.Type, inp *circ.UserInOut) {
	obj, ok := data.(map[string]interface{})
	if !ok {
		fmt.Println("Error in objectToBuf: data provided is no object")
		os.Exit(64)
	}
	if len(obj) != len(t.List) {
		fmt.Println("Error in objectToBuf: the object has improper length")
		os.Exit(64)
	}
	for i, st := range t.List {
		dataToBuf(obj[t.Keys[i]], st, inp)
	}
}
