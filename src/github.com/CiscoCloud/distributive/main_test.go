package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

var checklistsDir = "../../../../samples/"
var checklistPath = checklistsDir + "filesystem.json"
var checklistURL = "http://pastebin.com/raw.php?i=L0FhxKpG"

func TestMakeReport(t *testing.T) {
	//t.Parallel()
	chklst := Checklist{
		Name:     "test checklist",
		Codes:    []int{1, 1, 1, 0, 0, 0},
		Messages: []string{"msg1", "", "msg3", "", "", ""},
		Origin:   "testing",
	}
	chklst.makeReport()
	testStrs := []string{"Total: 6", "Passed: 3", "Failed: 3", "msg1", "msg3"}
	for _, testStr := range testStrs {
		if !strings.Contains(chklst.Report, testStr) {
			t.Error("Report didn't contain the string '" + testStr + "'")
		}
	}
}

/*
func TestGetWorker(t *testing.T) {
	//t.Parallel()
	testStrs := []string{"file", "up", "ip4", "userHasUID", "command", "temp"}
	for _, testStr := range testStrs {
		check := Check{
			Check: testStr,
		}
		wrk := getWorker(check)
		if wrk == nil {
			t.Error("Check had a nil function associated with it: " + testStr)
		}
	}
}
*/

func testChecklist(chklst Checklist, err error, origin string, t *testing.T) {
	if err != nil {
		t.Errorf("Error when creating checklist: %s", err.Error())
	} else if len(chklst.Checklist) < 1 {
		t.Errorf("Checklist had no checks associated with it: %s", origin)
	} else if chklst.Name == "" {
		t.Errorf("Checklist had no name: %s", origin)
	}
}

func testChecklists(chklsts []Checklist, err error, origin string, t *testing.T) {
	if len(chklsts) < 1 {
		t.Error("len(chklsts) < 1")
	}
	for _, chklst := range chklsts {
		testChecklist(chklst, err, origin, t)
	}
}

func TestChecklistFromBytes(t *testing.T) {
	//t.Parallel()
	byts, err := ioutil.ReadFile(checklistPath)
	if err != nil {
		t.Error("Couldn't read file: " + checklistPath)
	}
	chklst, err := checklistFromBytes(byts)
	testChecklist(chklst, err, checklistPath+" (bytes)", t)
}

func TestChecklistFromFile(t *testing.T) {
	//t.Parallel()
	chklst, err := checklistFromFile(checklistPath)
	testChecklist(chklst, err, checklistPath+" (file)", t)
}

func TestChecklistFromURL(t *testing.T) {
	//t.Parallel()
	chklst, err := checklistFromURL(checklistURL)
	testChecklist(chklst, err, checklistURL, t)
}

func TestChecklistsFromDir(t *testing.T) {
	//t.Parallel()
	chklsts, err := checklistsFromDir(checklistsDir)
	testChecklists(chklsts, err, checklistsDir+" (dir)", t)
}

func TestGetChecklists(t *testing.T) {
	//t.Parallel()
	chklsts := getChecklists(checklistPath, checklistsDir, checklistURL, false)
	testChecklists(chklsts, nil, "(all)", t)
}

/************************* FUZZING *************************/
/* https://github.com/dvyukov/go-fuzz */

func partiallyValid(chklst Checklist) bool {
	// are any of the fields not zero valued?
	switch {
	case chklst.Name != "":
		return true
	case chklst.Notes != "":
		return true
	case len(chklst.Codes) > 0:
		return true
	case len(chklst.Messages) > 0:
		return true
	case chklst.Origin != "":
		return true
	case chklst.Report != "":
		return true
	case chklst.Failed:
		return true
	}
	return false
}

func Fuzz(data []byte) int {
	chklst, err := checklistFromBytes(data)
	if err == nil || partiallyValid(chklst) {
		return 0
	} else {
		return 1
	}
}
