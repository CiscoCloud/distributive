package main

import (
	"io/ioutil"
	"testing"
)

var checklistsDir = "../../../../samples/"
var checklistPath = checklistsDir + "filesystem.json"
var checklistURL = "http://pastebin.com/raw.php?i=L0FhxKpG"

func testGetChecklists(t *testing.T) {
	// raise this error when you don't get the right number of checks
	lengthError := func(expected int, actual int) {
		msg := "Checklist didn't have expected length. Expected: %s Given: %s"
		t.Errorf(msg, expected, actual)
	}
	// test getting checklist from file
	chklsts := getChecklists(checklistPath, "", "", false)
	if len(chklsts) != 1 {
		lengthError(1, len(chklsts))
	}
	// test getting checklists from dir
	files, err := ioutil.ReadDir(checklistsDir)
	if err != nil {
		t.Errorf("Error reading checklist dir: %s", checklistsDir)
	}
	chklsts = getChecklists("", checklistsDir, "", false)
	if len(chklsts) != len(files) {
		lengthError(len(files), len(chklsts))
	}
	// test getting checklists from URL
	chklsts = getChecklists("", "", checklistURL, false)
	if len(chklsts) != 1 {
		lengthError(1, len(chklsts))
	}
}
