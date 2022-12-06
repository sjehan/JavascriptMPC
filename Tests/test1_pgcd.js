var $parties = 2
var $intsize = 16

var in_0 = 0
var in_1 = 0
var out_0 = 0

if (in_0 > in_1){
	out_0 = in_1
	in_1 = in_0
	in_0 = out_0
}

// var i = 0
for(var $i = 0; $i < 10; $i++){
	out_0 = in_0
	in_0 = aux(in_0, in_1)
	in_1 = out_0
}

out_0 = in_0

/**********************************************/

function aux(a, b){
	var result = b % a
	if (result == 0){
		result = a
	}
	return result
}