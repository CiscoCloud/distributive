package checks

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/fsstatus"
	"github.com/CiscoCloud/distributive/memstatus"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*
#### MemoryUsage
Description: Is system memory usage below this threshold?
Parameters:
- Percent (int8 percentage): Maximum acceptable percentage memory used
Example parameters:
- 95%, 90%, 87%
*/

// TODO use a uint
type MemoryUsage struct{ maxPercentUsed uint8 }

func init() {
    chkutil.Register("MemoryUsage", func() chkutil.Check {
        return &MemoryUsage{}
    })
    chkutil.Register("SwapUsage", func() chkutil.Check {
        return &SwapUsage{}
    })
    chkutil.Register("FreeMemory", func() chkutil.Check {
        return &FreeMemory{}
    })
    chkutil.Register("FreeSwap", func() chkutil.Check {
        return &FreeSwap{}
    })
    chkutil.Register("CPUUsage", func() chkutil.Check {
        return &CPUUsage{}
    })
    chkutil.Register("DiskUsage", func() chkutil.Check {
        return &DiskUsage{}
    })
    chkutil.Register("InodeUsage", func() chkutil.Check {
        return &InodeUsage{}
    })
}

func (chk MemoryUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	per, err := strconv.ParseInt(strings.Replace(params[0], "%", "", -1), 10, 8)
	if strings.HasPrefix(params[0], "-") || err != nil {
		return chk, errutil.ParameterTypeError{params[0], "uint8"}
	}
	chk.maxPercentUsed = uint8(per)
	return chk, nil
}

func (chk MemoryUsage) Status() (int, string, error) {
	actualPercentUsed, err := memstatus.FreeMemory("percent")
	if err != nil {
		return 1, "", err
	}
	if actualPercentUsed < int(chk.maxPercentUsed) {
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

// TODO use a uint
type SwapUsage struct{ maxPercentUsed int8 }

func (chk SwapUsage) ID() string { return "SwapUsage" }

func (chk SwapUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	per, err := strconv.ParseInt(strings.Replace(params[0], "%", "", -1), 10, 8)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "positive int8"}
	}
	if per < 0 {
		return chk, errutil.ParameterTypeError{params[0], "positive int8"}
	}
	chk.maxPercentUsed = int8(per)
	return chk, nil
}

func (chk SwapUsage) Status() (int, string, error) {
	actualPercentUsed, err := memstatus.UsedSwap("percent")
	if err != nil {
		return 1, "", err
	}
	if actualPercentUsed < int(chk.maxPercentUsed) {
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
	var actualAmount int
	switch strings.ToLower(swapOrMem) {
	case "memory":
		actualAmount, err = memstatus.FreeMemory(units)
	case "swap":
		actualAmount, err = memstatus.FreeSwap(units)
	default:
		log.Fatalf("Invalid option passed to freeMemoOrSwap: %s", swapOrMem)
	}
	if err != nil {
		return 1, "", err
	} else if actualAmount > amount {
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

func (chk FreeMemory) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	_, _, err := chkutil.SeparateByteUnits(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "amount"}
	}
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

func (chk FreeSwap) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	_, _, err := chkutil.SeparateByteUnits(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "amount"}
	}
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

// TODO use a uint
type CPUUsage struct{ maxPercentUsed int8 }

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

// TODO use a uint
type DiskUsage struct {
	path           string
	maxPercentUsed int8
}

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
	// TODO: migrate to fsstatus
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

/*
#### InodeUsage
Description: Is the inode usage below this percentage?
Parameters:
- Filesystem (string): Filesystem as shown by `df -i`
- Percent (int8 percentage): Maximum acceptable percentage used
Example parameters:
- /dev/sda1, /mnt/my-disk/, tmpfs
- 95%, 90%, 87%
*/

type InodeUsage struct {
	filesystem     string
	maxPercentUsed uint8
}

func (chk InodeUsage) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	per, err := strconv.ParseUint(strings.Replace(params[1], "%", "", -1), 10, 8)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int8"}
	}
	chk.filesystem = params[0]
	chk.maxPercentUsed = uint8(per)
	return chk, nil
}

func (chk InodeUsage) Status() (int, string, error) {
	actualPercentUsed, err := fsstatus.PercentInodesUsed(chk.filesystem)
	if err != nil {
		return 1, "Unexpected error", err
	}
	if actualPercentUsed < chk.maxPercentUsed {
		return errutil.Success()
	}
	msg := "More disk space used than expected"
	slc := []string{fmt.Sprint(actualPercentUsed) + "%"}
	return errutil.GenericError(msg, fmt.Sprint(chk.maxPercentUsed)+"%", slc)
}
