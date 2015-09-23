package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"testing"
)

var smallInts = [][]string{{"0"}, {"1"}, {"2"}}

var bigIntsUnder100 = [][]string{{"100"}, {"99"}, {"98"}}

var reallyBigInts = [][]string{
	{"999999999999999999"}, {"888888888888888888"}, {"777777777777777777"},
}

// Some of these will fail if the resource usage is below 3%, above 98%, etc.

// this takes the place of individual tests for getSwap, getMemory, etc.
/*
func TestGetSwapOrMemory(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"total", "used", "free"} {
		for _, swapOrMem := range []string{"swap", "memory"} {
			for _, units := range []string{"b", "kb", "mb", "gb", "tb"} {
				result := getSwapOrMemory(status, swapOrMem, units)
				if result < 0 {
					msg := "getSwapOrMemory gave a negative result"
					msg += "\n\tStatus: " + status
					msg += "\n\tSwapOrMemory: " + swapOrMem
					msg += "\n\tUnits: " + units
					msg += "\n\tResult: " + fmt.Sprint(result)
					t.Error(msg)
				}
			}
		}
	}
}
*/

func TestGetUsedPercent(t *testing.T) {
	t.Parallel()
	for _, swapOrMem := range []string{"swap", "memory"} {
		used := getUsedPercent(swapOrMem)
		if used < 0 {
			msg := "getUsedPercent reported negative " + swapOrMem + " usage"
			msg += "\n\tUsed: " + fmt.Sprint(used)
		}
	}
}

func TestMemoryUsage(t *testing.T) {
	t.Parallel()
	validInputs := append(smallInts, bigIntsUnder100...)
	invalidInputs := append(append(reallyBigInts, notInts...), negativeInts...)
	testParameters(validInputs, invalidInputs, MemoryUsage{}, t)
	testCheck(bigIntsUnder100, smallInts, MemoryUsage{}, t)
}

func TestSwapUsage(t *testing.T) {
	t.Parallel()
	validInputs := append(smallInts, bigIntsUnder100...)
	invalidInputs := append(append(notLengthOne, notInts...), negativeInts...)
	testParameters(validInputs, invalidInputs, SwapUsage{}, t)
	testCheck(bigIntsUnder100, [][]string{}, SwapUsage{}, t)
}

func testFreeMemoryOrSwap(t *testing.T, chk chkutil.Check) {
	bWinners := suffixParameter(smallInts, "b")
	kbWinners := suffixParameter(smallInts, "kb")
	mbWinners := suffixParameter(smallInts, "mb")
	mbLosers := suffixParameter(reallyBigInts, "mb")
	gbLosers := suffixParameter(reallyBigInts, "gb")
	tbLosers := suffixParameter(reallyBigInts, "tb")
	goodEggs := append(append(bWinners, kbWinners...), mbWinners...)
	badEggs := append(append(mbLosers, gbLosers...), tbLosers...)

	validInputs := append(goodEggs, badEggs...)
	invalidInputs := append(names, notInts...)

	testParameters(validInputs, invalidInputs, chk, t)
	testCheck(goodEggs, badEggs, chk, t)
}

func TestFreeMemory(t *testing.T) {
	t.Parallel()
	testFreeMemoryOrSwap(t, FreeMemory{})
}

func TestFreeSwap(t *testing.T) {
	t.Parallel()
	testFreeMemoryOrSwap(t, FreeSwap{})
}

// $1 - path, $2 maxpercent
func TestDiskUsage(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(dirParameters, "95")
	invalidInputs := append(notLengthTwo,
		[][]string{{"", ""}, {}, {"/", "garble"}}...,
	)
	goodEggs := [][]string{[]string{"/", "99"}}
	badEggs := [][]string{[]string{"/", "1"}}
	testParameters(validInputs, invalidInputs, DiskUsage{}, t)
	testCheck(goodEggs, badEggs, DiskUsage{}, t)
}
