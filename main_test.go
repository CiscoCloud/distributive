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

func testChecklist(chklst Checklist, origin string, t *testing.T) {
	if len(chklst.Checklist) < 1 {
		t.Error("Checklist had no checks associated with it\n\t" + origin)
	} else if chklst.Name == "" {
		t.Error("Checklist had no name\n\t" + origin)
	}
}

func testChecklists(chklsts []Checklist, origin string, t *testing.T) {
	if len(chklsts) < 1 {
		t.Error("len(chklsts) < 1")
	}
	for _, chklst := range chklsts {
		testChecklist(chklst, origin, t)
	}
}

func TestChecklistFromBytes(t *testing.T) {
	//t.Parallel()
	byts, err := ioutil.ReadFile(checklistPath)
	if err != nil {
		t.Error("Couldn't read file: " + checklistPath)
	}
	chklst := checklistFromBytes(byts)
	testChecklist(chklst, checklistPath+" (bytes)", t)
}

func TestChecklistFromFile(t *testing.T) {
	//t.Parallel()
	chklst := checklistFromFile(checklistPath)
	testChecklist(chklst, checklistPath+" (file)", t)
}

func TestChecklistFromURL(t *testing.T) {
	//t.Parallel()
	chklst := checklistFromURL(checklistURL)
	testChecklist(chklst, checklistURL, t)
}

func TestChecklistsFromDir(t *testing.T) {
	//t.Parallel()
	chklsts := checklistsFromDir(checklistsDir)
	testChecklists(chklsts, checklistsDir+" (dir)", t)
}

func TestGetChecklists(t *testing.T) {
	//t.Parallel()
	chklsts := getChecklists(checklistPath, checklistsDir, checklistURL, false)
	testChecklists(chklsts, "(all)", t)
}
