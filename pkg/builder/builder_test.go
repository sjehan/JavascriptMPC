package builder

import (
	"encoding/gob"
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	"os"
	"testing"
)

func TestFileReading(t *testing.T) {
	var circuit = new(circ.Circuit)
	file, err := os.Open("example1.re")
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(circuit)
		if err == nil {
			fmt.Println(circuit.Parties)
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}
	file.Close()
}
