var $parties = 4
var $intsize = 8

var in_0 = [0,0,0]
var in_1 = [0,0,0]

var in_2 = {boolField:false, intField1:0, intField2:0}
var in_3 = {boolField:false, intField1:0, intField2:0}

var out_0 = (in_0[0] > in_1[0]) && (in_0[1] > in_1[1]) && (in_0[2] > in_1[2])
var out_1 = in_0[0] != in_1[2]
var out_2 = {boolField:false, intField1:0, intField2:0}

out_2.boolField = in_2.boolField || in_3.boolField
out_2.intField1 = min(in_2.intField1, in_3.intField1)
out_2.intField2 = max(in_2.intField2, in_3.intField2)

function min (a, b) {
	if (b < a) {
		a = b
	}
	return a
}

function max (a, b) {
	if (b > a) {
		a = b
	}
	return a
}