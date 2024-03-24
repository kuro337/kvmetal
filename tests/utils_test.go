package tests

import (
	"os"
	"path/filepath"
	"testing"

	"kvmgo/types"
	"kvmgo/utils"
)

func TestFileInterface(t *testing.T) {
	// Setup: create a new file and directory
	basePath := "exampleDir"
	fPath, err := types.NewPath(basePath, false)
	if err != nil {
		t.Errorf("Failed to Resolve New Path to Abs ERROR:%s", err)
	}

	fPath.PrintPaths()

	absPath, err := fPath.ToAbs()
	if err != nil {
		t.Fatalf("Error converting to absolute path: %v", err)
	}

	if !absPath.Valid() {
		if err := absPath.CreateFolder(); err != nil {
			t.Fatalf("Failed to create folder: %v", err)
		}
	}

	filePath, err := types.NewPath(filepath.Join(absPath.Get(), "newFile.txt"), false)
	if err != nil {
		t.Errorf("Failed to Resolve Path ERROR:%s", err)
	}
	if err := filePath.CreateFile(); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	filePath.PrintPaths()

	// // New: demonstrate and verify deletion of file and directory
	if err := filePath.DeleteFile(); err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	if err := absPath.DeleteDir(); err != nil {
		t.Errorf("Failed to delete folder: %v", err)
	}

	// Verify deletion
	if filePath.Valid() {
		t.Error("File still exists after deletion")
	}

	if absPath.Valid() {
		t.Error("Directory still exists after deletion")
	}
}

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
// # search "someterm"
// alias search='rg --column --line-number --no-heading --color=always --smart-case "$@" | fzf --ansi --multi --reverse --preview "bat --color=always --style=numbers --line-range=:500 {}"'

// # interactive search
// alias searchi='rg --column --line-number --no-heading --color=always --smart-case | fzf --ansi --multi --reverse --preview "bat --color=always --style=numbers --line-range=:500 {}"'

// git status
// git restore tests/config_test.go tests/discovery_test.go tests/kafka_test.go tests/keygen_test.go tests/logger_test.go tests/network_test.go tests/p
