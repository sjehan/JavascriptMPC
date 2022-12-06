package compiler

import (
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
)

// outputEquals ouputs the gates necessary to test equality bit by
// bit between two variables.
// Precondiction: all vectors are of proper size( |leftv| == |rightv|
// and |destv| == 1)
func outputEquals(leftv, rightv wr.WireSet) *wr.Wire {
	var outputwire *wr.Wire
	var currentxor *wr.Wire
	var t *wr.Wire

	for i, wl := range leftv {
		t = invertWireNoInvertOutput(wl)
		currentxor = outputGateNoInvertOutput(6, t, rightv[i])

		if i > 1 {
			outputGateToDest(8, currentxor, outputwire, outputwire)
		} else if i == 0 {
			outputwire = currentxor
		}
	}
	return outputwire
}

// outputLessThan compares leftv and rightv and returns the wire containing the result
// of this comparison.
// Do do that is subtracts leftv from rightv
// Precondiction: all vectors are of proper size, |leftv| == |rightv| and |destv| >= 1
// (should be == 1 but > will suffice).
func outputLessThan(leftv, rightv wr.WireSet) *wr.Wire {
	length := len(leftv)
	var outputwire *wr.Wire

	if length == 1 {
		return outputGate(4, rightv[0], leftv[0])
	} else {
		carry := pool.GetWire()
		xorab := pool.GetWire()
		xorac := pool.GetWire()
		and1 := pool.GetWire()
		na := pool.GetWire()

		length++
		leftv = append(leftv, leftv[length-2])
		rightv = append(rightv, rightv[length-2])

		for i := 0; i < length; i++ {
			na = invertWireNoInvertOutput(leftv[i])

			xorab = clearWireForReuse(xorab)
			outputGateToDest(6, rightv[i], na, xorab)

			if i < length-1 {
				xorac = clearWireForReuse(xorac)
				outputGateToDest(6, carry, na, xorac)

				and1 = clearWireForReuse(and1)
				outputGateNoInvertOutputToDest(8, xorab, xorac, and1)

				carry = clearWireForReuse(carry)
				outputGateNoInvertOutputToDest(6, na, and1, carry)
			} else {
				t := invertWireNoInvertOutput(xorab)
				outputwire = outputGateNoInvertOutput(6, t, carry)
			}
		}
		// length--
		leftv = leftv[:len(leftv)-1]
		rightv = rightv[:len(rightv)-1]
	}
	return outputwire
}

// outputSubtract computes the difference between leftv and rightv,
// producing gates when it is necessary.
// Precondiction - all vectors are of proper size:
// |leftv| == |rightv| == |destv|
func outputSubtract(leftv, rightv, destv wr.WireSet) {
	length := len(leftv)

	if length == 1 {
		destv[0] = outputGate(6, rightv[0], leftv[0])
	} else {
		carry := pool.GetWire()
		xorab := pool.GetWire()
		xorac := pool.GetWire()
		and1 := pool.GetWire()
		na := pool.GetWire()

		for i := 0; i < length; i++ {
			na = invertWireNoInvertOutput(leftv[i])
			xorab = clearWireForReuse(xorab)
			outputGateToDest(6, rightv[i], na, xorab)

			t := invertWireNoInvertOutput(xorab)
			outputGateNoInvertOutputToDest(6, t, carry, destv[i])

			if i < length-1 {
				xorac = clearWireForReuse(xorac)
				outputGateToDest(6, carry, na, xorac)

				and1 = clearWireForReuse(and1)
				outputGateNoInvertOutputToDest(8, xorab, xorac, and1)

				carry = clearWireForReuse(carry)
				outputGateNoInvertOutputToDest(6, na, and1, carry)
			}
		}
		pool.FreeWire(carry)
		pool.FreeWire(xorab)
		pool.FreeWire(xorac)
		pool.FreeWire(and1)
		pool.FreeWire(na)
	}
}

// outputSubtract computes the sum of leftv and rightv,
// producing gates when it is necessary.
// Precondiction - all vectors are of proper size:
// |leftv| == |rightv| == |destv|
func outputAddition(leftv, rightv, destv wr.WireSet) {
	length := len(leftv)

	if length == 1 {
		outputGateToDest(6, rightv[0], leftv[0], destv[0])
	} else {
		carry := pool.GetWire()
		xorab := pool.GetWire()
		xorac := pool.GetWire()
		and1 := pool.GetWire()

		for i := 0; i < length; i++ {
			xorab = clearWireForReuse(xorab)
			outputGateNoInvertOutputToDest(6, rightv[i], leftv[i], xorab)
			outputGateNoInvertOutputToDest(6, xorab, carry, destv[i])

			if i < length-1 {
				xorac = clearWireForReuse(xorac)
				outputGateToDest(6, carry, leftv[i], xorac)

				and1 = clearWireForReuse(and1)
				outputGateNoInvertOutputToDest(8, xorab, xorac, and1)

				carry = clearWireForReuse(carry)
				outputGateNoInvertOutputToDest(6, leftv[i], and1, carry)
			}
		}
		pool.FreeWire(carry)
		pool.FreeWire(xorab)
		pool.FreeWire(xorac)
		pool.FreeWire(and1)
	}
}

// outputMultSigned computes the multiplication of leftv and rightv when the two
// are signed values, producing gates when it is necessary.
// Only does right side of multiplication trapazoid (i.e. if your mult
// is 32 bits by 32 bits we don't need the result of the left 32 bits
// since it goes back into a 32bit int, not a 64 bit int).
// Precondiction - all vectors are of proper size:
// |leftv| == |rightv| == |destv|
// left is x input, right is y input
// mult algorithm from MIT slides from course 6.111, fall 2012, lecture 8/9, slide 33
func outputMultSigned(leftv, rightv, destv wr.WireSet) {
	length := len(leftv)

	if length == 1 {
		outputGateToDest(8, leftv[0], rightv[0], destv[0])
	} else {
		rowinputsleft := wr.EmptyWireSet(typ.Num(length))
		rowinputsright := wr.EmptyWireSet(typ.Num(length))

		carry := pool.GetWire()
		xorab := pool.GetWire()
		andn := pool.GetWire()

		// number of rows
		for i := 0; i < length-1; i++ {
			// create inputs to each adder
			if i == 0 {
				for k := 0; k < length; k++ {
					rowinputsleft[k] = outputGate(8, leftv[k], rightv[0])
				}
				// only on first row do we do this
				rowinputsleft[length-1] = invertWireNoAllocUnlessNecessary(rowinputsleft[length-1])

				for k := 0; k < length-1; k++ {
					rowinputsright[k] = outputGate(8, leftv[k], rightv[1])
				}
				assignWire(destv[0], rowinputsleft[0])

				// shift down
				for k := 0; k < length-1; k++ {
					assignWire(rowinputsleft[k], rowinputsleft[k+1])
				}
				if i == length-2 {
					rowinputsright[0] = invertWireNoAllocUnlessNecessary(rowinputsright[0])
				}
			} else {
				for k := 0; k < length-1-i; k++ {
					rowinputsright[k] = clearWireForReuse(rowinputsright[k])
					outputGateToDest(8, leftv[k], rightv[i+1], rowinputsright[k])
				}
				// last row
				if i == length-2 {
					rowinputsright[0] = invertWireNoAllocUnlessNecessary(rowinputsright[0])
				}
			}
			// create each adder
			for j := 0; j < length-i-1; j++ {
				//performs the HA or FA
				//output half adder
				if j == 0 {
					// xorab = clearWireForReuse(xorab) // appears not to be useful
					outputGateToDest(6, rowinputsright[0], rowinputsleft[0], destv[i+1])
					// carry = clearWireForReuse(carry) // appears not to be useful

					if i != length-2 {
						outputGateToDest(8, rowinputsright[0], rowinputsleft[0], carry)
					}
				} else { //output full adder
					xorab = clearWireForReuse(xorab)
					outputGateToDest(6, rowinputsright[j], rowinputsleft[j], xorab)

					rowinputsleft[j-1] = clearWireForReuse(rowinputsleft[j-1])
					outputGateToDest(6, xorab, carry, rowinputsleft[j-1])

					if j < length-1-i-1 {
						andn = clearWireForReuse(andn)
						outputGateToDest(6, carry, rowinputsleft[j], andn)
						outputGateToDest(8, xorab, andn, andn)

						// carry = clearWireForReuse(carry) // appears not to be useful
						outputGateToDest(6, rowinputsleft[j], andn, carry)
					}
				}

				for k := 0; k < 2; k++ {
					if andn.Refs() == 0 {
						andn.State = wr.ZERO
						andn.FreeRefs()
					}
					if xorab.Refs() == 0 {
						xorab.State = wr.ZERO
						xorab.FreeRefs()
					}
				}
			}
		}
		pool.FreeSinglesIfNoRefs()
	}
}

// outputMultUnsigned computes the multiplication of leftv and rightv when the two
// are unsigned values, producing gates when it is necessary.
func outputMultUnsigned(leftv, rightv, destv wr.WireSet) {
	length := len(leftv)

	if length == 1 {
		destv[0] = outputGate(8, leftv[0], rightv[0])
	} else {
		rowinputsleft := wr.EmptyWireSet(typ.Num(length))
		rowinputsright := wr.EmptyWireSet(typ.Num(length))

		carry := pool.GetWire()
		xorab := pool.GetWire()
		andn := pool.GetWire()

		// number of rows
		for i := 0; i < length-1; i++ {
			//create inputs to each adder
			if i == 0 {
				for k := 0; k < length-i; k++ {
					rowinputsleft[k] = outputGate(8, leftv[k], rightv[0])
				}
				// only on first row do we do this
				for k := 0; k < length-i-1; k++ {
					rowinputsright[k] = outputGate(8, leftv[k], rightv[1])
				}
				assignWire(destv[0], rowinputsleft[0])

				// shift down
				for k := 0; k < length-1; k++ {
					assignWire(rowinputsleft[k], rowinputsleft[k+1])
				}
			} else {
				for k := 0; k < length-1-i; k++ {
					//cout << "> gate\n";
					rowinputsright[k] = clearWireForReuse(rowinputsright[k])
					outputGateToDest(8, leftv[k], rightv[i+1], rowinputsright[k])
				}
				// last row: nothing
			}
			// create each adder
			for j := 0; j < length-i-1; j++ {
				// performs the HA or FA
				// output half adder
				if j == 0 {
					xorab = clearWireForReuse(xorab)
					outputGateToDest(6, rowinputsright[0], rowinputsleft[0], xorab)
					assignWire(destv[i+1], xorab)
					carry = clearWireForReuse(carry)

					if i != length-2 {
						outputGateToDest(8, rowinputsright[0], rowinputsleft[0], carry)
					}
				} else { //output full adder
					xorab = clearWireForReuse(xorab)
					outputGateToDest(6, rowinputsright[j], rowinputsleft[j], xorab)

					rowinputsleft[j-1] = clearWireForReuse(rowinputsleft[j-1])
					outputGateToDest(6, xorab, carry, rowinputsleft[j-1])

					if j < length-1-i-1 {
						andn = clearWireForReuse(andn)
						outputGateToDest(6, carry, rowinputsleft[j], andn)
						outputGateToDest(8, xorab, andn, andn)

						carry = clearWireForReuse(carry)
						outputGateToDest(6, rowinputsleft[j], andn, carry)
					}
				}

				for k := 0; k < 2; k++ {
					if andn.Refs() == 0 {
						clearWireForReuse(andn)
					}
					if xorab.Refs() == 0 {
						clearWireForReuse(xorab)
					}
				}
			}
		}
		pool.FreeSinglesIfNoRefs()
	}
}

// outputDivideUnsigned computes the division or modulus of leftv and rightv when
// the two are unsigned values, producing gates when it is necessary.
// leftv - dividend, rightv - divisor
func outputDivideUnsigned(leftv, rightv, destv wr.WireSet, IsModDiv bool) {
	origlength := len(leftv)
	length := origlength + 1

	carry := pool.GetWire()
	xorab := pool.GetWire()
	xorac := pool.GetWire()
	and1 := pool.GetWire()
	xortout := pool.GetWire()

	t := W_1

	/*extend extra bit for correctness purposes*/
	lleft := append(leftv, W_0)
	lright := append(rightv, W_0)
	ldest := wr.EmptyWireSet(typ.Num(length))

	inputx := wr.EmptyWireSet(typ.Num(length))
	inputy := wr.EmptyWireSet(typ.Num(length))
	remainw := wr.EmptyWireSet(typ.Num(length))

	// divisor
	copy(inputx, lright)
	inputy[0] = lleft[length-1]
	for i := 1; i < length; i++ {
		inputy[i] = pool.GetWire()
	}

	keepwiresA := wr.EmptyWireSet(0)
	keepwiresB := wr.EmptyWireSet(0)

	for i := 0; i < length; i++ {
		// setting each input row
		if i == 0 {
			carry.State = wr.ONE
		} else {
			inputy[0] = lleft[length-1-i]
			for j := 1; j < length; j++ {
				assignWire(inputy[j], remainw[j-1])
			}
			assignWire(carry, t)
		}

		/*controlled add / subtract*/
		for j := 0; j < length; j++ {
			xortout = clearWireForReuse(xortout)
			outputGateNoInvertOutputToDest(6, t, inputx[j], xortout)

			xorab = clearWireForReuse(xorab)
			outputGateNoInvertOutputToDest(6, inputy[j], xortout, xorab)

			//full adder part
			if remainw[j] != nil {
				//save wires from layer i so they can be used at layer i+1 (otherwise each remainw wire requires a new wire, very inefficient)
				if remainw[j].Refs() > 0 {
					if i%2 == 0 {
						keepwiresA = append(keepwiresA, remainw[j])
						if len(keepwiresB) > 0 {
							remainw[j] = keepwiresB.PopBack()
						} else {
							remainw[j] = pool.GetWire()
						}
					} else {
						keepwiresB = append(keepwiresB, remainw[j])
						if len(keepwiresA) > 0 {
							remainw[j] = keepwiresA.PopBack()
						} else {
							remainw[j] = pool.GetWire()
						}
					}
					if remainw[j] == nil {
						remainw[j] = pool.GetWire()
					}
				}

				remainw[j] = clearWireForReuse(remainw[j])
				outputGateToDest(6, xorab, carry, remainw[j])
			} else {
				remainw[j] = outputGate(6, xorab, carry)
			}

			if j < length-1 {
				xorac = clearWireForReuse(xorac)
				outputGateToDest(6, carry, xortout, xorac)

				and1 = clearWireForReuse(and1)
				outputGateNoInvertOutputToDest(8, xorab, xorac, and1)

				carry = clearWireForReuse(carry)
				outputGateNoInvertOutputToDest(6, xortout, and1, carry)
			}

			if xortout.Refs() == 0 {
				clearWireForReuse(xortout)
			}
			if xorab.Refs() == 0 {
				clearWireForReuse(xorab)
			}
			if xorac.Refs() == 0 {
				clearWireForReuse(xorac)
			}
			if and1.Refs() == 0 {
				clearWireForReuse(and1)
			}
		}

		t = invertWireNoInvertOutput(remainw[length-1])

		if !IsModDiv {
			ldest[(length-1)-i] = t
		}
	}

	/*get modulus*/
	if IsModDiv {
		for i := 0; i < length; i++ {
			ldest[i] = remainw[i]
		}
		addDest := pool.GetWires(typ.Num(len(ldest)))
		outputAddition(ldest, lright, addDest)

		for i := 0; i < len(ldest); i++ {
			assignWireCond(ldest[i], addDest[i], ldest[len(destv)])
		}
		pool.FreeSet(addDest)
	}

	/* Reduce to original length */
	for i := 0; i < origlength; i++ {
		assignWire(destv[i], ldest[i])
	}
	pool.FreeSinglesIfNoRefs()
}

// outputDivideSigned computes the division or modulus of leftv and rightv when
// the two are signed values, producing gates when it is necessary.
// leftv - dividend, rightv - divisor
// IsModDiv value is true when we want the modulus
// If we simply want to divide, IsModDiv equals false
func outputDivideSigned(leftv, rightv, destv wr.WireSet, IsModDiv bool) {
	origlength := typ.Num(len(leftv))
	length := origlength + 1

	carry := pool.GetWire()
	xorab := pool.GetWire()
	xorac := pool.GetWire()
	and1 := pool.GetWire()
	xortout := pool.GetWire()

	t := W_1

	lleft := wr.EmptyWireSet(length)
	lright := wr.EmptyWireSet(length)
	ldest := wr.EmptyWireSet(length)

	for i := typ.Num(0); i < origlength; i++ {
		lleft[i] = pool.GetWire()
		assignWire(lleft[i], leftv[i])
		lright[i] = pool.GetWire()
		assignWire(lright[i], rightv[i])
	}
	/*extend extra bit for correctness purposes*/
	lleft[length-1] = W_0
	lright[length-1] = W_0

	ifsubtractl := pool.GetWire()
	ifsubtractr := pool.GetWire()
	assignWire(ifsubtractl, lleft[origlength-1])
	assignWire(ifsubtractr, lright[origlength-1])

	zeros := wr.EmptyWireSet(origlength)
	subDestr := pool.GetWires(origlength)
	subDestl := pool.GetWires(origlength)

	for i := typ.Num(0); i < origlength; i++ {
		zeros[i] = W_0
	}
	outputSubtract(zeros, leftv, subDestl)
	outputSubtract(zeros, rightv, subDestr)
	for i := typ.Num(0); i < origlength; i++ {
		assignWireCond(lleft[i], subDestl[i], ifsubtractl)
		assignWireCond(lright[i], subDestr[i], ifsubtractr)
	}

	inputx := wr.EmptyWireSet(length)
	inputy := wr.EmptyWireSet(length)
	remainw := wr.EmptyWireSet(length)

	// divisor
	copy(inputx, lright)
	inputy[0] = lleft[length-1]
	for i := typ.Num(1); i < length; i++ {
		inputy[i] = pool.GetWire()
	}

	keepwiresA := wr.EmptyWireSet(0)
	keepwiresB := wr.EmptyWireSet(0)

	for i := typ.Num(0); i < length; i++ {
		// setting each input row
		if i == 0 {
			carry.State = wr.ONE
		} else {
			inputy[0] = lleft[length-1-i]
			for j := typ.Num(1); j < length; j++ {
				assignWire(inputy[j], remainw[j-1])
			}
			assignWire(carry, t)
		}

		// controlled add / subtract
		for j := typ.Num(0); j < length; j++ {
			xortout = clearWireForReuse(xortout)
			outputGateNoInvertOutputToDest(6, t, inputx[j], xortout)

			xorab = clearWireForReuse(xorab)
			outputGateNoInvertOutputToDest(6, inputy[j], xortout, xorab)

			//full adder part
			if remainw[j] != nil {
				//save wires from layer i so they can be used at layer i+1 (otherwise each remainw wire requires a new wire, very inefficient)
				if remainw[j].Refs() > 0 {
					if i%2 == 0 {
						keepwiresA = append(keepwiresA, remainw[j])
						if len(keepwiresB) > 0 {
							remainw[j] = keepwiresB.PopBack()
						} else {
							remainw[j] = pool.GetWire()
						}
					} else {
						keepwiresB = append(keepwiresB, remainw[j])
						if len(keepwiresA) > 0 {
							remainw[j] = keepwiresA.PopBack()
						} else {
							remainw[j] = pool.GetWire()
						}
					}
					if remainw[j] == nil {
						remainw[j] = pool.GetWire()
					}
				}

				remainw[j] = clearWireForReuse(remainw[j])
				outputGateToDest(6, xorab, carry, remainw[j])
			} else {
				remainw[j] = outputGate(6, xorab, carry)
			}

			if j < length-1 {
				xorac = clearWireForReuse(xorac)
				outputGateToDest(6, carry, xortout, xorac)

				and1 = clearWireForReuse(and1)
				outputGateNoInvertOutputToDest(8, xorab, xorac, and1)

				carry = clearWireForReuse(carry)
				outputGateNoInvertOutputToDest(6, xortout, and1, carry)
			}

			if xortout.Refs() == 0 {
				clearWireForReuse(xortout)
			}
			if xorab.Refs() == 0 {
				clearWireForReuse(xorab)
			}
			if xorac.Refs() == 0 {
				clearWireForReuse(xorac)
			}
			if and1.Refs() == 0 {
				clearWireForReuse(and1)
			}
		}

		t = invertWireNoInvertOutput(remainw[length-1])

		if !IsModDiv {
			ldest[length-1-i] = t
		}
	}

	/*get modulus*/
	if IsModDiv {
		copy(ldest, remainw)
		addDest := pool.GetWires(typ.Num(len(ldest)))
		outputAddition(ldest, lright, addDest)

		for i := 0; i < len(ldest); i++ {
			assignWireCond(ldest[i], addDest[i], ldest[len(destv)])
		}

		//signed portion of modulus below:
		resultsubDest := pool.GetWires(length)
		zeros = append(zeros, W_0)
		outputSubtract(zeros, ldest, resultsubDest)

		for i := typ.Num(0); i < length; i++ {
			assignWireCond(ldest[i], resultsubDest[i], ifsubtractl)
		}
		pool.FreeSet(addDest)
		pool.FreeSet(resultsubDest)
	} else {
		resultsubDest := pool.GetWires(length)
		outputSubtract(zeros, ldest, resultsubDest)

		result := outputGateNoInvertOutput(6, ifsubtractl, ifsubtractr)

		for i := typ.Num(0); i < length; i++ {
			assignWireCond(ldest[i], resultsubDest[i], result)
		}
		pool.FreeSet(resultsubDest)
	}

	/* Reduce to original length */
	for i := typ.Num(0); i < origlength; i++ {
		assignWire(destv[i], ldest[i])
	}
	pool.FreeSet(subDestr)
	pool.FreeSet(subDestl)
	pool.FreeSinglesIfNoRefs()
}
