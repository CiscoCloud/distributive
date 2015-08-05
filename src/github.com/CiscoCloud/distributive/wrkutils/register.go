package wrkutils

import "sync"

var mutex = &sync.Mutex{}

// constructors are registered, have their parameter length checked, and then
// are passed all of Parameters
var Workers = make(map[string]Worker)

// a dictionary with the number of parameters that each method takes
var ParameterLength = make(map[string]int)

// RegisterCheck
func RegisterCheck(name string, work Worker, numParams int) {
	mutex.Lock()
	Workers[name] = work
	ParameterLength[name] = numParams
	mutex.Unlock()
}
