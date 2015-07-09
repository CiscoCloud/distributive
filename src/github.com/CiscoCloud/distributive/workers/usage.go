package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// getSwapOrMemory returns output from `free`, it is an abstraction of
// getSwap and getMemory. inputs: status: free | used | total
// swapOrMem: memory | swap, units: b | kb | mb | gb | tb
// TODO: support kib/gib style units, with proper transformations.
func getSwapOrMemory(status string, swapOrMem string, units string) int {
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
		log.WithFields(log.Fields{
			"status":   status,
			"expected": []string{"total", "used", "free"},
		}).Fatal("Internal error: invalid status in getSwapOrMemory")
	} else if _, ok := unitsToFlag[units]; !ok {
		log.WithFields(log.Fields{
			"units":    units,
			"expected": []string{"b", "kb", "mb", "gb", "tb"},
		}).Fatal("Internal error: invalid units in getSwapOrMemory")
	} else if _, ok := typeToRow[swapOrMem]; !ok {
		log.WithFields(log.Fields{
			"option":   swapOrMem,
			"expected": []string{"memory", "swap"},
		}).Fatal("Internal error: invalid option in getSwapOrMemory")
	}
	// execute free and return the appropriate output
	cmd := exec.Command("free", unitsToFlag[units])
	outStr := wrkutils.CommandOutput(cmd)
	table := tabular.ProbabalisticSplit(outStr)
	column := tabular.GetColumnByHeader(status, table)
	row := typeToRow[swapOrMem]
	// check for errors in output of `free`
	if column == nil {
		log.WithFields(log.Fields{
			"header": status,
			"table":  "\n" + tabular.ToString(table),
		}).Fatal("Free column was empty")
	}
	if row >= len(column) {
		log.WithFields(log.Fields{
			"output": outStr,
			"column": column,
			"row":    row,
		}).Fatal("`free` didn't output enough rows")
	}
	toReturn, err := strconv.ParseInt(column[row], 10, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"cell":   column[row],
			"error":  err.Error(),
			"output": outStr,
		}).Fatal("Couldn't parse output of `free` as an int")
	}
	return int(toReturn)
}

// getSwap returns KiB of swap with a certain status: free | used | total
// and with a certain given unit: b | kb | mb | gb | tb
func getSwap(status string, units string) int {
	return getSwapOrMemory(status, "swap", units)
}

// getMemory returns KiB of swap with a certain status: free | used | total
// and with a certain given unit: b | kb | mb | gb | tb
func getMemory(status string, units string) int {
	return getSwapOrMemory(status, "memory", units)
}

// getUsedPercent returns the % of physical memory or swap currently in use
func getUsedPercent(swapOrMem string) float32 {
	used := getSwapOrMemory("used", swapOrMem, "b")
	total := getSwapOrMemory("total", swapOrMem, "b")
	return (float32(used) / float32(total)) * 100
}

// memoryUsage checks to see whether or not the system has a memory usage
// percentage below a certain threshold
func memoryUsage(parameters []string) (exitCode int, exitMessage string) {
	maxPercentUsed := wrkutils.ParseMyInt(parameters[0])
	actualPercentUsed := getUsedPercent("memory")
	if actualPercentUsed < float32(maxPercentUsed) {
		return 0, ""
	}
	msg := "Memory usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return wrkutils.GenericError(msg, fmt.Sprint(maxPercentUsed), slc)
}

// freeMemOrSwap is an abstraction of freeMemory and freeSwap, which measures
// if the desired resource has a quantity free above the amount specified
func freeMemOrSwap(input string, swapOrMem string) (exitCode int, exitMessage string) {
	// get numbers and units
	units := wrkutils.GetByteUnits(input)
	re := regexp.MustCompile("\\d+")
	amountString := re.FindString(input)
	// report errors
	if amountString == "" {
		log.WithFields(log.Fields{
			"input":  input,
			"regexp": re.String(),
		}).Fatal("Configuration error: couldn't extract number from string")
	} else if units == "" {
		log.WithFields(log.Fields{
			"input": input,
		}).Fatal("Configuration error: couldn't extract byte units from string")
	}
	amount := wrkutils.ParseMyInt(amountString)
	actualAmount := getSwapOrMemory("free", swapOrMem, units)
	if actualAmount > amount {
		return 0, ""
	}
	msg := "Free " + swapOrMem + " lower than defined threshold"
	actualString := fmt.Sprint(actualAmount) + units
	return wrkutils.GenericError(msg, input, []string{actualString})

}

// freeMemory checks that a given amount of memory is currently free
func freeMemory(parameters []string) (exitCode int, exitMessage string) {
	return freeMemOrSwap(parameters[0], "memory")
}

// freeSwap checks that a given amount of swap is currently free
func freeSwap(parameters []string) (exitCode int, exitMessage string) {
	return freeMemOrSwap(parameters[0], "swap")
}

// memoryUsage checks to see whether or not the system has a memory usage
// percentage below a certain threshold
func swapUsage(parameters []string) (exitCode int, exitMessage string) {
	maxPercentUsed := wrkutils.ParseMyInt(parameters[0])
	actualPercentUsed := getUsedPercent("swap")
	if actualPercentUsed < float32(maxPercentUsed) {
		return 0, ""
	}
	msg := "Swap usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return wrkutils.GenericError(msg, fmt.Sprint(maxPercentUsed), slc)
}

// getCPUSample helps cpuUsage do its thing. Taken from a stackoverflow:
// http://stackoverflow.com/questions/11356330/getting-cpu-usage-with-golang
func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

// cpuUsage checks to see whether or not CPU usage is below a certain %.
func cpuUsage(parameters []string) (exitCode int, exitMessage string) {
	idle0, total0 := getCPUSample()
	time.Sleep(3 * time.Second)
	idle1, total1 := getCPUSample()
	idleTicks := float32(idle1 - idle0)
	totalTicks := float32(total1 - total0)
	actualPercentUsed := 100 * (totalTicks - idleTicks) / totalTicks
	maxPercentUsed := wrkutils.ParseMyInt(parameters[0])
	if actualPercentUsed < float32(maxPercentUsed) {
		return 0, ""
	}
	msg := "CPU usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return wrkutils.GenericError(msg, fmt.Sprint(maxPercentUsed), slc)
}
