package cli

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"kvmgo/utils"
	kvm "kvmgo/vm"
)

/*
Clean up running VMs
go run main.go --cleanup=kubecontrol,kubeworker

go run main.go --cluster --control=kubecontrol --workers=kubeworker1,kubeworker2
go run main.go --cluster --control=kubecontrol
go run main.go --running

*/

type Action int

const (
	Unknown Action = iota // Default value, represents no action or unrecognized action
	Launch
	Cleanup
	Metadata
	Configure
	Running
)

type Config struct {
	Action  Action
	Cluster bool
	Cleanup []string
	Control string
	Workers []string
}

func ParseFlags() *Config {
	var action Action
	cluster := flag.Bool("cluster", false, "Launch a cluster with control and worker nodes")
	cleanup := flag.String("cleanup", "", "Cleanup nodes by name, comma-separated")
	control := flag.String("control", "", "Name of the control node")
	workers := flag.String("workers", "", "Names of the worker nodes, comma-separated")
	running := flag.Bool("running", false, "View virtual machines running")

	flag.Parse()

	if *cluster {
		action = Launch
	} else if *cleanup != "" {
		action = Cleanup
	} else if *running {
		action = Running
	} // Add more conditions as needed

	config := &Config{
		Action:  action,
		Cluster: *cluster,
		Control: *control,
		Workers: strings.Split(*workers, ","),
	}

	if *cleanup != "" {
		config.Cleanup = strings.Split(*cleanup, ",")
	}

	return config
}

func Evaluate() {
	config := ParseFlags()
	switch config.Action {
	case Launch:
		launchCluster(config.Control, config.Workers)
	case Cleanup:
		cleanupNodes(config.Cleanup)
	case Running:
		utils.ListVMs(2, true)
	default:
		log.Println("No action specified or recognized.")
	}
}

func launchCluster(controlNode string, workerNodes []string) {
	fmt.Printf("Launching control node: %s\n", controlNode)
	for _, worker := range workerNodes {
		if worker != "" {
			fmt.Printf("Launching worker node: %s with control node: %s\n", worker, controlNode)
		}
	}
}

func cleanupNodes(nodes []string) {
	vms, err := utils.ListVMs(2, false)
	if err != nil {
		fmt.Printf("Error listing VMs: %v\n", err)
		return
	}

	vmMap := make(map[string]string)
	for _, vm := range vms {
		vmMap[vm.Name] = vm.State
	}

	var foundVMNames []string

	log.Printf("Clean up Virtual Machines:")
	for _, nodeName := range nodes {
		state, exists := vmMap[nodeName]
		if nodeName != "" && exists {
			log.Printf(" %s %s (%s)\n", utils.TICK_GREEN, nodeName, state)
			foundVMNames = append(foundVMNames, nodeName) // Append only the VM name
		} else {
			log.Printf("VM not found: %s %s\n", nodeName, utils.CROSS_RED)
		}
	}

	if len(foundVMNames) > 0 {
		log.Printf("Proceed? (y/n)")
		if askForConfirmation() {
			for _, vmName := range foundVMNames {
				fmt.Printf("Cleaning up node: %s\n", vmName)

				err := kvm.FullCleanup(vmName)
				if err != nil {
					fmt.Printf("Failed to clean up VM %s: %v\n", vmName, err)
				}
			}
		} else {
			fmt.Println("Cleanup aborted.")
		}
	} else {
		fmt.Println("No valid VMs were specified for cleanup.")
	}
}

// askForConfirmation prompts the user for a yes/no answer and returns true for yes.
func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}
	response = strings.TrimSpace(response)
	return response == "y" || response == "Y"
}
