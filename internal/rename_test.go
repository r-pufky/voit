package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/r-pufky/voit/models"
)

// TODO - update test case for automatic collision handling.
func TestExecuteRename(t *testing.T) {
	buf := &bytes.Buffer{}
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	os.WriteFile(src, []byte("hello"), 0644)

	files := []models.File{
		{
			Source:  src,
			Matched: true,
			Target:  dst,
		},
	}

	ExecuteRename(buf, files, false, false)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Errorf("File was not renamed to %s", dst)
	}
	if _, err := os.Stat(src); err == nil {
		t.Errorf("Source file %s still exists", src)
	}
}

func TestExecuteRenameFSCollision(t *testing.T) {
	buf := &bytes.Buffer{}
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	os.WriteFile(src, []byte("new"), 0644)
	dst := filepath.Join(tempDir, "dest.txt")
	os.WriteFile(dst, []byte("original"), 0644)

	expectedDst := filepath.Join(tempDir, "dest_1.txt")

	files := []models.File{
		{
			Source:  src,
			Matched: true,
			Target:  dst,
		},
	}

	ExecuteRename(buf, files, false, false)

	if data, _ := os.ReadFile(dst); string(data) != "original" {
		t.Errorf("Original file incorrectly overwritten.")
	}

	if _, err := os.Stat(expectedDst); os.IsNotExist(err) {
		t.Errorf("File not renamed: %s.", expectedDst)
	}

	if data, _ := os.ReadFile(expectedDst); string(data) != "new" {
		t.Errorf("Renamed file content is incorrect.")
	}

	if _, err := os.Stat(src); err == nil {
		t.Errorf("Source file %s still exists.", src)
	}
}
