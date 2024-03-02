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
	LogLaunchInit(vmConfig.VMName, vmConfig.Memory, vmConfig.CPUCores)

	vmConfig.PullImage()

	if err := vmConfig.CreateBaseImage(); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Setup VM ERROR:%s", err))
		Cleanup(vmConfig.VMName)
		return nil, err
	}

	fmt.Printf(utils.LogSection("SETTING UP VM"))

	if err := vmConfig.SetupVM(); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Setup VM ERROR:%s", err))
		Cleanup(vmConfig.VMName)
		return nil, err

	}

	fmt.Printf(utils.LogSection("GENERATING CLOUDINIT USERDATA"))

	if err := vmConfig.GenerateCloudInitImgFromPath(vmConfig.UserData); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Generate Cloud-Init Disk ERROR:%s", err))
		Cleanup(vmConfig.VMName)
		return nil, err
	}

	fmt.Printf(utils.LogSection("LAUNCHING VM"))

	if err := vmConfig.CreateVM(); err != nil {
		utils.LogError(fmt.Sprintf("Failed to Create VM ERROR:%s", err))
		log.Printf("Check sudo cat /var/log/libvirt/qemu/%s.log for verbose failure logs", vmConfig.VMName)
		return nil, err

	}

	// for now create a default forwarding config
	if err := qemu_hooks.DomainAddForwardingConfigIfRunning(vmConfig.VMName); err != nil {
		log.Printf("Could Not Generating Default Forwarding Commands. ERROR:%s,", err)
	}

	slog.Info("VM created successfully")

	fmt.Printf(utils.TurnBold("For VM Boot Logs: Check /var/log/cloud-init-output.log to view boot logs."))

	return vmConfig, nil
}

func LogLaunchInit(vmName string, mem, cores int) {
	fmt.Println(utils.LogMainAction(fmt.Sprintf("Launching new VM %s : %d mem %d vcpu",
		vmName,
		mem,
		cores,
	)))

	utils.LogSection("CREATING BASE IMAGE")
}
