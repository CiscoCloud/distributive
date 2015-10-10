package memstatus

import (
	"testing"
)

var units = []string{"b", "kb", "mb", "gb", "tb"}

func TestSwapOrMemory(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"free", "used", "total"} {
		for _, swapOrMem := range []string{"swap", "memory"} {
			for _, unit := range units {
				amt, err := swapOrMemory(status, swapOrMem, unit)
				if err != nil {
					t.Error("swapOrMemory failed unexpectedly")
				}
				if amt < 0 {
					t.Error("swapOrMemory reported negative")
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
			t.Error("FreeMemory failed unexpectedly")
		}
		if amt < 0 {
			t.Error("FreeMemory reported negative")
		}
	}
}

func TestUsedMemory(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := UsedMemory(unit)
		if err != nil {
			t.Error("UsedMemory failed unexpectedly")
		}
		if amt < 0 {
			t.Error("UsedMemory reported negative")
		}
	}
}

func TestFreeSwap(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := FreeSwap(unit)
		if err != nil {
			t.Error("FreeSwap failed unexpectedly")
		}
		if amt < 0 {
			t.Error("FreeSwap reported negative")
		}
	}
}

func TestUsedSwap(t *testing.T) {
	t.Parallel()
	for _, unit := range append(units, "percent") {
		amt, err := UsedSwap(unit)
		if err != nil {
			t.Error("UsedSwap failed unexpectedly")
		}
		if amt < 0 {
			t.Error("UsedSwap reported negative")
		}
	}
}
