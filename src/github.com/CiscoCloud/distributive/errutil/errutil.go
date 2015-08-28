package errutil

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"reflect"
	"strings"
)

// ParameterLengthError is the type of error a Check raises when it recieves the
// incorrect number of parameters.
type ParameterLengthError struct {
	Expected int
	Params   []string
}

func (e ParameterLengthError) Error() string {
	msg := "Expected " + fmt.Sprint(e.Expected) + " parameters, got "
	msg += fmt.Sprint(len(e.Params)) + ". They were " + fmt.Sprint(e.Params)
	return msg
}

// ParameterTypeError is the type of error a Check returns when it can't parse
// its parameters as their expected types.
type ParameterTypeError struct{ Parameter, Expected string }

func (e ParameterTypeError) Error() string {
	return "Expected parameter of type " + e.Expected + ", got " + e.Parameter
}

// FileError is an abstraction of CouldntReadError and CouldntWriteError
func PathError(path string, err error, action string) {
	if err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Fatal("Couldn't " + action + " file/dir")
	}
}

// Success is what a check should return if it is successful
func Success() (int, string, error) { return 0, "", nil }

// CouldntWriteError logs.Fatal an error relating to writing a file
func CouldntWriteError(path string, err error) { PathError(path, err, "write") }

// CouldntReadError logs.Fatal an error related to reading a file
func CouldntReadError(path string, err error) { PathError(path, err, "read") }

// GenericError is a general error where the requested variable was not found in
// a given list of variables. This is pure DRY.
func GenericError(msg string, specified interface{}, actual interface{}) (int, string, error) {
	ReflectError(actual, reflect.Slice, "GenericError")

	threshold := 50
	actualStrSlc := []string{}
	for i := 0; i < reflect.ValueOf(actual).Len() && i < threshold; i++ {
		valueString := fmt.Sprint(reflect.ValueOf(actual).Index(i))
		actualStrSlc = append(actualStrSlc, valueString)
	}
	actualStr := strings.Join(actualStrSlc, ", ")
	msg += ":\n\tSpecified: " + fmt.Sprint(specified)
	msg += "\n\tActual: " + actualStr
	return 1, msg, nil
}

// ExecError logs.Fatal with a useful message for errors that occur when
// using os/exec to run commands
func ExecError(cmd *exec.Cmd, out string, err error) {
	if err != nil {
		msg := "Failed to execute command"
		if strings.Contains(out, "permission denied") {
			msg = "Permission denied when running command"
		} else if strings.Contains(err.Error(), "not found in $PATH") {
			msg = "Couldn't find executable when running command"
		}
		log.WithFields(log.Fields{
			"command": cmd.Args,
			"path":    cmd.Path,
			"output":  out,
			"error":   err.Error(),
		}).Fatal(msg)
	}
}

// IndexError logs a message about an attempt to access an array element outside
// the range of a given list
func IndexError(msg string, i int, slc interface{}) {
	ReflectError(slc, reflect.Slice, "IndexError")
	length := reflect.ValueOf(slc).Len()
	if i >= length || i < 0 {
		log.WithFields(log.Fields{
			"index":  i,
			"slice":  fmt.Sprint(slc),
			"length": length,
		}).Fatal("IndexError: " + msg)
	}
}

// ReflectError logs a message about a failure of types during reflection
func ReflectError(value interface{}, expectedKind reflect.Kind, funcName string) {
	kind := reflect.ValueOf(value).Kind()
	if kind != expectedKind {
		log.WithFields(log.Fields{
			"expected": expectedKind,
			"actual":   kind,
			"value":    value,
		}).Fatal("ReflectError: Value didn't have expected kind in " + funcName)
	}
}
