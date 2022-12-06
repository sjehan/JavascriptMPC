// 4x4 matrices multiplication

var $parties = 2
var $intsize = 8
var $size = 4

var out_0 = [[0,0,0,0],
			[0,0,0,0],
			[0,0,0,0],
			[0,0,0,0]]

var in_0 = out_0
var in_1 = out_0

for(var $i = 0; $i < $size; $i++) {
	for(var $j = 0; $j < $size; $j++) {
		for(var $k = 0; $k < $size; $k++) {
			out_0[$i][$j] = addAndmult(out_0[$i][$j], in_0[$i][$k], in_1[$k][$j])
		}
	}
}

function addAndmult (a, x, y) {
	return a + x * y
}