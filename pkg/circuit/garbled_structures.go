package circuit

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type GarbledKey []byte // The crypted representation of a bit

type DecodingKey [2]bool // The variable used to decrypt one bit of output

type GarbledValue struct { // The full representation of an encrypted bit
	P   bool // permutation bit
	Key GarbledKey
}

type GarbledTable [3]GarbledValue // The type to represent a garbled gate

type TableSet []GarbledTable // The representation of the garbled part of a circuit

type UserEncoder []GarbledValue // The type of variable used to encrypt input values of one party

type UserDecoder []DecodingKey // The type of variable used to decrypt output values of one party

type EncodingSet struct { // The set of all necessary variables to encode inputs for all users
	SecretKey GarbledKey
	User      []UserEncoder
}

type DecodingSet struct { // The set of all necessary variables to decode outputs for all users
	User []UserDecoder
}

var RandGen *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

/********************** Methods on GarbledValue and GarbledKey ***********************/

// NewGarbledValue returns a pointer to a newly created garbled value
func NewGarbledValue(p bool, key GarbledKey) GarbledValue {
	return GarbledValue{p, key}
}

// RandomGarbledKey creates a random key
func RandomGarbledKey(n uint8) (gk GarbledKey) {
	for len(gk) < int(n) {
		gk = append(gk, byte(RandGen.Uint32()))
	}
	return gk
}

// NullKey creates a key equal to zero with the appropriate size
func NullKey(n uint8) (gk GarbledKey) {
	for len(gk) < int(n) {
		gk = append(gk, byte(0))
	}
	return gk
}

// RandomGarbledValue creates a new GarbledValue whose key and p values are random
func RandomGarbledValue(n uint8) GarbledValue {
	return GarbledValue{RandGen.Intn(2) == 1, RandomGarbledKey(n)}
}

// The XOR method for keys returns the value of the XOR operation between the two keys
func (k0 GarbledKey) XOR(k1 GarbledKey) GarbledKey {
	if len(k0) != len(k1) {
		fmt.Println("Error in XOR of GarbledKey: received keys of different sizes")
		k0.Print("\t")
		k1.Print("\t")
		os.Exit(64)
	}
	kXOR := make([]byte, len(k0))
	for i := 0; i < len(k0); i++ {
		kXOR[i] = k0[i] ^ k1[i]
	}
	return kXOR
}

// The XOR method for garbled values returns the value of the XOR operation between the values of two wires
func (gv0 GarbledValue) XOR(gv1 GarbledValue) GarbledValue {
	return GarbledValue{gv0.P != gv1.P, gv0.Key.XOR(gv1.Key)}
}

// The Copy method returns a new garbled value with the same fields as the first one
func (gv0 GarbledValue) Copy() GarbledValue {
	return GarbledValue{gv0.P, gv0.Key}
}

/***************** Methods on EncodingSet and DecodingSet *****************/

// NewEncodingSet creates a new variable of type EncodingSet
func NewEncodingSet(r GarbledKey, parties uint8) EncodingSet {
	return EncodingSet{r, make([]UserEncoder, parties, parties)}
}

// NewDecodingSet creates a new variable of type DecodingSet
func NewDecodingSet(parties uint8) DecodingSet {
	return DecodingSet{make([]UserDecoder, parties, parties)}
}

// Encode uses a UserEncoder variable to encode one's own input values represented
// by the slice in. The argument r is the offset used in Free-XOR.
func (ue UserEncoder) Encode(r GarbledKey, in *UserInOut) UserEncoder {
	rv := GarbledValue{true, r}
	for i, x := range *in {
		if x {
			ue[i] = ue[i].XOR(rv)
		}
	}
	return ue
}

// Decode uses a UserDecoder variable to return the clear output corresponding to
// the encrypted output obtained by evaluating a circuit.
func (ud UserDecoder) Decode(output []DecodingKey) []bool {
	result := make([]bool, len(output))
	for i, x := range output {
		if !x[0] {
			result[i] = x[1] != ud[i][0]
		} else {
			result[i] = x[1] != ud[i][1]
		}
	}
	return result
}

/***************** Methods on GarbledTable *****************/

// GetValue returns the value of the garble table which is of interest when
// we evaluate a gate. It uses the permute and point technique.
func (gt GarbledTable) GetValue(px, py bool) GarbledValue {
	var r uint8 = 0
	if px {
		r += 2
	}
	if py {
		r += 1
	}
	if r == 0 {
		return gt[0].XOR(gt[0])
	} else {
		return gt[r-1]
	}
}

/********************** Methods on TableSet ***********************/

// SaveToFile saves a circuit into a file whose path is given in
// argument and using standard gobs encoding
func (TS *TableSet) SaveToFile(path string) {
	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error in ")
		fmt.Println(err)
		os.Exit(64)
	}
	encoder := gob.NewEncoder(outputFile)
	encoder.Encode(TS)
	outputFile.Close()
}

// RetrieveFromFile is used to get the circuit from a file generated
// with method SaveToFile
func RetrieveTableSet(path string) TableSet {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error: could not open input file")
		fmt.Println(err)
		os.Exit(64)
	}
	decoder := gob.NewDecoder(file)
	var TS TableSet
	err = decoder.Decode(&TS)
	if err != nil {
		fmt.Println("Error: could not decode circuit.")
		fmt.Println(err)
		os.Exit(64)
	}
	file.Close()
	return TS
}
