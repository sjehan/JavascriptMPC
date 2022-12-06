var $intsize = 8
var $parties = 3

var in_0 = [[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0]]

var in_1 = [[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0],
			[0,0,0,0,0,0,0,0,0,0]]

var in_2 = 0

var out_0 = [[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false],
			[false,false,false,false,false,false,false,false,false,false]]

var out_1 = 0

var global = in_2 - 4

var i = 0
var j = 0
for(i = 0; i < 10; i++) {
	for(j = 0; j < 10; j++) {
		out_0[i][j] = (in_0[i][j] == in_1[i][j])
	}
}

out_1 = addAndmult(in_0[0][0], in_1[0][0])


function addAndmult (x, y) {
	return (x + y) * global
}