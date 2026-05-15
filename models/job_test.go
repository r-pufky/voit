package models

import "testing"

func TestJobStruct(t *testing.T) {
	job := Job{
		Dir:           "/tmp/test.txt",
		Source:        "test.txt",
		SourceAbsPath: "/tmp",
		Target:        "rename.txt",
		TargetAbsPath: "/tmp",
		Width:         8,
	}
	if job.Source != "test.txt" {
		t.Errorf("Expected test.txt, got %s", job.Source)
	}
}
