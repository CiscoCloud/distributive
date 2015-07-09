package wrkutils

// constructors are registered, have their parameter length checked, and then
// are passed all of Parameters
var Workers = make(map[string]Worker)

// a dictionary with the number of parameters that each method takes
var ParameterLength = make(map[string]int)

// RegisterCheck TODO
func RegisterCheck(name string, work Worker, numParams int) {
	Workers[name] = work
	ParameterLength[name] = numParams
}
