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
					t.Errorf("swapOrMemory failed unexpectedly: %v", err)
				}
				if amt < 0 {
					t.Logf("swapOrMemory reported negative: %v", amt)
				}
			}
		}
	}
	logFreeOutput(t)
}

func TestFreeMemory(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := FreeMemory(unit)
		if err != nil {
			t.Error("FreeMemory failed unexpectedly")
		}
		if amt < 0 {
			t.Errorf("FreeMemory reported negative: %v", amt)
		}
	}
	logFreeOutput(t)
}

func TestUsedMemory(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := UsedMemory(unit)
		if err != nil {
			t.Error("UsedMemory failed unexpectedly")
		}
		if amt < 0 {
			t.Errorf("UsedMemory reported negative: %v", amt)
		}
	}
	logFreeOutput(t)
}

func TestFreeSwap(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := FreeSwap(unit)
		if err != nil {
			t.Error("FreeSwap failed unexpectedly")
		}
		if amt < 0 {
			t.Errorf("FreeSwap reported negative: %v", amt)
		}
	}
	logFreeOutput(t)
}

func TestUsedSwap(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := UsedSwap(unit)
		if err != nil {
			t.Error("UsedSwap failed unexpectedly")
		}
		if amt < 0 {
			t.Errorf("UsedSwap reported negative: %v", amt)
		}
	}
	logFreeOutput(t)
}
