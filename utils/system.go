package utils

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecCmd("ls -a",true) will run a command and return the full string result and print/not print
func ExecCmd(command string, print bool) (string, error) {
	// Splitting command into command and arguments
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	// Log the output, whether successful or not
	if print {
		log.Printf("%s", out.String())
	}

	if err != nil {
		return "", err
	}

	// Return the full, trimmed output
	return strings.TrimSpace(out.String()), nil
}

func EnableSystemdService(mountPath, serviceName string) error {
	cmd := exec.Command("sudo", "chroot", mountPath, "/bin/bash", "-c", "systemctl enable "+serviceName)
	return cmd.Run()
}

/*
Creates an Absolute Path - based on Path provided during Execution Relative to currdir

	relativePath := "data/images"
	absPath, err := CreateAbsPathFromRoot(relativePath)
*/
func CreateAbsPathFromRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

/* Convert Abs Path to Base Path   */
func BasePathfromAbs(absPath string) (string, error) {
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

func NavigateToPath(path string) error {
	if err := os.Chdir(path); err != nil {

		log.Printf("Failed to change directory: %v", err)

		return err
	}

	return nil
}
