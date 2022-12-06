package types

import (
	"fmt"
	"os"
)

// MaxType takes two number types and returns the largest one necessary to sustain operations on them
func MaxType(t1, t2 *Type) *Type {
	if t1.IsIntType() && t2.IsIntType() {
		if t1.Size() >= t2.Size() {
			return t1
		} else {
			return t2
		}
	}
	if t1.IsUIntType() && t2.IsUIntType() {
		if t1.Size() >= t2.Size() {
			return t1
		} else {
			return t2
		}
	}
	if t1.IsUIntType() && t2.IsIntType() && t1.Size() <= t2.Size() {
		return t2
	}
	if t1.IsIntType() && t2.IsUIntType() && t1.Size() >= t2.Size() {
		return t1
	}
	fmt.Println("Error in MaxType: invalid operation")
	t1.Print("\t")
	t2.Print("\t")
	return nil
}

// CheckRecursiveObj checks if there are any recursive definitions in object types
func CheckRecursiveObj(t *Type, vec []*Type) {
	if t.IsObjType() {
		for _, t2 := range vec {
			if t == t2 {
				fmt.Println("Error in CheckRecursiveObj: recursion found")
				os.Exit(64)
			}
		}
		vec = append(vec, t)
		for _, t2 := range t.List {
			CheckRecursiveObj(t2, vec)
		}
	}
}
