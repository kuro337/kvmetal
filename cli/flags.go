package cli

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kvmgo/configuration/presets"
	"kvmgo/utils"
	"kvmgo/vm"
	kvm "kvmgo/vm"
)

/*
-- Presets

-- Kubernetes Control Plane + Worker

	go run main.go --launch-vm=control  --preset=kubecontrol --mem=4096 --cpu=2
	go run main.go --launch-vm=worker   --preset=kubeworker  --mem=4096 --cpu=2

-- Full Cluster

	go run main.go --cluster --control=kubecontrol --workers=kubeworker1,kubeworker2

-- Data Streaming, Processing

	go run main.go --launch-vm=spark      --preset=spark 		  --mem=8192 --cpu=4
	go run main.go --launch-vm=opensearch --preset=opensearch --mem=8192 --cpu=4
	go run main.go --launch-vm=kafka      --preset=kafka      --mem=8192 --cpu=4
	go run main.go --launch-vm=hadoop     --preset=hadoop     --mem=8192 --cpu=4

-- Expose a VM
go run main.go --expose-vm=hadoop --port=8081 --hostport=8003 --external-ip=192.168.1.224 --protocol=tcp

-- Clean up running VMs

	go run main.go --cleanup=kubecontrol,kubeworker
	go run main.go --cleanup=spark
	go run main.go --cleanup=hadoop

-- Launching VM with Userdata Defined

	go run main.go --launch-vm=kubecontrol --mem=4086 --cpu=4 --userdata=data/userdata/kube/control.txt

-- Launch a new VM

	go run main.go --launch-vm=spark --mem=24576 --cpu=8

# VM with zsh

	go run main.go --launch-vm=hadoop --mem=8192 --cpu=4 --userdata=data/userdata/shell/user-data.txt

# launches hadoop

	go run main.go --launch-vm=hadoop --mem=8192 --cpu=4 --userdata=data/userdata/shell/user-data.txt

To get detailed info about the VM

	virsh dominfo spark
*/
func Evaluate() {
	config := ParseFlags()
	if config.Help == true {
		utils.MockANSIPrint()
	}
	switch config.Action {
	case Launch:
		launchCluster(config.Control, config.Workers)
	case Cleanup:
		cleanupNodes(config.Cleanup)
	case Running:
		utils.ListVMs(2, true)
	case New:
		launchVM(*config)
	default:
		log.Println("No action specified or recognized.")
	}
}

type Action int

const (
	Unknown Action = iota
	Launch
	Cleanup
	Metadata
	Configure
	New
	Running
)

type Config struct {
	VM           string
	Action       Action
	Cluster      bool
	Cleanup      []string
	Control      string
	Workers      []string
	Name         string
	Memory       int
	CPU          int
	BootScript   string
	Userdata     string
	UserdataFile string
	Help         bool
	Preset       string
}

func ParseFlags() *Config {
	var action Action

	cluster := flag.Bool("cluster", false, "Launch a cluster with control and worker nodes")
	cleanup := flag.String("cleanup", "", "Cleanup nodes by name, comma-separated")
	control := flag.String("control", "", "Name of the control node")
	workers := flag.String("workers", "", "Names of the worker nodes, comma-separated")
	running := flag.Bool("running", false, "View virtual machines running")
	launch_vm := flag.String("launch-vm", "", "Launch a new VM with the specified name")
	memory := flag.String("mem", "", "Specify Memory for the VM")
	cpu := flag.String("cpu", "", "Specify Cores for the VM")
	bootScript := flag.String("boot", "", "Path to the custom boot script")
	userdata := flag.String("userdata", "", "Path to the User Data Cloud init script to be used Directly")
	preset := flag.String("preset", "", "Choose from a preconfigured Setup such as Hadoop, Spark, Kubernetes")
	help := flag.Bool("help", false, "View Help for kVM application")
	vm := flag.String("vm", "", "Virtual Machine (Domain Name)")
	exposeVM := flag.String("expose-vm", "", "Name of the VM to expose ports for")
	externalIP := flag.String("external-ip", "0.0.0.0", "External IP to map the port to, defaults to 0.0.0.0")
	hostPort := flag.Int("hostport", 0, "Host port to map to the VM port")
	vmPort := flag.Int("port", 0, "VM port to be exposed")
	protocol := flag.String("protocol", "tcp", "Protocol for the port mapping, defaults to tcp")

	flag.Parse()

	utils.LogWhiteBlueBold(fmt.Sprintf("VM passed: %s", *vm))

	if *exposeVM != "" && *hostPort != 0 && *vmPort != 0 {
		netConfig := ParseNetExposeFlags(*exposeVM, *vmPort, *hostPort, *externalIP, *protocol)
		if netConfig != nil {
			CreateAndSetNetExposeConfig(*netConfig)
		}
	}

	if *cluster {
		action = Launch
	} else if *cleanup != "" {
		action = Cleanup
	} else if *running {
		action = Running
	} else if *launch_vm != "" {
		action = New
	}

	config := &Config{
		Action:  action,
		VM:      *vm,
		Name:    *launch_vm,
		Cluster: *cluster,
		Control: *control,
		Workers: strings.Split(*workers, ","),
		Help:    *help,
	}

	if *memory != "" {
		parsedMem, err := strconv.Atoi(*memory)
		if err != nil {
			log.Printf("Failed to parse memory value: %v.Setting default memory as 2048mb", err)
		}
		config.Memory = parsedMem // config.Memory , log color , and Config linen 228
	}

	if *cpu != "" {
		parsedCpu, err := strconv.Atoi(*cpu)
		if err != nil {
			log.Fatalf("Failed to parse CPU value: %v. Setting to default as 2", err)
		}
		config.CPU = parsedCpu

	}

	if *preset != "" {
		switch *preset {
		case "hadoop":
			utils.LogRichLightPurple("Preset: Hadoop")
			config.Userdata = presets.CreateHadoopUserData("ubuntu", "password", *launch_vm)
		case "kubecontrol":
			utils.LogRichLightPurple("Preset: Kube Control Plane")
			config.Userdata = presets.CreateKubeControlPlaneUserData("ubuntu", "password", *launch_vm)
		case "kubeworker":
			utils.LogRichLightPurple("Preset: Kube Worker Node")
			config.Userdata = presets.CreateKubeWorkerUserData("ubuntu", "password", *launch_vm)
		default:
			utils.LogError("Invalid Preset Passed")
		}
	}

	if *userdata != "" {

		absUserdataPath, err := filepath.Abs(*userdata)
		if err != nil {
			log.Printf("Path could not be resolved. Make sure --userdata path is valid.")
			absUserdataPath = ""
		}
		config.UserdataFile = absUserdataPath
	}

	if *bootScript != "" {

		absBootScriptPath, err := filepath.Abs(*bootScript)
		if err != nil {
			log.Printf("Path could not be resolved. Make sure --boot path is valid.")
			absBootScriptPath = ""
		}
		config.BootScript = absBootScriptPath
	}

	if *cleanup != "" {
		config.Cleanup = strings.Split(*cleanup, ",")
	}

	return config
}

// launchVM launches a new Virtual Machine
func launchVM(launchConfig Config) {
	log.Printf("Launching new VM: %s\n", launchConfig.Name)

	vmConfig := CreateVMConfig(launchConfig)

	vm.LaunchNewVM(vmConfig)
}

func launchCluster(controlNode string, workerNodes []string) {
	fmt.Printf("Launching control node: %s\n", controlNode)
	for _, worker := range workerNodes {
		if worker != "" {
			fmt.Printf("NOTE:PLACEHOLDER. Actually Launches 1 Control + 1 Worker.Launching worker node: %s with control node: %s\n", worker, controlNode)
		}
	}
	err := vm.LaunchCluster(controlNode, "worker")
	if err != nil {
		log.Printf(utils.TurnError("Failed to Launch k8 Cluster"))
	}
}

func CreateVMConfig(config Config) *vm.VMConfig {
	if config.UserdataFile != "" && config.Userdata != "" {
		utils.LogWarning("Both User Data and --preset cannot be used. --preset overrides.")
	}

	return vm.NewVMConfig(config.Name).
		SetImageURL("https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img").
		SetImagesDir("data/images").
		SetUserData(config.UserdataFile).
		SetCores(config.CPU).
		SetMemory(config.Memory).
		SetCloudInitDataInline(config.Userdata)
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

				err := kvm.RemoveVMCompletely(vmName)
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
