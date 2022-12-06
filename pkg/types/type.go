package types

import (
	"fmt"
)

type Num uint32

type VarType byte

const (
	VOID = iota
	BOOL
	INT
	UINT
	ARRAY
	OBJECT
	FUNCTION
)

type Type struct {
	BaseType VarType
	L        Num // describes the size in bites of an integer, or the length of an array/structure
	SubType  *Type
	List     []*Type
	Keys     []string
}

/*                Type creators                             */
/*----------------------------------------------------------*/

var VoidType *Type = &Type{BaseType: VOID}

var BoolType *Type = &Type{BaseType: BOOL, L: 1}

func NewIntType(l Num) *Type {
	return &Type{BaseType: INT, L: l}
}

func NewUIntType(l Num) *Type {
	return &Type{BaseType: UINT, L: l}
}

func NewArrayType(l Num, t *Type) *Type {
	return &Type{BaseType: ARRAY, L: l, SubType: t}
}

func NewObjType() *Type {
	return &Type{OBJECT, 0, nil, make([]*Type, 0), make([]string, 0)}
}

func NewFunctionType(t *Type) *Type {
	return &Type{BaseType: FUNCTION, L: 0, SubType: t, List: make([]*Type, 0)}
}

/*                        Generic methods                      */
/*-------------------------------------------------------------*/

// Print sends to the standard outputs a description of the type given
func (t Type) Print(indent string) {
	switch t.BaseType {
	case VOID:
		fmt.Print(indent, "Void")
	case BOOL:
		fmt.Print(indent, "bool")
	case INT:
		fmt.Print(indent, "int[", t.L, "]")
	case UINT:
		fmt.Print(indent, "uint[", t.L, "]")
	case ARRAY:
		fmt.Print(indent, "[", t.L, "] × ")
		t.SubType.Print("")
	case OBJECT:
		fmt.Print(indent, "Object: { ")
		for i, ot := range t.List {
			fmt.Print(t.Keys[i], "(")
			ot.Print("")
			fmt.Print(") ")
		}
		fmt.Print(" }")
	case FUNCTION:
		fmt.Print(indent, "Function (")
		for _, pt := range t.List {
			pt.Print("")
			fmt.Print(", ")
		}
		fmt.Print(") -> ")
		t.SubType.Print("")
	}
}

// Equals tests the equivalence of two given types
func (t *Type) Equals(t2 *Type) bool {
	if t == t2 {
		return true
	}
	if t.BaseType != t2.BaseType {
		return false
	} else {
		switch t.BaseType {
		case INT, UINT:
			return t.L == t2.L
		case ARRAY:
			return (t.L == t2.L) && t.SubType.Equals(t2.SubType)
		case OBJECT:
			if len(t.List) != len(t2.List) {
				return false
			}
			for i, et := range t.List {
				if (t.Keys[i] != t2.Keys[i]) || !et.Equals(t2.List[i]) {
					return false
				}
			}
		case FUNCTION:
			if !t.SubType.Equals(t2.SubType) || (len(t.List) != len(t2.List)) {
				return false
			}
			for i, pt := range t.List {
				if !pt.Equals(t2.List[i]) {
					return false
				}
			}
		}
	}
	return true
}

// Size returns the size in bits of variables represented by the given type
func (t Type) Size() Num {
	switch t.BaseType {
	case ARRAY:
		return t.L * t.SubType.Size()
	case OBJECT:
		var x Num = 0
		for _, v := range t.List {
			x += v.Size()
		}
		return x
	case FUNCTION:
		x := t.SubType.Size()
		for _, v := range t.List {
			x += v.Size()
		}
		return x
	}
	return t.L
}

// AddKeyType is used to add a field to object types
func (ot *Type) AddKeyType(key string, t *Type) {
	ot.List = append(ot.List, t)
	ot.Keys = append(ot.Keys, key)
}

// AddType is used to add an argument to function types
func (ft *Type) AddType(t *Type) {
	ft.List = append(ft.List, t)
}

/*            Functions to check type                       */
/*----------------------------------------------------------*/

func (t Type) IsVoid() bool {
	return t.BaseType == VOID
}

func (t Type) IsBoolType() bool {
	return t.BaseType == BOOL
}

func (t Type) IsIntType() bool {
	return t.BaseType == INT
}

func (t Type) IsUIntType() bool {
	return t.BaseType == UINT
}

func (t Type) IsArrayType() bool {
	return t.BaseType == ARRAY
}

func (t Type) IsObjType() bool {
	return t.BaseType == OBJECT
}

func (t Type) IsFunctionType() bool {
	return t.BaseType == FUNCTION
}
