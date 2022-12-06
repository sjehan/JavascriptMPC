package garbler

import (
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	"testing"
)

func TestGate(t *testing.T) {
	fmt.Println("Starting TestGate")
	N = 1
	offsetR = circ.RandomGarbledKey(N)
	offsetR.Print("\t Secret key: ")
	fmt.Println()

	w0, w1 := circ.RandomGarbledValue(N), circ.RandomGarbledValue(N)
	w0.Print("\t w0: ")
	fmt.Println()
	w1.Print("\t w1: ")
	fmt.Println()

	gt, w2 := tableFromWires(w0, w1, 5)
	fmt.Println("\tTable:")
	gt.Print("\t")
	fmt.Println()

	w2.Print("\t w2: ")
	fmt.Println()
	dk := outKey(w2)
	dk.Print("\t dk:")
	fmt.Println()

	var wa, wb, wc circ.GarbledValue
	for i := 0; i < 4; i++ {
		fmt.Println("\t i = ", i)
		wa = getVal(w0, i/2 == 1)
		wa.Print("\t\t wa: ")
		wb = getVal(w1, i%2 == 1)
		wb.Print("\t\t wb: ")

		wc = circ.HashGate(wa.Key, wb.Key, gateIndex, N)
		wc.Print("\t\t H(wa,wb): ")
		if wa.P || wb.P {
			wc = wc.XOR(gt.GetValue(wa.P, wb.P))
		}
		wc.Print("\t\t wc: ")

		dkc := circ.DecodingKey{wc.P, circ.HashOut(wc.Key, outIndex)}
		dk.Print("\t\t dkc:")
		if !dkc[0] {
			fmt.Println("\t\t result: ", dkc[1] != dk[0])
		} else {
			fmt.Println("\t\tresult: ", dkc[1] != dk[1])
		}
	}
}
