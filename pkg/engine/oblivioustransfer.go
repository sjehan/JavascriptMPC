package engine

import (
	"crypto/elliptic"
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	"math/big"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/sha3"
)

/* Protocol from "Efficient and Universally Composable Protocols
 * for Oblivious Transfer from the CDH Assumption" */

var n uint8            // This is the global security parameter, i.e., length in bytes of garbled keys
var RandGen *rand.Rand // The generator for random numbers

// max is a usual maximum function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// This function, which has to be called is used to initialized basic parameters
func Init(newN uint8) {
	n = newN
	RandGen = rand.New(rand.NewSource(time.Now().UnixNano()))
}

/******** Functions on elliptic curves ***********/

var cur elliptic.Curve = elliptic.P224()     // The elliptic curve used
var par *elliptic.CurveParams = cur.Params() // The parameters of the curve
var order *big.Int = par.N                   // The order of the group used
var byteSize int = par.BitSize / 8           // Basically the size in bytes of the former

var Gsize int = max(64, byteSize)
var Hsize int = max(int(n)+1, Gsize)

// The use of the Element type provides an abstract way to manipulate the other functions
// independently of the underlying group which is used
type Element struct {
	X, Y *big.Int
}

// NewElement returns a new element of the kind defined above
func NewElement() Element {
	return Element{new(big.Int), new(big.Int)}
}

// Bytes returns a representation as bytes of an element
func (e Element) Bytes() []byte {
	return elliptic.Marshal(cur, e.X, e.Y)
}

// BytesToElement is the revert function of the Bytes methods
func BytesToElement(a []byte) Element {
	x, y := elliptic.Unmarshal(cur, a)
	if x == nil {
		fmt.Println("Abort FromBytes: incorrect data")
		os.Exit(64)
	}
	return Element{x, y}
}

// BytesToBig is a function used to convert bytes to a big integer in a
// way such that this integer can be used as an exponent of the generator
// of the group
func BytesToBig(a []byte) *big.Int {
	b := new(big.Int).SetBytes(a[:byteSize])
	return b.Mod(b, order)
}

// Invert provides a simple way to compute the inverse of an element
func Invert(e Element) Element {
	e.Y.Sub(par.P, e.Y)
	return e
}

// Mult provides the equivalent of a multiplication between two elements
// in a multiplicative group
func Mult(e1, e2 Element) Element {
	x, y := cur.Add(e1.X, e1.Y, e2.X, e2.Y)
	return Element{x, y}
}

// BaseExp provides the equivalent of an exponentiation of the group basis
// in a multiplicative group
func BaseExp(k *big.Int) Element {
	x, y := cur.ScalarBaseMult(k.Bytes())
	return Element{x, y}
}

// Exp provides the equivalent of an exponentiation of a given element
// in a multiplicative group
func Exp(e Element, k *big.Int) Element {
	x, y := cur.ScalarMult(e.X, e.Y, k.Bytes())
	return Element{x, y}
}

// Prints to the standard output information about the given element object
func (e Element) Print(indent string) {
	fmt.Println(indent, "{", e.X, ",", e.Y, "}")
}

/******** Encoding and decoding functions ***********/

// Encode crypts a message M using a key
func Encode(key, M circ.GarbledValue) circ.GarbledValue {
	return key.XOR(M)
}

// Decode decrypts an encrypted message e using a key
func Decode(key, e circ.GarbledValue) circ.GarbledValue {
	return key.XOR(e)
}

// H is a function as defined in the original algorithm which uses a hash function
// to create a garbled value from an element and a sequence of bytes
func H(base []byte, a Element) circ.GarbledValue {
	h := make([]byte, Hsize)
	sha3.ShakeSum256(h, append(base, a.Bytes()...))
	return circ.GarbledValue{h[n]&1 == 1, h[:n]}
}

// G is, as defined in the original algorithm, a mapping from the group used to itself
// which relies on a hash function
func G(a Element) Element {
	h := make([]byte, Gsize)
	sha3.ShakeSum256(h, a.Bytes())
	return BaseExp(BytesToBig(h))
}

/******** Actual top level functions for oblivious transfer ***********/

// Sender is a type used to perform all operations for the one who
// garbles the circuit and posess the keys
type Sender struct {
	y     *big.Int
	Hbase []byte
	T     Element
}

// Receiver is a type used to perfrom all operations for the one who runs
// the circuit and has to find out the encrypted version of its own inputs
type Receiver struct {
	c  bool // The actual input boolean
	vR circ.GarbledValue
}

// NewSender returns a pointer to a new Sender object
func NewSender() *Sender {
	if n == 0 {
		fmt.Println("Execution package not initialized")
		os.Exit(64)
	}
	return &Sender{new(big.Int), make([]byte, 0), NewElement()}
}

// NewReceiver returns a pointer to a new Receiver object
func NewReceiver() *Receiver {
	if n == 0 {
		fmt.Println("Execution package not initialized")
		os.Exit(64)
	}
	return &Receiver{}
}

// This is the initial step of the process when the sender randomly generates
// some values which will be used for encryption
func (sd *Sender) Step0() []byte {
	sd.y.Rand(RandGen, order)
	S := BaseExp(sd.y)
	sd.T = G(S)
	sd.Hbase = S.Bytes()
	return sd.Hbase
}

// This is the first step for the receiver when using the input value C, the
// receiver will generate all values necessary for decryption and send an
// element to the sender
func (rc *Receiver) Step1(Sdata []byte, C bool) []byte {
	var x *big.Int = new(big.Int)
	rc.c = C

	// We pick a random x
	x = x.Rand(RandGen, order)

	// We compute R, S and T when needed
	S := BytesToElement(Sdata)
	R := BaseExp(x)
	if C {
		R = Mult(R, G(S))
	}

	// We compute the key
	rc.vR = H(append(S.Bytes(), R.Bytes()...), Exp(S, x))

	return R.Bytes()
}

// This is the following step when the sender will use the element given by the receiver to
// encrypt the garbled value then send it to the receiver
func (sd *Sender) Step2(Rdata []byte, m0, m1 circ.GarbledValue) (v0, v1 circ.GarbledValue) {
	R := BytesToElement(Rdata)
	sd.Hbase = append(sd.Hbase, R.Bytes()...)

	e0 := Exp(R, sd.y)
	e1 := Mult(e0, Invert(Exp(sd.T, sd.y)))
	k0 := H(sd.Hbase, e0)
	k1 := H(sd.Hbase, e1)

	v0 = Encode(k0, m0)
	v1 = Encode(k1, m1)
	return v0, v1
}

// This is the final step in which the receiver will be able to decrypt one of the
// two values received to get its input
func (rc *Receiver) Step3(v0, v1 circ.GarbledValue) circ.GarbledValue {
	if !rc.c {
		return Decode(rc.vR, v0)
	} else {
		return Decode(rc.vR, v1)
	}
}
