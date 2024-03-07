package qemu_hooks

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"kvmgo/utils"
)

// Execute Commands will Run the Commands Generated to React to Port Forwarding Events
func ExecuteCommands(commands []string) error {
	for _, cmdStr := range commands {
		parts := strings.Fields(cmdStr)
		if len(parts) < 2 {
			return fmt.Errorf("invalid command: %s", cmdStr)
		}
		cmd := exec.Command(parts[0], parts[1:]...)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error executing command '%s': %v", cmdStr, err)
		}
	}
	return nil
}

// DisableBridgeFiltering Disables Bridge Filtering for Port Forwarding to Work if it is activated
func DisableBridgeFiltering() error {
	log.Printf("Disabling Bridge Filtering")

	procFile := "/proc/sys/net/bridge/bridge-nf-call-iptables"
	if _, err := os.Stat(procFile); err == nil {
		// The file exists, disable bridge filtering
		err := os.WriteFile(procFile, []byte("0\n"), 0o644)
		if err != nil {
			log.Fatalf("Failed to write to %s: %v", procFile, err)
			return fmt.Errorf(utils.TurnError(fmt.Sprintf("Failed Disabling Bridge Filtering. ERROR:%s", err)))
		}
	} else if os.IsNotExist(err) {
		log.Printf("File %s does not exist, skipping", procFile)
	} else {
		return fmt.Errorf(utils.TurnError(fmt.Sprintf("Failed to check if %s exists: %v", procFile, err)))
	}
	log.Print(utils.TurnSuccess("Successfully Disabled Bridge Filtering"))
	return nil
}
