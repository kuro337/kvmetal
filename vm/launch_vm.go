package vm

import (
	"fmt"
	"log"
	"log/slog"

	"kvmgo/network/qemu_hooks"
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
		log.Printf("Check sudo cat /var/log/libvirt/qemu/%s.log for verbose failure logs", vmConfig.VMName)
		log.Printf("Known Issue: Ensure no invalid hooks are present in /etc/libvirt/hooks/qemu")

		return nil, err

	}

	// for now create a default forwarding config

	running, _ := utils.IsVMRunning(vmConfig.VMName)
	if running == true {
		if err := qemu_hooks.GenerateDefForwardConf(vmConfig.VMName); err != nil {
			log.Printf("Failed Generating Default Forwarding Commands. ERROR:%s,", err)
		}
	}

	slog.Info("VM created successfully")

	utils.LogBold("For VM Boot Logs: Check /var/log/cloud-init-output.log to view boot logs.")

	return vmConfig, nil
}
