package fpath

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FPath interface {
	New(string) FPath
	FromBase(string) FPath
	FromAbs(string) FPath
	Get() string
	Abs() string
	Base() string
	ToBase() (FPath, error)
	Relative() (string, error)
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
// By default it will automatically try to convert the Path Passed to an Absolute Path.
// By passing lazy as true -> we can defer the Path and cwd Creation to a later point.
// Use Relative() to get transform the Stored Absolute Path to a Relative Path from where fn is called.
//
// Note: if the File Path cannot be converted to an Absolute Path (i.e the given path cannot be resolved
// from cwd, an error will be returned.
func NewPath(path string, lazy bool) (*FilePath, error) {
	if lazy {
		return &FilePath{basePath: path}, nil
	}
	if filepath.IsAbs(path) {
		return &FilePath{absPath: path}, nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("\nNOTE: Pass lazy=true if the Path is meant to be Resolved at a Later Point.")
		return nil, fmt.Errorf("Failed to Resolve Current Passed Path to Absolute Path. ERROR:%s", err)

	}
	return &FilePath{absPath: absPath}, nil
}

func SecurePath(path string) *FilePath {
	if path == "" || len(path) <= 1 {
		log.Fatalf("Not allowed: %s\n", path)
	}

	if path[0] == '/' {
		path = path[1:]
	}
	home := os.Getenv("HOME")
	if strings.HasSuffix(path, "usr") || strings.HasSuffix(path, "etc") {
		log.Fatalf("Not allowed: %s\n", path)
	}

	if !strings.HasPrefix(path, home[1:]) {
		log.Fatalf("Not allowed: %s. Must be in %s\n", path, home)
	}

	fpath, err := NewPath(path, false)
	if err != nil {
		log.Fatalf("Invalid Path passed:%s", err)
	}

	return fpath
}

// New method creates a new FilePath instance with the provided path.
// The path is treated as an absolute path.
func (f FilePath) New(path string) FPath {
	return &FilePath{basePath: path}
}

// New method creates a new FilePath instance with the provided path.
// The path is treated as an absolute path.
func (f *FilePath) PrintPaths() {
	if f.cwd == "" || f.absPath == "" {
		_, _ = f.Resolve()
	}

	log.Printf("Abs  Path: %s", f.absPath)
	log.Printf("Base Path: %s", f.basePath)
	log.Printf("cwd      : %s", f.cwd)
}

// Resolves and sets the Abs Path - in case it is required later.
func (f *FilePath) Resolve() (FPath, error) {
	var errstr strings.Builder
	if f.cwd == "" {
		cwd, err := os.Getwd()
		if err != nil {
			errstr.WriteString(fmt.Sprintf("cwd Failed: %s ", err.Error()))
		}
		f.cwd = cwd
	}

	if f.absPath == "" {
		if _, err := f.ToAbs(); err != nil {
			errstr.WriteString(fmt.Sprintf("ToAbs failed: %s ", err.Error()))
		}
	}

	return f, fmt.Errorf(errstr.String())
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

// If the Abs Path is set - returns the Path Relative to current wd. If Abs Path is not set - returns ""
// "" is treated as an Invalid Path by os Functions
func (f *FilePath) Relative() (string, error) {
	if f.absPath == "" {
		log.Printf("Absolute Path must be set to call Relative - as a Base Path can be unreliably resolved.")
		log.Printf("Consider calling Resolve() to map the current Base Path to an Absolute Path or using ToAbs() before using Relative.")
		return "", fmt.Errorf("Absolute Path must be set to call Relative - as a Base Path can be unreliably resolved.")
	}
	base, err := basePathfromAbs(f.absPath)
	if err != nil {
		return "", err
	}
	f.basePath = base
	return base, nil
}

// ToAbs converts a base path to an absolute path.
func (f *FilePath) ToAbs() (FPath, error) {
	if f.absPath != "" {
		log.Printf("File Path was already absolute")
		return f, nil
	}
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
	//	return fmt.Errorf("Planned Deletion File was : %s", f.Get())

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

	return fmt.Errorf("Planned Deletion Dir was : %s", f.Get())
	// return os.RemoveAll(f.Get())
}

func (f *FilePath) Navigate() error {
	return navigateToPath(f.Get())
}

func (fp FilePath) MarshalYAML() (interface{}, error) {
	return fp.Get(), nil
}

func (fp FilePath) MarshalJSON() ([]byte, error) {
	return json.Marshal(fp.Get())
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

// FileExists checks if a file or directory exists at the given path.
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // File exists
	}
	// IsNotExist validates against the err
	if os.IsNotExist(err) {
		return false, nil // File does not exist
	}
	return false, err // Other errors, such as permission issues
}

// LogCwd() logs cwd
func LogCwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
		return "", err
	}
	log.Printf("Current cwd: %s\n", cwd)
	return cwd, nil
}

func CreateDirIfNotExists(path string) error {
	if path == "" {
		return fmt.Errorf("Empty Path passed")
	}
	// Check if the poolPath directory exists and create it if it doesn't
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			fmt.Printf("Failed to create directory for storage pool: %v\n", err)
			return err
		}
	}
	return nil
}
