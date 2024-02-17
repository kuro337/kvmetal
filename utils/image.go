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
	"time"
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

// Unmounts VM from Host mount path after creation
func UnmountImage(mountPath string) error {
	return exec.Command("sudo", "guestunmount", mountPath).Run()
}

// Clears the Temp Mount Dir for the VM needed during Creation
func ClearMountPath(vmName string) error {
	removeCmd := exec.Command("sudo", "rm", "-rf", "/mnt/"+vmName)
	if err := removeCmd.Run(); err != nil {
		log.Printf("Failed to remove VM directory %s: %v", vmName, err)
		return err
	} else {
		log.Printf("Removed VM directory %s successfully", vmName)
	}
	return nil
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

// CreateBaseImage creates an Image for the VM from a Distro such as Ubuntu or any other one
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

// RebootVM restarts the VM. This is useful for rebooting once boot scripts are finished.
func RebootVM(vm_name, path string) error {
	// virsh reboot vmname

	cmd := exec.Command("virsh", "reboot", vm_name)

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

// RemoveVM shuts down the running VM : virsh shutdown <vm_name>
func ShutdownVM(vmName string) error {
	destroyCmd := exec.Command("virsh", "shutdown", vmName)
	if _, err := destroyCmd.Output(); err != nil {
		log.Printf("Attempting to shutdown VM '%s', it might not be running. Error: %v", vmName, err)
	}

	// Check every 3s for up to 15s if the VM has been shut down
	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		running, err := IsVMRunning(vmName)
		if err != nil {
			log.Printf("Error checking if VM '%s' is running: %v", vmName, err)
			continue
		}
		if !running {
			log.Printf("VM '%s' has been successfully shut down.", vmName)
			return nil
		}
	}

	// If the VM is still running after the waiting period, forcefully destroy it
	log.Printf("VM '%s' is still running after waiting period, attempting to destroy it forcefully...", vmName)
	destroyCmd = exec.Command("virsh", "destroy", vmName)
	if _, err := destroyCmd.Output(); err != nil {
		log.Printf("Failed to forcefully destroy VM '%s'. Error: %v", vmName, err)
		return err
	}

	log.Printf("VM '%s' has been forcefully destroyed.", vmName)
	return nil
}

func UndefineAndRemoveVM(vmName string) error {
	if err := ShutdownVM(vmName); err != nil {
		return err
	}

	log.Printf("Undefining VM '%s' and removing all storage...", vmName)
	undefineCmd := exec.Command("virsh", "undefine", vmName, "--remove-all-storage")
	if _, err := undefineCmd.Output(); err != nil {
		log.Printf("Failed to undefine VM '%s' and remove all storage. Error: %v", vmName, err)
		return err
	}

	log.Printf("VM '%s' has been successfully undefined and all storage removed.", vmName)
	return nil
}
