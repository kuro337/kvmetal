package vm

import (
	"fmt"
	"log"
	"log/slog"

	"kvmgo/utils"
)

/* Launches a new Ubuntu VM with nothing setup */
func LaunchNewVM(vmConfig *VMConfig) (*VMConfig, error) {
	utils.LogMainAction(fmt.Sprintf("Launching new VM %s : %d mem %d vcpu", vmConfig.VMName, vmConfig.Memory, vmConfig.CPUCores))

	vmConfig.PullImage()

	utils.LogSection("CREATING BASE IMAGE")

	if err := vmConfig.CreateBaseImage(); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Setup VM ERROR:%s", err))
		Cleanup(vmConfig.VMName)
		return nil, err
	}
	log.Printf("Modified Base Image Creation Success at %s", vmConfig.ImagesDir)

	utils.LogSection("SETTING UP VM")

	if err := vmConfig.SetupVM(); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Setup VM ERROR:%s", err))
		Cleanup(vmConfig.VMName)
		return nil, err

	}

	utils.LogSection("GENERATING CLOUDINIT USERDATA")

	if err := vmConfig.GenerateCloudInitImgFromPath(vmConfig.UserData); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Generate Cloud-Init Disk ERROR:%s", err))
		Cleanup(vmConfig.VMName)
		return nil, err
	}

	log.Printf("Successfully Created Cloud-Init user-data .img file: %s", vmConfig.UserData)

	utils.LogSection("LAUNCHING VM")

	if err := vmConfig.CreateVM(); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Create VM ERROR:%s", err))
		return nil, err

	}

	utils.IsVMRunning(vmConfig.VMName)

	slog.Info("VM created successfully")

	return vmConfig, nil
}
