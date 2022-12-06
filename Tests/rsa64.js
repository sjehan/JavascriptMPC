// RSA

var intsize_default = 16
var parties = 2

var in_0 = 0 // base
var in_1 = 0 // e
var out_0 = 1

var i = 0
var mod = 19 // in_0 ^ in_1
for(i = 0 ; i < intsize_default ; i++) {
	if(GetWire(in_1, 0)) {
		out_0 = modMul(out_0, in_0, mod)
	}
	in_1 = in_1 >> 1
	in_0 = modMul(in_0, in_0, mod)
}

function modMul (x, y, mod) {
	return (x * y) % mod
}