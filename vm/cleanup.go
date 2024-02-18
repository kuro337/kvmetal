package vm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"kvmgo/network/qemu_hooks"
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
func RemoveVMCompletely(vmName string) error {
	if err := utils.UndefineAndRemoveVM(vmName); err != nil {
		return err
	}

	err := qemu_hooks.ClearVMConfig(vmName)
	if err != nil {
		log.Printf("Error clearing VM config: %v", err)
	}

	utils.LogStep("Checking if VM is still Mounted and Cleaning Mount Paths")

	DeleteMountPathIfExist(vmName)

	log.Printf("VM '%s' and associated resources have been completely removed successfully.", vmName)

	return nil
}

/*
Cleanup unmounts the Disk associated with the VM and then Deletes the mount folder

Note: The Mount path is only used during Creation for the Host to access the VM
Once we launch the VM - we already unmount it - but an additional unmount check is here
to ensure we unmount it.
*/
func Cleanup(vmName string) error {
	isRunning, err := utils.IsVMRunning(vmName)
	if err != nil {
		log.Printf("Failed to check if VM is running - could not cleanup ERROR:%s", err)
		return err
	}

	if !isRunning {
		log.Printf("VM '%s' is shut down. Proceeding with cleanup...", vmName)

		if IsMounted(vmName) {
			if err := UnmountVM(vmName); err != nil {
				log.Printf("Failed to unmount VM '%s'. Error: %v", vmName, err)
			}
		}

		DeleteMountPathIfExist(vmName)
	} else {
		log.Printf("VM '%s' is still running. Please shut down and undefine the VM before cleanup.", vmName)
	}

	return nil
}

// IsMounted checks if the specified VM's mount path is currently mounted.
func IsMounted(vmName string) bool {
	mountPath := "/mnt/" + vmName
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get mount information. Error: %v", err)
		return false
	}

	return strings.Contains(string(output), mountPath)
}

// Attempt to unmount the VM's mount path sudo guestunmount /mnt/controlplanevm
func UnmountVM(vmName string) error {
	mountPath := "/mnt/" + vmName
	unmountCmd := exec.Command("sudo", "guestunmount", mountPath)
	if err := unmountCmd.Run(); err != nil {
		return err
	}
	log.Printf("Unmounted VM '%s' successfully.", vmName)
	return nil
}

// DeleteMountPathIfExist Checks if the Path is valid and exists and then deletes it
func DeleteMountPathIfExist(vmName string) {
	mountPath := "/mnt/" + vmName

	if _, err := os.Stat(mountPath); !os.IsNotExist(err) {
		log.Printf("Mount path '%s' exists, attempting to delete...", mountPath)
		if err := os.RemoveAll(mountPath); err != nil {
			log.Printf("Failed to delete mount path '%s'. Error: %v", mountPath, err)
		} else {
			log.Printf("Mount path '%s' successfully deleted.", mountPath)
		}
	} else {
		log.Printf("Mount path '%s' does not exist, no need to delete.", mountPath)
	}
}

func LogDeletion(path string) {
	fmt.Printf("%sDeleting: %s%s\n", BOLD, path, NC)
}
