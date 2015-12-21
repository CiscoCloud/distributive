// memstatus provides functions that provide information about both RAM and swap
// on the host.
package memstatus

import (
	"errors"
	"github.com/CiscoCloud/distributive/tabular"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// swapOrMemory returns output from `free`, it is an abstraction of swap and
// memory. inputs: status: free | used | total; swapOrMem: memory | swap;
// units: b | kb | mb | gb | tb
func swapOrMemory(status string, swapOrMem string, units string) (int, error) {
	statusToColumn := map[string]int{
		"total": 1,
		"used":  2,
		"free":  3,
	}
	unitsToFlag := map[string]string{
		"b":  "--bytes",
		"kb": "--kilo",
		"mb": "--mega",
		"gb": "--giga",
		"tb": "--tera",
	}
	typeToRow := map[string]int{
		"memory": 0,
		"swap":   1,
	}
	// check to see that our keys are really in our dict
	if _, ok := statusToColumn[status]; !ok {
		return 0, errors.New("Invalid status in swapOrMemory: " + status)
	} else if _, ok := unitsToFlag[units]; !ok {
		return 0, errors.New("Invalid units in swapOrMemory: " + units)
	} else if _, ok := typeToRow[swapOrMem]; !ok {
		return 0, errors.New("Invalid option in swapOrMemory: " + swapOrMem)
	}
	// execute free and return the appropriate output
	cmd := exec.Command("free", unitsToFlag[units])
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	// TODO probabalisticsplit isn't handling this appropriately
	//table := tabular.ProbabalisticSplit(outStr)
	colSep := regexp.MustCompile(`\s+`)
	rowSep := regexp.MustCompile(`\n+`)
	table := tabular.SeparateString(rowSep, colSep, string(out))
	column := tabular.GetColumnByHeader(status, table)

	// filter out useless row from some versions of `free`
	for i, row := range table {
		if len(row) > 0 && strings.Contains(row[0], "-/+") {
			table = append(table[:i], table[i+1:]...)
		}
	}

	row := typeToRow[swapOrMem]
	// check for errors in output of `free`
	if column == nil || len(column) < 1 {
		errors.New("Free column was empty")
	}
	if row >= len(column) {
		errors.New("`free` didn't output enough rows")
	}
	toReturn, err := strconv.ParseInt(column[row], 10, 64)
	if err != nil {
		errors.New("Couldn't parse output of `free` as an int")
	}
	return int(toReturn), nil
}

// FreeMemory returns the amount of memory that's currently unoccupied.
// units : b, kb, mb, gb, tb, percent
func FreeMemory(units string) (int, error) {
	if strings.ToLower(units) == "percent" {
		free, err := swapOrMemory("free", "memory", "b")
		if err != nil {
			return 0, err
		}
		total, err := swapOrMemory("total", "memory", "b")
		if err != nil {
			return 0, err
		}
		return int((float32(free) / float32(total)) * 100), nil
	}
	return swapOrMemory("free", "memory", units)
}

// FreeMemory returns the amount of memory that's currently unoccupied.
// units : b, kb, mb, gb, tb, percent
func UsedMemory(units string) (int, error) {
	if strings.ToLower(units) == "percent" {
		used, err := swapOrMemory("used", "memory", "b")
		if err != nil {
			return 0, err
		}
		total, err := swapOrMemory("total", "memory", "b")
		if err != nil {
			return 0, err
		}
		return int((float32(used) / float32(total)) * 100), nil
	}
	return swapOrMemory("used", "memory", units)
}

func FreeSwap(units string) (int, error) {
	if strings.ToLower(units) == "percent" {
		free, err := swapOrMemory("free", "swap", "b")
		if err != nil {
			return 0, err
		}
		total, err := swapOrMemory("total", "swap", "b")
		if err != nil {
			return 0, err
		}
		return int((float32(free) / float32(total)) * 100), nil
	}
	return swapOrMemory("free", "swap", units)
}

func UsedSwap(units string) (int, error) {
	if strings.ToLower(units) == "percent" {
		used, err := swapOrMemory("used", "swap", "b")
		if err != nil {
			return 0, err
		}
		total, err := swapOrMemory("total", "swap", "b")
		if err != nil {
			return 0, err
		}
		return int((float32(used) / float32(total)) * 100), nil
	}
	return swapOrMemory("used", "swap", units)

}
