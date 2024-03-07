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
	"kvmgo/constants"
	"kvmgo/network/qemu_hooks"
	"kvmgo/utils"
	"kvmgo/vm"

	kvm "kvmgo/vm"
)

/*
-- Presets

#cloud-config

hostname: worker
fqdn: worker.kuro.com
passwd: password
lock_passwd: false
sudo: ALL=(ALL) NOPASSWD:ALL
package-update: true
package_upgrade: true
password: password
ssh_pwauth: true
chpasswd: { expire: False }

#cloud-config

hostname: control
fqdn: control.kuro.com
passwd: password
lock_passwd: false
sudo: ALL=(ALL) NOPASSWD:ALL
package-update: true
package_upgrade: true
password: password
ssh_pwauth: true
chpasswd: { expire: False }

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

go run main.go --expose-vm=worker \
--port=8088 \
--hostport=9000 \
--external-ip=192.168.1.225 \
--protocol=tcp

go run main.go --expose-vm=kafkatest \
--port=9092 \
--hostport=9092 \
--external-ip=192.168.1.225 \
--protocol=tcp

-- Clean up running VMs

	go run main.go --cleanup=kubecontrol,kubeworker
	go run main.go --cleanup=spark
	go run main.go --cleanup=hadoop

-- Launching VM with Userdata Defined

	go run main.go --launch-vm=kubecontrol --mem=4086 --cpu=4 --userdata=data/userdata/kube/control.txt

-- Launch a new VM

	go run main.go --launch-vm=test   --mem=1024 --cpu=1
	go run main.go --launch-vm=consul   --mem=2048 --cpu=2
	go run main.go --launch-vm=postgres --mem=8192 --cpu=4

	go run main.go --launch-vm=spark --mem=24576 --cpu=8

# VM with zsh

	go run main.go --launch-vm=cilium --mem=8192 --cpu=4 --userdata=data/userdata/shell/user-data.txt

# launches hadoop

	go run main.go --launch-vm=hadoop --mem=8192 --cpu=4 --userdata=data/userdata/shell/user-data.txt

	go run main.go --disable-bridge-filtering

To get detailed info about the VM

	virsh dominfo spark


	KAFKA

go run main.go --launch-vm=kafka --preset=kafka --mem=8192 --cpu=4

go run main.go --expose-vm=kafka \
--port=9095 \
--hostport=9094 \
--external-ip=192.168.1.225 \
--protocol=tcp
*/
func Evaluate() {
	config, err := ParseFlags()
	if err != nil {
		log.Print("Parsing Failed - Exiting")
		os.Exit(1)
	}

	if config.Help {
		utils.MockANSIPrint()
	}

	switch config.Action {
	case Launch:
		launchCluster(config.Control, config.Workers)
	case Cleanup:
		cleanupNodes(config.Cleanup)
	case Running:
		_, _ = utils.ListVMs(2, true)
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
	SSH          string
	BootScript   string
	Userdata     string
	UserdataFile string
	Help         bool
	Preset       string
}

func ParseFlags() (*Config, error) {
	var action Action

	vm := flag.String("vm", "", "Virtual Machine (Domain Name)")
	cpu := flag.String("cpu", "", "Specify Cores for the VM")
	help := flag.Bool("help", false, "View Help for kVM application")
	preset := flag.String("preset", "", "Choose from a preconfigured Setup such as Hadoop, Spark, Kubernetes")
	memory := flag.String("mem", "", "Specify Memory for the VM")
	vmPort := flag.Int("port", 0, "VM port to be exposed")
	hostPort := flag.Int("hostport", 0, "Host port to map to the VM port")
	cluster := flag.Bool("cluster", false, "Launch a cluster with control and worker nodes")
	cleanup := flag.String("cleanup", "", "Cleanup nodes by name, comma-separated")
	control := flag.String("control", "", "Name of the control node")
	workers := flag.String("workers", "", "Names of the worker nodes, comma-separated")
	running := flag.Bool("running", false, "View virtual machines running")
	userdata := flag.String("userdata", "", "Path to the User Data Cloud init script to be used Directly")
	protocol := flag.String("protocol", "tcp", "Protocol for the port mapping, defaults to tcp")
	exposeVM := flag.String("expose-vm", "", "Name of the VM to expose ports for")
	launch_vm := flag.String("launch-vm", "", "Launch a new VM with the specified name")
	bootScript := flag.String("boot", "", "Path to the custom boot script")
	externalIP := flag.String("external-ip", "0.0.0.0", "External IP to map the port to, defaults to 0.0.0.0")
	DisableBridgeFiltering := flag.Bool("disable-bridge-filtering", false, "Disable bridge filtering for Port Forwarding")

	flag.Parse()

	if *exposeVM != "" && *hostPort != 0 && *vmPort != 0 {
		err := HandleVMNetworkExposure(*exposeVM, *vmPort, *hostPort, *externalIP, *protocol)
		if err != nil {
			log.Printf("Failed To Create Forwarding Config ERROR:%s,", err)
		}
	}

	if *DisableBridgeFiltering {

		err := qemu_hooks.DisableBridgeFiltering()
		if err != nil {
			log.Printf("Failed to Disable Bridge Filtering for Port Forwarding Enablement. ERROR:%s", err)
		}
		log.Printf(utils.TurnSuccess("Successfully Disabled Bridge Filtering"))
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
		VM:      *vm,
		Name:    *launch_vm,
		Action:  action,
		Cluster: *cluster,
		Control: *control,
		Workers: strings.Split(*workers, ","),
		Help:    *help,
	}

	mem, vcpu := ParseMemoryCPU(*memory, *cpu)
	config.CPU = vcpu
	config.Memory = mem
	config.SSH = utils.ReadFileFatal(constants.SshPub)

	if *preset != "" {
		config.Userdata = CreateUserdataFromPreset(*preset, config.Name, config.SSH)
	}

	if *userdata != "" {
		resolvedPath, _ := ResolvePath(*userdata, "--userdata")
		config.UserdataFile = resolvedPath
	}

	if *bootScript != "" {
		resolvedPath, _ := ResolvePath(*bootScript, "--boot")
		config.BootScript = resolvedPath
	}

	if *cleanup != "" {
		config.Cleanup = strings.Split(*cleanup, ",")
	}

	return config, nil
}

func launchVM(launchConfig Config) {
	vmConfig := CreateVMConfig(launchConfig)
	if _, err := vm.LaunchNewVM(vmConfig); err != nil {
		log.Printf("Failed vm.LaunchNewVM(vmConfig) go_err ERROR:%s,", err)
	}
}

func launchCluster(controlNode string, workerNodes []string) {
	fmt.Printf("Launching control node: %s\n", controlNode)
	for _, worker := range workerNodes {
		if worker != "" {
			log.Printf("NOTE:PLACEHOLDER. Actually Launches 1 Control + 1 Worker.Launching worker node: %s with control node: %s\n", worker, controlNode)
		}
	}
	err := vm.LaunchCluster(controlNode, "worker")
	if err != nil {
		log.Print(utils.TurnError("Failed to Launch k8 Cluster"))
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
		SetPubkey(constants.SshPub).
		SetCloudInitDataInline(config.Userdata)
}

func cleanupNodes(nodes []string) {
	vms, err := utils.ListVMs(2, false)
	if err != nil {
		log.Printf("Error listing VMs: %v\n", err)
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
			foundVMNames = append(foundVMNames, nodeName)
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

// Generates the VM according to Presets such as Kubernetes, Spark, Hadoop, and more
func CreateUserdataFromPreset(preset, launch_vm, sshpub string) string {
	log.Print(utils.TurnRichLightPurple(fmt.Sprintf("Preset: %s", preset)))
	switch preset {
	case "kafka":
		return presets.CreateKafkaUserData("ubuntu", "password", launch_vm, sshpub)
	case "hadoop":
		return presets.CreateHadoopUserData("ubuntu", "password", launch_vm, sshpub)
	case "kubecontrol":
		return presets.CreateKubeControlPlaneUserData("ubuntu", "password", launch_vm, sshpub, true)
	case "kubeworker":
		return presets.CreateKubeWorkerUserData("ubuntu", "password", sshpub, launch_vm)
	default:
		utils.LogError("Invalid Preset Passed")
		return ""
	}
}

func ParseMemoryCPU(mem, cpu string) (int, int) {
	memory := 2048
	vcpu := 2

	parsedMem, err := strconv.Atoi(mem)
	if err != nil && mem != "" {
		log.Printf("Failed to parse memory value: %v.Setting default memory as 2048mb", err)
	} else {
		memory = parsedMem
	}
	parsedCpu, err := strconv.Atoi(cpu)
	if err != nil && cpu != "" {
		log.Printf("Failed to parse CPU value: %v. Setting to default as 2", err)
	} else {
		vcpu = parsedCpu
	}

	return memory, vcpu
}

func ResolvePath(path, cliflag string) (string, error) {
	absBootScriptPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("Path could not be resolved. Make sure %s path is valid.", cliflag)
		return "", err
	}
	return absBootScriptPath, err
}
