package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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
	outStr := chkutil.CommandOutput(cmd)
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

// getSwap returns amount of swap with a certain status: free | used | total
// and with a certain given unit: b | kb | mb | gb | tb
func getSwap(status string, units string) int {
	return getSwapOrMemory(status, "swap", units)
}

// getMemory returns amount of swap with a certain status: free | used | total
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

/*
#### MemoryUsage
Description: Is system memory usage below this threshold?
Parameters:
  - Percent (int8 percentage): Maximum acceptable percentage memory used
Example parameters:
  - 95%, 90%, 87%
*/

type MemoryUsage struct{ maxPercentUsed int8 }

func (chk MemoryUsage) ID() string { return "MemoryUsage" }

func (chk MemoryUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	per, err := strconv.ParseInt(strings.Replace(params[0], "%", "", -1), 10, 8)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "int8"}
	}
	chk.maxPercentUsed = int8(per)
	return chk, nil
}

func (chk MemoryUsage) Status() (int, string, error) {
	actualPercentUsed := getUsedPercent("memory")
	if actualPercentUsed < float32(chk.maxPercentUsed) {
		return errutil.Success()
	}
	msg := "Memory usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return errutil.GenericError(msg, fmt.Sprint(chk.maxPercentUsed), slc)
}

/*
#### SwapUsage
Description: Like MemoryUsage, but with swap
*/

type SwapUsage struct{ maxPercentUsed int8 }

func (chk SwapUsage) ID() string { return "SwapUsage" }

func (chk SwapUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	per, err := strconv.ParseInt(strings.Replace(params[0], "%", "", -1), 10, 8)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "int8"}
	}
	chk.maxPercentUsed = int8(per)
	return chk, nil
}

func (chk SwapUsage) Status() (int, string, error) {
	actualPercentUsed := getUsedPercent("swap")
	if actualPercentUsed < float32(chk.maxPercentUsed) {
		return errutil.Success()
	}
	msg := "Swap usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return errutil.GenericError(msg, fmt.Sprint(chk.maxPercentUsed), slc)
}

// freeMemOrSwap is an abstraction of FreeMemory and FreeSwap, which measures
// if the desired resource has a quantity free above the amount specified
func freeMemOrSwap(input string, swapOrMem string) (int, string, error) {
	amount, units, err := chkutil.SeparateByteUnits(input)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatal("Couldn't separate string into a scalar and units")
	}
	actualAmount := getSwapOrMemory("free", swapOrMem, units)
	if actualAmount > amount {
		return errutil.Success()
	}
	msg := "Free " + swapOrMem + " lower than defined threshold"
	actualString := fmt.Sprint(actualAmount) + units
	return errutil.GenericError(msg, input, []string{actualString})
}

/*
#### FreeMemory
Description: Is at least this amount of memory free?
Parameters:
  - Amount (string with byte unit): minimum acceptable amount of free memory
Example parameters:
  - 100mb, 1gb, 3TB, 20kib
*/

type FreeMemory struct{ amount string }

func (chk FreeMemory) ID() string { return "FreeMemory" }

func (chk FreeMemory) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	// TODO validate byte units, etc. here
	chk.amount = params[0]
	return chk, nil
}

func (chk FreeMemory) Status() (int, string, error) {
	return freeMemOrSwap(chk.amount, "memory")
}

/*
#### FreeSwap
Description: Like FreeMemory, but with swap instead.
*/

type FreeSwap struct{ amount string }

func (chk FreeSwap) ID() string { return "FreeSwap" }

func (chk FreeSwap) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	// TODO validate byte units, etc. here
	chk.amount = params[0]
	return chk, nil
}

func (chk FreeSwap) Status() (int, string, error) {
	return freeMemOrSwap(chk.amount, "swap")
}

// getCPUSample helps CPUUsage do its thing. Taken from a stackoverflow:
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

/*
#### CPUUsage
Description: Is the cpu usage below this percentage in a 3 second interval?
Parameters:
  - Percent (int8 percentage): Maximum acceptable percentage used
Example parameters:
  - 95%, 90%, 87%
*/

type CPUUsage struct{ maxPercentUsed int8 }

func (chk CPUUsage) ID() string { return "CPUUsage" }

func (chk CPUUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	per, err := strconv.ParseInt(strings.Replace(params[0], "%", "", -1), 10, 8)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "int8"}
	}
	chk.maxPercentUsed = int8(per)
	return chk, nil
}

func (chk CPUUsage) Status() (int, string, error) {
	// TODO check that parameters are in range 0 < x < 100
	cpuPercentUsed := func(sampleTime time.Duration) float32 {
		idle0, total0 := getCPUSample()
		time.Sleep(sampleTime)
		idle1, total1 := getCPUSample()
		idleTicks := float32(idle1 - idle0)
		totalTicks := float32(total1 - total0)
		return (100 * (totalTicks - idleTicks) / totalTicks)
	}
	actualPercentUsed := cpuPercentUsed(3 * time.Second)
	if actualPercentUsed < float32(chk.maxPercentUsed) {
		return errutil.Success()
	}
	msg := "CPU usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return errutil.GenericError(msg, fmt.Sprint(chk.maxPercentUsed), slc)
}

/*
#### DiskUsage
Description: Is the disk usage below this percentage?
Parameters:
  - Path (filepath): Path to the disk
  - Percent (int8 percentage): Maximum acceptable percentage used
Example parameters:
  - /dev/sda1, /mnt/my-disk/
  - 95%, 90%, 87%
*/

type DiskUsage struct {
	path           string
	maxPercentUsed int8
}

func (chk DiskUsage) ID() string { return "DiskUsage" }

func (chk DiskUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	} else if _, err := os.Stat(params[0]); err != nil {
		return chk, errutil.ParameterTypeError{params[0], "dir"}
	}
	per, err := strconv.ParseInt(strings.Replace(params[1], "%", "", -1), 10, 8)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int8"}
	}
	chk.path = params[0]
	chk.maxPercentUsed = int8(per)
	return chk, nil
}

func (chk DiskUsage) Status() (int, string, error) {
	// percentFSUsed gets the percent of the filesystem that is occupied
	percentFSUsed := func(path string) int {
		// get FS info (*nix systems only!)
		var stat syscall.Statfs_t
		syscall.Statfs(path, &stat)

		// blocks * size of block = available size
		totalBytes := stat.Blocks * uint64(stat.Bsize)
		availableBytes := stat.Bavail * uint64(stat.Bsize)
		usedBytes := totalBytes - availableBytes
		percentUsed := int((float64(usedBytes) / float64(totalBytes)) * 100)
		return percentUsed

	}
	actualPercentUsed := percentFSUsed(chk.path)
	if actualPercentUsed < int(chk.maxPercentUsed) {
		return errutil.Success()
	}
	msg := "More disk space used than expected"
	slc := []string{fmt.Sprint(actualPercentUsed) + "%"}
	return errutil.GenericError(msg, fmt.Sprint(chk.maxPercentUsed)+"%", slc)
}
