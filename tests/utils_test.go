package tests

import (
	"os"
	"path/filepath"
	"testing"

	"kvmgo/utils"
)

func TestCreateDirIfNotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	defer os.RemoveAll("tester")

	newDir := filepath.Join(tmpDir, "newdir")

	if err := utils.CreateDirIfNotExist(newDir); err != nil {
		t.Errorf("CreateDirIfNotExist failed: %s", err)
	}

	if err := utils.CreateDirIfNotExist("tester"); err != nil {
		t.Errorf("CreateDirIfNotExist failed: %s", err)
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Errorf("Directory was not created")
	}
}

// go test
