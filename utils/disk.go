package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Dir %s did not exist - creating.", path)
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			return err
		}
	}
	log.Printf("Created dir %s", path)
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ImageExists(imageName, dir string) bool {
	imagePath := filepath.Join(dir, imageName)
	_, err := os.Stat(imagePath)
	return !os.IsNotExist(err)
}

func CopyScriptsAndService(bootScript, systemdScript, mountPath, vmName string) error {
	ubuntuUserPath := filepath.Join(mountPath, "home", "ubuntu")
	systemdPath := filepath.Join(mountPath, "etc", "systemd", "system")

	bootScriptDest := filepath.Join(ubuntuUserPath, filepath.Base(bootScript))
	systemdServiceDest := filepath.Join(systemdPath, vmName+".service")

	// Create directories and copy files
	cmds := []*exec.Cmd{
		exec.Command("sudo", "mkdir", "-p", ubuntuUserPath),
		exec.Command("sudo", "cp", bootScript, bootScriptDest),
		exec.Command("sudo", "chmod", "+x", bootScriptDest),
		exec.Command("sudo", "mkdir", "-p", systemdPath),
		exec.Command("sudo", "cp", systemdScript, systemdServiceDest),
		exec.Command("sudo", "chmod", "644", systemdServiceDest),
	}

	for _, cmd := range cmds {
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// PathResolvable checks if a file exists at the specified path and returns true if it does.
func PathResolvable(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func PrintCurrentPath() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Error getting current directory:", err)
		return
	}
	log.Printf("Current Path:%s", dir)
}

/*
WriteArraytoFile writes the slice of strings to the specified file path.
Usage:

	err := WriteArraytoFile(arr,"/home/user/commands.txt")
*/
func WriteArraytoFile(commands []string, filePath string) error {
	// Ensure the directory exists or create it
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for commands file: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open commands file: %w", err)
	}
	defer file.Close()

	for _, cmd := range commands {
		if _, err := file.WriteString(cmd + "\n"); err != nil {
			return fmt.Errorf("failed to write command to file: %w", err)
		}
	}

	return nil
}

func ReadFileFatal(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read qemu hooks %s", err)
	}
	return string(content)
}
