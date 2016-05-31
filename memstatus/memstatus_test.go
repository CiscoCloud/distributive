package memstatus

import (
	"os/exec"
	"testing"
)

var units = []string{"b", "kb", "mb", "gb", "tb"}

func logFreeOutput(t *testing.T) {
	cmd := exec.Command("free")
	out, _ := cmd.CombinedOutput()
	t.Logf("Output of `free`: %v", string(out))
}

func TestSwapOrMemory(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"free", "used", "total"} {
		for _, swapOrMem := range []string{"swap", "memory"} {
			for _, unit := range units {
				amt, err := swapOrMemory(status, swapOrMem, unit)
				if err != nil {
					logFreeOutput(t)
					t.Errorf("swapOrMemory failed unexpectedly: %v", err)
				}
				if amt < 0 {
					logFreeOutput(t)
					t.Logf("swapOrMemory reported negative: %v", amt)
				}
			}
		}
	}
}

func TestFreeMemory(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := FreeMemory(unit)
		if err != nil {
			logFreeOutput(t)
			t.Error("FreeMemory failed unexpectedly")
		}
		if amt < 0 {
			logFreeOutput(t)
			t.Errorf("FreeMemory reported negative: %v", amt)
		}
	}
}

func TestUsedMemory(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := UsedMemory(unit)
		if err != nil {
			logFreeOutput(t)
			t.Error("UsedMemory failed unexpectedly")
		}
		if amt < 0 {
			logFreeOutput(t)
			t.Errorf("UsedMemory reported negative: %v", amt)
		}
	}
}

func TestFreeSwap(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := FreeSwap(unit)
		if err != nil {
			logFreeOutput(t)
			t.Error("FreeSwap failed unexpectedly")
		}
		if amt < 0 {
			logFreeOutput(t)
			t.Errorf("FreeSwap reported negative: %v", amt)
		}
	}
}

func TestUsedSwap(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := UsedSwap(unit)
		if err != nil {
			logFreeOutput(t)
			t.Error("UsedSwap failed unexpectedly")
		}
		if amt < 0 {
			logFreeOutput(t)
			t.Errorf("UsedSwap reported negative: %v", amt)
		}
	}
}
