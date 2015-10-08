package checklists

import (
	"testing"
)

var validChecklistPaths = []string{
	"../samples/filesystem.json",
	"../samples/misc.json",
	"../samples/network.json",
	//"../samples/packages.json",
	"../samples/systemctl.json",
	"../samples/usage.json",
	"../samples/users-and-groups.json",
}

func TestChecklistFromBytes(t *testing.T) {
	t.Parallel()
	goodChklsts := [][]byte{[]byte(`
	{
		"Name": "test1",
		"Checklist" : [ { "ID" : "file", "Parameters" : ["/dev/null"] } ]
	}`),
		[]byte(`
	{
		"Name": "test2",
		"Checklist" : [ { "ID" : "directory", "Parameters" : ["/"] } ]
	}`),
	}
	// won't work until logging and failing is properly decoupled from
	// constructing checklists
	/*
		badChklsts := [][]byte{[]byte(`
		{
			"asdf": "test1",
			"Checklist" : [ { "ID" : "", "Parameters" : ["/dev/null"] } ]
		}`),
			[]byte(`
		{
			"Name": "test2",
			"aslk" : [ { "ID" : "directory", "Parameters" : ["/"] } ]
		}`),
			[]byte(`
		{
			"Name": "test2",
			"Parameters" : [ { "asdf" : "directory", "Parameters" : ["/"] } ]
		}`),
		}
	*/
	for _, goodEgg := range goodChklsts {
		if _, err := ChecklistFromBytes(goodEgg); err != nil {
			t.Errorf("ChecklistFromBytes failed on:\n%s", string(goodEgg))
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
	// currently failing with error message about check ID
	t.Parallel()
	// should add more
	urls := [1]string{"http://pastebin.com/raw.php?i=GKk5yZEK"}
	for _, url := range urls {
		_, err := ChecklistFromURL(url)
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
