package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	images_dir        = "data/images"
	modified_img_dir  = "data/images/modified"
	userdata_file     = "data/userdata/default/user-data.txt"
	userdata_img_path = "data/userdata/default/user-data.img"
	artifacts         = "data/artifacts"
)

func PullImage(url, dir string) error {
	imageName := filepath.Base(url)
	imagePath := filepath.Join(dir, imageName)

	if ImageExists(imageName, dir) {
		slog.Info("Image already exists:", "image", imageName)
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func MountImage(imagePath, mountPath string) error {
	log.Printf("imagePath:%s     mountPath:%s\n", imagePath, mountPath)

	// Log the current working directory for debugging
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
		return err
	}
	log.Printf("Current working directory: %s", cwd)

	// Assuming imagePath is just the filename or a relative path from the cwd
	absImagePath, err := filepath.Abs(imagePath)
	if err != nil {
		log.Printf("Error converting to absolute path: %v", err)
		return err
	}

	log.Printf("Creating directory: sudo mkdir -p %s", mountPath)
	if err := exec.Command("sudo", "mkdir", "-p", mountPath).Run(); err != nil {
		log.Printf("Error creating directory: %v", err)
		return err
	}

	log.Printf("Mounting image: sudo guestmount -a %s -i --rw %s", absImagePath, mountPath)
	cmd := exec.Command("sudo", "guestmount", "-a", absImagePath, "-i", "--rw", mountPath)

	// Capture standard error
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error mounting image: %v", err)
		log.Printf("guestmount stderr: %s", stderr.String())
		return err
	}

	return nil
}

func UnmountImage(mountPath string) error {
	return exec.Command("sudo", "guestunmount", mountPath).Run()
}

func CreateUserDataFile(userData, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(userData)
	return err
}

func CreateBaseImage(imageURL, vmName string) (string, error) {
	baseImageName := filepath.Base(imageURL)
	modifiedImageName := vmName + "-vm-disk.qcow2"

	// Generate the modified image
	qemuCmd := fmt.Sprintf("qemu-img create -b %s -F qcow2 -f qcow2 %s 20G", baseImageName, modifiedImageName)
	log.Printf("Running qemu-img: %s", qemuCmd)
	cmd := exec.Command("sh", "-c", qemuCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to generate modified image: %v", err)
		log.Printf("Command output:\n%s", output)
		return "", err
	}

	log.Printf("Successfully Generated Modified Image: %s", modifiedImageName)

	return modifiedImageName, nil
}

/*
Static Function to pull files from a running VM
Usage:

	// Pulls the data to data/artifacts/kubecontrol by default
	err := PullFromRunningVM("vm_name", "/home/ubuntu/init.log")

	if err!=nil {...}
*/
func PullFromRunningVM(vm_name, path string) error {
	local := filepath.Join(artifacts, vm_name)

	if err := CreateDirIfNotExist(local); err != nil {
		log.Printf("Failed 	utils.CreateDirIfNotExist(local) ERROR:%s,", err)
		return err
	}

	cmd := exec.Command("sudo", "virt-copy-out", "-d", vm_name, path, local)

	log.Printf("Running command: %s\n", cmd.String())

	cmd.Run()

	return nil
}

/*
Static Function to check for files from a running VM
Usage:

			// ls - Checks if the files exist in the VM

			vmName  := "kubecontrol"
	    filePath := "/home/ubuntu/init.log"

	    exists := FileExistsInVM(vmName, filePath)
	    if exists {
	        log.Printf("File '%s' exists in VM '%s'", filePath, vmName)
	    } else {
	        log.Printf("File '%s' does not exist in VM '%s'", filePath, vmName)
	    }

		//  CLI Usage: sudo virt-ls -d vmname /home/ubuntu/init.sh
*/
func FileExistsInVM(vmName, filePath string) bool {
	dirPath := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

	cmd := exec.Command("sudo", "virt-ls", "-d", vmName, dirPath)

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing virt-ls command: %v", err)
		return false
	}

	return strings.Contains(string(output), fileName)
}

func IsVMRunning(vmName string) (bool, error) {
	cmd := exec.Command("virsh", "list", "--all")

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, vmName) && strings.Contains(line, "running") {
			//	log.Printf("%s is running!", vmName)
			return true, nil
		}
	}
	// log.Printf("%s is not running.", vmName)

	return false, scanner.Err()
}

// RemoveVM destroys the running VM : virsh destroy <vm_name>
func RemoveVM(vmName string) error {
	destroyCmd := exec.Command("virsh", "destroy", vmName)
	if _, err := destroyCmd.Output(); err != nil {
		log.Printf("Failed to destroy VM '%s', it might not be running. Error: %v", vmName, err)
	}

	undefineCmd := exec.Command("virsh", "undefine", vmName)
	if _, err := undefineCmd.Output(); err != nil {
		log.Printf("Failed to undefine VM '%s'. Error: %v", vmName, err)
		return err
	}

	log.Printf("VM '%s' has been successfully destroyed and undefined.", vmName)
	return nil
}
