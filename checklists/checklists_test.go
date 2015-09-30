package checklists

import (
	"testing"
)

func checklistIsFull(chklst Checklist) {
	// TODO
}

func TestMakeReport(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestChecklistFromBytes(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestChecklistFromFile(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestChecklistFromStdin(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestChecklistFromDir(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestChecklistFromURL(t *testing.T) {
	t.Parallel()
	// TODO
}

/************************* FUZZING *************************/
/* https://github.com/dvyukov/go-fuzz */

func partiallyValid(chklst Checklist) bool {
	// are any of the fields not zero valued?
	switch {
	case len(chklst.Checks) != 0:
		return true
	case chklst.Name != "":
		return true
	case chklst.Notes != "":
		return true
	case chklst.Origin != "":
		return true
	}
	return false
}

func Fuzz(data []byte) int {
	chklst, err := ChecklistFromBytes(data)
	if err == nil || partiallyValid(chklst) {
		return 0
	} else {
		return 1
	}
}
