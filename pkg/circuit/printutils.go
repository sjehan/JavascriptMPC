package circuit

import "fmt"

// Method to print a whole circuit
func (C Circuit) Print(indent string) {
	fmt.Println("\n", indent, "----- Printing circuit -----\n")
	fmt.Println(indent, "Parties: ", C.Parties)
	fmt.Println(indent, "IntSize: ", C.IntSize)
	fmt.Println(indent, "TotalWires: ", C.TotalWires)

	fmt.Println("\n", indent, "Inputs:")
	for i, v := range C.Inputs {
		if v != nil {
			fmt.Println(indent+"\t Party", i)
			v.Print(indent + "\t\t")
			fmt.Println()
		}
	}
	fmt.Println("\n", indent, "Outputs:")
	for i, v := range C.Outputs {
		if v != nil {
			fmt.Println(indent+"\t Party", i)
			v.Print(indent + "\t\t")
			fmt.Println()
		}
	}

	fmt.Println("\n", indent, "Main list:")
	C.Function.Print(indent)

	if len(C.Funcs) > 0 {
		fmt.Println("\n", indent, "Other functions:")
		for i, f := range C.Funcs {
			fmt.Println(i)
			f.Print(indent + "\t")
		}
	}
}

// Method to print a variable
func (v Var) Print(indent string) {
	fmt.Println(indent, "Wirebase: ", v.Wirebase)
	fmt.Print(indent, "Type: ")
	v.Type.Print("")
}

// Method to print a function
func (f Function) Print(indent string) {
	fmt.Print(indent, "Function")
	if f.XORgates != 0 {
		fmt.Print(", XORgates: ", f.XORgates)
	}
	if f.NonXORgates != 0 {
		fmt.Print(", NonXORgates: ", f.NonXORgates)
	}
	fmt.Println()
	for _, com := range f.Commands {
		com.Print(indent + "\t")
	}
}

// Method to print a simple command
func (cm Command) Print(indent string) {
	switch cm.Kind {
	case GATE_0, GATE_1, GATE_2, GATE_3, GATE_4, GATE_5, GATE_6, GATE_7, GATE_8, GATE_9, GATE_10,
		GATE_11, GATE_12, GATE_13, GATE_14, GATE_15:
		fmt.Printf(indent+"GATE, %d(%d, %d) -> %d\n", cm.Gate(), cm.X, cm.Y, cm.To)
	case COPY:
		fmt.Printf(indent+"COPY, %d -> %d\n", cm.X, cm.To)
	case FUNCTION_CALL:
		if cm.Y == 0 {
			fmt.Printf(indent+"FUNCTION_CALL(%d)\n", cm.X)
		} else {
			fmt.Printf(indent+"FUNCTION_CALL(%d) × %d\n", cm.X, cm.Y)
		}
	case INPUT:
		fmt.Printf(indent+"INPUT, [%d] > %d\n", cm.X, cm.To)
	case OUTPUT:
		fmt.Printf(indent+"OUTPUT, %d -> [%d]\n", cm.X, cm.To)
	case MASS_COPY:
		fmt.Printf(indent+"COPY, (%d, %d) -> (%d, %d)\n", cm.X, cm.X+cm.Y-1, cm.To, cm.To+cm.Y-1)
	case MASS_INPUT:
		fmt.Printf(indent+"INPUT, [%d] > (%d, %d)\n", cm.X, cm.To, cm.To+cm.Y-1)
	case MASS_OUTPUT:
		fmt.Printf(indent+"OUTPUT, (%d, %d) -> [%d]\n", cm.X, cm.X+cm.Y-1, cm.To)
	case REPLICATE:
		fmt.Printf(indent+"REPLICATE, %d -> (%d, %d)\n", cm.X, cm.To, cm.To+cm.Y-1)
	}
}

// Method to print a UserInOut object with 0s and 1s
func (uio UserInOut) Print(indent string) {
	fmt.Print(indent + "( ")
	for _, b := range uio {
		if b {
			fmt.Print("1 ")
		} else {
			fmt.Print("0 ")
		}
	}
	fmt.Print(") ")
}

// Method to print a garbled table to the standart output
func (gt GarbledTable) Print(indent string) {
	gt[0].Print(indent)
	gt[1].Print(indent)
	gt[2].Print(indent)
}

// Method to print a garbled key to the standart output
func (gk GarbledKey) Print(indent string) {
	fmt.Printf(indent+"%X\n", gk)
}

// Method to print a garbled value to the standart output
func (gv GarbledValue) Print(indent string) {
	fmt.Printf(indent+"(%v, %X)\n", gv.P, gv.Key)
}

// Method to print a decoding key to the standart output
func (dk DecodingKey) Print(indent string) {
	fmt.Printf(indent+"(%v, %v)\n", dk[0], dk[1])
}

// Method to print a encoding function (set of garbling keys) to the standart output
func (e EncodingSet) Print(indent string) {
	fmt.Println(indent, "----- Printing encoding function -----")
	fmt.Printf(indent+" Secret key: %X\n", e.SecretKey)
	fmt.Println(indent, "Encoding keys for zero values:")
	for userID, inputs := range e.User {
		if len(inputs) > 0 {
			fmt.Println(indent+"\t", "User ", userID)
			for i, gv := range inputs {
				fmt.Printf(indent+"\t\t%d. ", i)
				gv.Print("")
			}
		}
	}
	fmt.Println()
}

// Method to print a decoding function (set of decoding keys) to the standart output
func (d DecodingSet) Print(indent string) {
	fmt.Println(indent, "----- Printing decoding function -----")
	for userID, keys := range d.User {
		if len(keys) > 0 {
			fmt.Println(indent+"\t", "User ", userID)
			for i, dk := range keys {
				fmt.Printf(indent+"\t\t%d. ", i)
				dk.Print("")
			}
		}
	}
	fmt.Println()
}

// Method to print a set of garbled tables to the standart output
func (ts TableSet) Print(indent string) {
	fmt.Println(indent, "----- Printing table set -----")
	for i, gt := range ts {
		fmt.Printf(indent+"%d.", i)
		gt.Print(indent + "\t")
	}
	fmt.Println()
}
