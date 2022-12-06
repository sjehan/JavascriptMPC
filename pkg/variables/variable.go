package variables

import (
	"fmt"
	circ "ixxoprivacy/pkg/circuit"
	typ "ixxoprivacy/pkg/types"
	wr "ixxoprivacy/pkg/wires"
	"os"
	"strings"
)

type VarInterface interface {
	Size() typ.Num
	Print(indent string)
	AssignPermWires(l typ.Num) typ.Num
	FillInWires(pool *wr.WirePool)
	Lock()
	Unlock()
	GetWire(i typ.Num) *wr.Wire

	GetName() string
	GetType() *typ.Type

	IsPerm() bool
	SetPerm()
	IsConst() bool
	SetConst()
	IsInput() bool
	IsOutput() bool

	IsBool() bool
	IsInt() bool
	IsArray() bool
	IsObject() bool
	IsFunction() bool
}

type Variable struct {
	*typ.Type
	Name    string
	isperm  bool
	isconst bool
}

func NewVariable(t *typ.Type, name string) *Variable {
	return &Variable{Type: t, Name: name}
}

/*        Getters and setters                    */
/*************************************************/

func (v Variable) GetName() string {
	return v.Name
}
func (v Variable) GetType() *typ.Type {
	return v.Type
}

func (v Variable) IsPerm() bool {
	return v.isperm
}
func (v *Variable) SetPerm() {
	v.isperm = true
}

func (v Variable) IsConst() bool {
	return v.isconst
}
func (v *Variable) SetConst() {
	v.isconst = true
}

func (v Variable) IsInput() bool {
	return strings.HasPrefix(v.Name, "in_")
}
func (v Variable) IsOutput() bool {
	return strings.HasPrefix(v.Name, "out_")
}

func (v *Variable) IsBool() bool {
	return false
}
func (v *Variable) IsInt() bool {
	return false
}
func (v *Variable) IsArray() bool {
	return false
}
func (v *Variable) IsObject() bool {
	return false
}
func (v *Variable) IsFunction() bool {
	return false
}

/*                Other methods                  */
/*************************************************/

func (v *Variable) FillInWires(pool *wr.WirePool) {}

func (v *Variable) AssignPermWires(l typ.Num) typ.Num {
	return l
}

func (v Variable) Print(indent string) {
	fmt.Println(indent, "%v %v", v.GetName(), "of type:")
	v.Type.Print(indent + "\t")
}

func (v *Variable) GetWire(i typ.Num) *wr.Wire {
	return nil
}

func (v *Variable) Lock() {}

func (v *Variable) Unlock() {}

// Wirebase returns the smallest wire number of wires belonging to a variable
func Wirebase(v VarInterface) typ.Num {
	switch vv := v.(type) {
	case *RegularInt, *ExtInt, *ArrayVariable, *ObjectVariable, *BoolVariable, *FunctionVariable:
		return vv.GetWire(0).Number
	default:
		fmt.Println("Error: wirebase used with void variable", v.GetName())
	}
	return 0
}

// VarFromType creates a variable to fit a certain type, with all wires inside being zeros
func VarFromType(t *typ.Type, name string) VarInterface {
	if t == nil {
		fmt.Println("Error in VarFromType: no type given for variable ", name)
		os.Exit(64)
	}
	switch t.BaseType {
	case typ.VOID:
		return nil
	case typ.BOOL:
		return NewBoolVariable(name)
	case typ.INT, typ.UINT:
		if strings.HasPrefix(name, "$") {
			return NewExtInt(t, name, 0)
		} else {
			return NewIntVariable(t, name)
		}
	case typ.ARRAY:
		return NewArrayVariable(t, name)
	case typ.OBJECT:
		return NewObjectVariable(t, name)
	default:
		fmt.Println("Error in VarFromType: Please initialize variable ", name, " with value not obtained from a function.")
	}
	return nil
}

// CircVar returns a Var object in the format of the circuit package from a Variable
func CircVar(v VarInterface) *circ.Var {
	return &circ.Var{
		Wirebase: Wirebase(v),
		Type:     v.GetType(),
	}
}
