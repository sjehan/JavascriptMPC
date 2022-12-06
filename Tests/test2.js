var $parties = 4
var $intsize = 8

var in_0 = 0
var in_1 = 0
var in_2 = 0
var in_3 = 0

var out_0 = 0
var out_1 = false
var out_2 = 0
var out_3 = 0

if(in_1 != 0){
	out_0 = RotateLeft(in_0, 3)
}

out_1 = GetWire(in_0, 2) && (in_1 > 10)

out_2 = (in_0 + in_1 + in_2) * in_3 / $parties

SetWire(out_3, 4, GetWire(in_0, 2))