package workers

import (
	"fmt"
	"testing"
)

var smallInts = []parameters{
	[]string{"0"},
	[]string{"1"},
	[]string{"2"},
}

var bigIntsUnder100 = []parameters{
	[]string{"100"},
	[]string{"99"},
	[]string{"98"},
}

var reallyBigInts = []parameters{
	[]string{"999999999999999999"},
	[]string{"888888888888888888"},
	[]string{"777777777777777777"},
}

// Some of these will fail if the resource usage is below 3%, above 98%, etc.

// this takes the place of individual tests for getSwap, getMemory, etc.
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
	testInputs(t, memoryUsage, bigIntsUnder100, smallInts)
}

func TestSwapUsage(t *testing.T) {
	t.Parallel()
	testInputs(t, swapUsage, bigIntsUnder100, []parameters{})
}

func testFreeMemoryOrSwap(t *testing.T, wrk worker) {
	bWinners := suffixParameter(smallInts, "b")
	kbWinners := suffixParameter(smallInts, "kb")
	mbWinners := suffixParameter(smallInts, "mb")
	mbLosers := suffixParameter(reallyBigInts, "mb")
	gbLosers := suffixParameter(reallyBigInts, "gb")
	tbLosers := suffixParameter(reallyBigInts, "tb")
	winners := append(append(bWinners, kbWinners...), mbWinners...)
	losers := append(append(mbLosers, gbLosers...), tbLosers...)
	testInputs(t, freeMemory, winners, losers)
}

func TestFreeMemory(t *testing.T) {
	t.Parallel()
	testFreeMemoryOrSwap(t, freeMemory)
}

func TestFreeSwap(t *testing.T) {
	t.Parallel()
	testFreeMemoryOrSwap(t, freeSwap)
}
