package utils

import (
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

func PrintCurrentPath() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Error getting current directory:", err)
		return
	}
	log.Printf("Current Path:%s", dir)
}
