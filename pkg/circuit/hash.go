package circuit

import (
	"crypto/sha512"
	"encoding/binary"
)

// We use hash as an abstration for the actual hash function which is sha512
func Hash(data []byte) []byte {
	arr := sha512.Sum512(data)
	return arr[:]
}

// hashGate produces the hash value used in case of a gate
func HashGate(k1, k2 GarbledKey, index uint32, n uint8) GarbledValue {
	data := append(k1, k2...)
	ibytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(ibytes, index)
	data = append(data, ibytes...)
	h := Hash(data)
	return NewGarbledValue(h[n]&1 == 1, h[:n])
}

// hashOut returns the boolean obtained as the first bit of a hash value
func HashOut(k1 GarbledKey, index uint32) bool {
	k2 := []byte("out")
	data := append(k1, k2...)
	ibytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(ibytes, index)
	data = append(data, ibytes...)
	return Hash(data)[0]&1 == 1
}
