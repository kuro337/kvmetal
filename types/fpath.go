package types

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type FPath interface {
	New(string) FPath
	FromBase(string) FPath
	FromAbs(string) FPath
	Get() string
	Abs() string
	Base() string
	ToBase() (FPath, error)
	ToAbs() (FPath, error)
	Navigate() error
	Valid() bool
	CreateFile() error
	CreateFolder() error
	DeleteFile() error
	DeleteDir() error
}

type FilePath struct {
	absPath   string
	basePath  string
	cwd       string
	extension string
}

// NewPath creates a new FilePath instance, resolving the path to an absolute path.
func NewPath(path string) (*FilePath, error) {
	var absPath string
	var err error

	if filepath.IsAbs(path) {
		absPath = path
	} else {
		absPath, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("Failed to resolve absolute path for %s: %v", path, err)
		}
	}

	return &FilePath{absPath: absPath}, nil
}

// New method creates a new FilePath instance with the provided path.
// The path is treated as an absolute path.
func (f FilePath) New(path string) FPath {
	return &FilePath{basePath: path}
}

// Resolves and sets the Abs Path - in case it is required later.
func (f *FilePath) Resolve() (FPath, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return f, fmt.Errorf("Failed to get current working directory: %v", err)
	}

	f.cwd = cwd

	return f, nil
}

// FromBase sets the base path and returns the FilePath instance.
func (f *FilePath) FromBase(path string) FPath {
	f.basePath = path
	return f
}

// FromAbs sets the absolute path and returns the FilePath instance.
func (f *FilePath) FromAbs(path string) FPath {
	f.absPath = path
	return f
}

// Get returns the absolute path if available; otherwise, it returns the base path.
func (f *FilePath) Get() string {
	if f.absPath != "" {
		return f.absPath
	}
	return f.basePath
}

// Abs returns the absolute path.
func (f *FilePath) Abs() string {
	return f.absPath
}

// Base returns the base path.
func (f *FilePath) Base() string {
	return f.basePath
}

// ToBase converts an absolute path to a base path.
func (f *FilePath) ToBase() (FPath, error) {
	base, err := basePathfromAbs(f.absPath)
	if err != nil {
		return nil, err
	}
	f.basePath = base
	return f, nil
}

// ToAbs converts a base path to an absolute path.
func (f *FilePath) ToAbs() (FPath, error) {
	abs, err := createAbsPathFromRoot(f.basePath)
	if err != nil {
		return nil, err
	}
	f.absPath = abs
	return f, nil
}

// Valid checks if the path exists on the file system.
func (f *FilePath) Valid() bool {
	_, err := os.Stat(f.Get())
	return !os.IsNotExist(err)
}

// CreateFile creates a new file at the FilePath's current path.
// If the file already exists, it will truncate it.
func (f *FilePath) CreateFile() error {
	if file, err := os.Create(f.Get()); err != nil {
		log.Printf("Failed to create file: %v", err)
		return err
	} else {
		defer file.Close()
		log.Printf("File created successfully: %s", f.Get())
	}
	return nil
}

// CreateFolder creates a new folder at the FilePath's current path.
func (f *FilePath) CreateFolder() error {
	if err := os.MkdirAll(f.Get(), os.ModePerm); err != nil {
		log.Printf("Failed to create folder: %v", err)
		return err
	} else {
		log.Printf("Folder created successfully: %s", f.Get())
	}
	return nil
}

func (f *FilePath) DeleteFile() error {
	info, err := os.Stat(f.Get())
	if os.IsNotExist(err) {
		log.Printf("File does not exist: %s", f.Get())
		return nil // or return an error if preferred
	} else if err != nil {
		return err
	}
	if info.IsDir() {
		log.Printf("Cannot delete a directory using DeleteFile: %s", f.Get())
		return fmt.Errorf("target is a directory, not a file: %s", f.Get())
	}
	return os.Remove(f.Get())
}

// DeleteDir ensures only a directory can be deleted, not a file.
func (f *FilePath) DeleteDir() error {
	info, err := os.Stat(f.Get())
	if os.IsNotExist(err) {
		log.Printf("Directory does not exist: %s", f.Get())
		return nil // or return an error if preferred
	} else if err != nil {
		return err
	}
	if !info.IsDir() {
		log.Printf("Cannot delete a file using DeleteDir: %s", f.Get())
		return fmt.Errorf("target is not a directory: %s", f.Get())
	}
	return os.RemoveAll(f.Get())
}

func (f *FilePath) Navigate() error {
	return navigateToPath(f.Get())
}

func createAbsPathFromRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

/* Convert Abs Path to Base Path */
func basePathfromAbs(absPath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	relPath, err := filepath.Rel(cwd, absPath)
	if err != nil {
		log.Printf("Warning: Could not convert %s to relative path: %v", absPath, err)
		return absPath, nil
	}
	return relPath, nil
}

func navigateToPath(path string) error {
	if err := os.Chdir(path); err != nil {
		log.Printf("Failed to change directory: %v", err)
		return err
	}
	return nil
}
