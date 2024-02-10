package vm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"kvmgo/utils"
)

const (
	BOLD = "\033[1m"
	NC   = "\033[0m" // No Color
)

/*
Destroys the VM and all resources associated with it.

	virsh destroy <vm_name> // destroys VM


	$(pwd)/data/artifacts/vm_name/userdata // removes artifacts for VM


	sudo guestunmount /mnt/vm_name // unmounts VM mount


	sudo rm -rf /mnt/vm_name // clears VM mount data from host
*/
func FullCleanup(vmName string) error {
	utils.LogSection("CLEANING VM & MOUNTS")

	if err := utils.RemoveVM(vmName); err != nil {
		log.Printf("Failed to shut down and undefine VM %s ERROR:%s,", vmName, err)
		return err
	}

	userdataDir := filepath.Join("data", "artifacts", vmName, "userdata")

	// Clears user-data.img file created for mounted disk metadata such as user/pass access
	if err := os.RemoveAll(userdataDir); err != nil {
		log.Printf("Failed to remove userdata directory %s, error: %v", userdataDir, err)
		return err
	}

	utils.LogStep("UNMOUNTING VM")

	err := Cleanup(vmName)
	if err != nil {
		log.Printf("Could not successfully unmount Disk and delete mount data for %s", vmName)
	} else {
		log.Printf("%sVM Fully Cleaned Up%s", utils.BOLD, utils.NC)
	}

	return nil
}

// Cleanup unmounts the Disk associated with the VM and then Deletes the mount folder
func Cleanup(vmName string) error {
	is_running, err := utils.IsVMRunning(vmName)
	if err != nil {
		log.Printf("Failed to check if VM is running - could not cleanup ERROR:%s", err)
		return err
	}

	if is_running == false {
		log.Printf("VM Shut Down and Safe for Cleanup - Proceeding with Cleanup")

		unmountCmd := exec.Command("sudo", "guestunmount", "/mnt/"+vmName)
		if err := unmountCmd.Run(); err != nil {
			log.Printf("Failed to unmount VM %s: %v", vmName, err)
		} else {
			log.Printf("Unmounted VM %s successfully", vmName)
		}

		removeCmd := exec.Command("sudo", "rm", "-rf", "/mnt/"+vmName)
		if err := removeCmd.Run(); err != nil {
			log.Printf("Failed to remove VM directory %s: %v", vmName, err)
		} else {
			log.Printf("Removed VM directory %s successfully", vmName)
		}
	} else {
		log.Printf("VM %s is still running - shut down VM and undefine before cleanup", vmName)
	}
	return nil
}

func LogDeletion(path string) {
	fmt.Printf("%sDeleting: %s%s\n", BOLD, path, NC)
}
