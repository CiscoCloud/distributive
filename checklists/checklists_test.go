package checklists

import (
	"testing"
)

var validChecklistPaths = []string{
	"../samples/filesystem.yml",
	"../samples/misc.yml",
	"../samples/network.yml",
	"../samples/packages.yml",
	//"../samples/systemctl.yml",
	"../samples/usage.yml",
	"../samples/users-and-groups.yml",
}

var valid1 = []byte(`{ "Name": "test1",
"Checklist" : [ { "ID" : "file", "Parameters" : ["/dev/null"] } ] }`)

var valid2 = []byte(`{ "Name": "test2",
"Checklist" : [ { "ID" : "directory", "Parameters" : ["/"] } ] }`)

var invalid1 = []byte(`{ "asdf": "test1",
"Checklist" : [ { "ID" : "", "Parameters" : ["/dev/null"] } ] }`)

var invalid2 = []byte(`{ "Name": "test2",
"aslk" : [ { "ID" : "directory", "Parameters" : ["/"] } ] }`)

var invalid3 = []byte(`{ "Name": "test2",
"Parameters" : [ { "asdf" : "directory", "Parameters" : ["/"] } ] }`)

func TestChecklistFromBytes(t *testing.T) {
	t.Parallel()
	goodChklsts := [][]byte{valid1, valid2}
	// won't work until logging and failing is properly decoupled from
	// constructing checklists
	/*
		badChklsts := [][]byte{invalid1, invalid2, invald3}
	*/
	for _, goodEgg := range goodChklsts {
		if i, err := ChecklistFromBytes(goodEgg); err != nil {
			fmtStr := "ChecklistFromBytes failed on valid input %v with error %v"
			t.Errorf(fmtStr, i, err)
		}
	}
	/*
		for _, badEgg := range badChklsts {
			if _, err := ChecklistFromBytes(badEgg); err == nil {
				t.Errorf("ChecklistFromBytes passed on:\n%s", string(badEgg))
			}
		}
	*/
}

func TestChecklistFromFile(t *testing.T) {
	t.Parallel()
	for _, path := range validChecklistPaths {
		if _, err := ChecklistFromFile(path); err != nil {
			t.Errorf("ChecklistFromFile failed on %s", path)
		}
	}
}

func TestChecklistsFromDir(t *testing.T) {
	t.Parallel()
	_, err := ChecklistsFromDir("../samples")
	if err != nil {
		t.Error("ChecklistsFromDir failed on ../samples")
	}
}

func TestChecklistFromURL(t *testing.T) {
	t.Parallel()
	// should add more
	urls := [1]string{"http://pastebin.com/raw.php?i=GKk5yZEK"}
	for _, url := range urls {
		_, err := ChecklistFromURL(url, true)
		if err != nil {
			t.Errorf("ChecklistFromURL failed on %s", url)
		}
	}
	// don't use cache, test again
	for _, url := range urls {
		_, err := ChecklistFromURL(url, false)
		if err != nil {
			t.Errorf("ChecklistFromURL failed on %s", url)
		}
	}
}

func TestMakeReport(t *testing.T) {
	t.Parallel()
	for _, path := range validChecklistPaths {
		chklst, _ := ChecklistFromFile(path)
		_, report := chklst.MakeReport()
		if len(report) < 1 {
			t.Error("Checklist had empty report!")
		}
	}
}
