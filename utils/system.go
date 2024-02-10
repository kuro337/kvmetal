package utils

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ExecCmd(command string, print bool) (string, error) {
	// Splitting command into command and arguments
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	// Log the output, whether successful or not
	if print == true {
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

func NavigateToPath(path string) error {
	if err := os.Chdir(path); err != nil {

		log.Printf("Failed to change directory: %v", err)

		return err
	}

	return nil
}
