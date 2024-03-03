package vm

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"kvmgo/configuration"
	"kvmgo/network"
	"kvmgo/utils"
)

type VMConfig struct {
	VMName         string
	InlineUserdata string
	ImageURL       string
	ImagesDir      string
	BootFilesDir   string
	ScriptsDir     string
	BootScript     string
	SystemdScript  string
	UserData       string
	RootDir        string
	CPUCores       int
	Memory         int
	EnableServices []string
	Artifacts      []string

	artifactPath string
}

func NewVMConfig(vmName string) *VMConfig {
	pwd, _ := os.Getwd()

	return &VMConfig{
		VMName:       vmName,
		RootDir:      pwd,
		artifactPath: "data/artifacts",
	}
}

func NewKVM(vmName string) *VMConfig {
	config := &VMConfig{
		VMName: vmName,
	}
	config.artifactPath = "data/artifacts"
	return config
}

func (config *VMConfig) SetImageURL(url string) *VMConfig {
	config.ImageURL = url
	return config
}

func (config *VMConfig) SetCores(vcpus int) *VMConfig {
	config.CPUCores = vcpus
	return config
}

func (config *VMConfig) SetCloudInitDataInline(cloudInitUserData string) *VMConfig {
	if cloudInitUserData != "" {
		utils.LogStep("Using Dynamic Preset Config for Userdata")
		config.InlineUserdata = cloudInitUserData
	}
	return config
}

func (config *VMConfig) SetMemory(memory_mb int) *VMConfig {
	config.Memory = memory_mb
	return config
}

func (config *VMConfig) SetBootServices(services []string) *VMConfig {
	config.EnableServices = services
	return config
}

// Set Images Dir where VM's image files are created and stored data/images
func (config *VMConfig) SetImagesDir(dir string) *VMConfig {
	config.ImagesDir = dir
	return config
}

func (config *VMConfig) SetBootFilesDir(dir string) *VMConfig {
	config.BootFilesDir = dir
	return config
}

func (config *VMConfig) SetArtifacts(artifacts []string) *VMConfig {
	config.Artifacts = artifacts
	return config
}

func (config *VMConfig) SetUserData(userData string) *VMConfig {
	config.UserData = userData
	if userData == "" {
		log.Print("No User Data Passed - Setting Default CloudInit UserData")
		config.DefaultUserData()
	}
	return config
}

func (config *VMConfig) DefaultUserData() *VMConfig {
	config.UserData = "/home/kuro/Documents/Code/Go/kvmgo/data/userdata/default/user_data.txt"
	return config
}

func (s *VMConfig) PullImage() {
	err := utils.PullImage(s.ImageURL, s.ImagesDir)
	if err != nil {
		slog.Error("Failed HTTP GET", "error", err)
		os.Exit(1)
	}
}

/*
GenerateCustomUserDataImg creates the raw disk and attaches it as a secondary disk to the VM for user-data.
This enables username/pw access for the VM.

	virsh domblklist vm_name  // view attached disks
	qemu-img info user-data.img // Viewing Disk Type

To view Logs for CloudInit user data if boot script was set check

	cat /var/log/cloud-init-output.log | less
*/
func (config *VMConfig) GenerateCustomUserDataImg(bootScriptPath string) error {
	// Create the directory for userdata if it doesn't exist
	userdataDirPath := filepath.Join(config.artifactPath, config.VMName, "userdata")
	if err := os.MkdirAll(userdataDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create userdata directory: %v", err)
	}

	userDataContent, _ := CreateCloudInitDynamically(config.VMName, bootScriptPath)

	utils.LogOffwhite("CloudInit UserData set to:")
	utils.LogDottedLineDelimitedText(userDataContent)

	// Create a temporary user-data file
	userDataFilePath := filepath.Join(userdataDirPath, "user-data.txt")
	err := os.WriteFile(userDataFilePath, []byte(userDataContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write user-data file: %v", err)
	}

	// Path for the output user-data.img
	outputImgPath := filepath.Join(userdataDirPath, "user-data.img")

	// Generate the user-data.img
	cmd := exec.Command("cloud-localds", outputImgPath, userDataFilePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run cloud-localds: %v", err)
	}

	// Optionally, remove the temporary user-data file after creating the image
	// if err := os.Remove(userDataFilePath); err != nil {
	// 	log.Printf("Warning: failed to remove temporary user-data file: %v", err)
	// }

	return nil
}

// Create Image from Direct Provided User Data Image
func (config *VMConfig) GenerateCloudInitImgFromPath(userDataPathAbs string) error {
	// Create the directory for userdata if it doesn't exist fpr VM
	// data/artifacts/<vmname>/userdata

	userdataDirPath := filepath.Join(config.artifactPath, config.VMName, "userdata")
	if err := os.MkdirAll(userdataDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create userdata directory: %v", err)
	}

	var userDataContent string

	// If Preset is used - we set UserData in Memory on the Config
	// Check if --preset flag is used to override the manually passed File and Log a Warning
	if config.InlineUserdata != "" {
		userDataContent = config.InlineUserdata
	} else {
		log.Printf("Using Default UserData from Disk : %s", userDataPathAbs)
		userDataBytes, err := os.ReadFile(userDataPathAbs)
		if err != nil {
			return fmt.Errorf("failed to read boot script: %v", err)
		}
		userDataContent = configuration.SubstituteHostnameUserData(
			string(userDataBytes),
			config.VMName)
	}

	log.Print(utils.StructureResultWithHeadingAndColoredMsg(
		"CloudInit UserData Set To", utils.PEACH,
		userDataContent,
	))

	// utils.LogOffwhite("CloudInit UserData set to:")
	// utils.LogDottedLineDelimitedText(userDataContent)

	// Create a temporary user-data file
	userDataFilePath := filepath.Join(userdataDirPath, "user-data.txt")
	err := os.WriteFile(userDataFilePath, []byte(userDataContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write user-data file: %v", err)
	}

	// Path for the output user-data.img
	outputImgPath := filepath.Join(userdataDirPath, "user-data.img")

	// Generate the user-data.img
	cmd := exec.Command("cloud-localds", outputImgPath, userDataFilePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run cloud-localds: %v", err)
	}

	log.Printf("Successfully Created Cloud-Init user-data .img file: %s", userDataPathAbs)

	return nil
}

/*
Generates Cloud Init Data that the VM runs at Bootup

This is how the runcmd properly formatted should be:

#cloud-config
hostname: hadoop
password: password
chpasswd: { expire: False }
ssh_pwauth: True
runcmd:
  - |
    #!/bin/bash
    # Update and upgrade packages non-interactively
    sudo DEBIAN_FRONTEND=noninteractive apt-get update && sudo DEBIAN_FRONTEND=noninteractive apt-get -y upgrade
*/
func CreateCloudInitDynamically(vmName, bootScriptPath string) (string, error) {
	var scriptContent string
	if bootScriptPath != "" {
		content, err := os.ReadFile(bootScriptPath) // Ensure you're using the appropriate I/O library for your Go version
		if err != nil {
			return "", fmt.Errorf("failed to read boot script: %v", err)
		}
		scriptContent = string(content)
	}

	// The crucial adjustment: Indent the script content for inclusion in the cloud-init YAML.
	indentedScriptContent := "    " + strings.ReplaceAll(scriptContent, "\n", "\n    ")

	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
password: password
chpasswd: { expire: False }
ssh_pwauth: True
runcmd:
  - |
%s`, vmName, indentedScriptContent)

	// if err := os.WriteFile("testingdynamicinit.yaml", []byte(userDataContent), 0o644); err != nil {
	// 	return fmt.Errorf("failed to write user data file: %v", err)
	// }

	return userDataContent, nil
}

func (config *VMConfig) GenerateUserDataImgDefault() error {
	//	config.UserData = "data/userdata/default/user_data.txt"

	absoluteOutputImgPath, err := config.GetImageUserDataPath()
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
		return err
	}

	if err := config.CreateUserDataDir(); err != nil {
		log.Printf("Failed 	config.CreateUserDataDir()  ERROR:%s,", err)
		return err
	}

	log.Printf("Absolute Path of Image User Data:%s", absoluteOutputImgPath)
	userDataDir := "data/userdata/default"
	absoluteUserDataDir, err := filepath.Abs(userDataDir)
	if err != nil {
		log.Printf("Error getting absolute path for user data directory: %v", err)
		return err
	}
	log.Printf("Absolute path for user data directory: %s", absoluteUserDataDir)

	// Navigate to the user data directory
	if err := utils.NavigateToPath(absoluteUserDataDir); err != nil {
		log.Printf("Failed to navigate to user data directory: %v", err)
		return err
	}

	utils.PrintCurrentPath()

	log.Printf("Running cloud-localds %s user-data.txt", absoluteOutputImgPath)
	cmd := exec.Command("cloud-localds", absoluteOutputImgPath, "user-data.txt")
	return cmd.Run()
}

// Navigates to Dir for VM and creates the base image using qemu-img create -b
func (s *VMConfig) CreateBaseImage() error {
	log.Print("Creating Base Image")

	_ = s.navigateToDirWithISOImages()

	modifiedImageOutputPath, err := utils.CreateBaseImage(s.ImageURL, s.VMName)
	if err != nil {
		log.Printf("Failed to create base image ERROR:%s", err)
		return err
	}

	log.Printf("Successfully Created new Base Image at %s/%s",
		s.ImagesDir, modifiedImageOutputPath)

	_ = s.navigateToRoot()

	return nil
}

// SetupVM() creates a Mount Path to Copy Boot scripts into the VM, Copies Dynamic Data into the VM, and then clears the Mount Data.
func (s *VMConfig) SetupVM() error {
	utils.LogStep("MOUNTING IMAGE")

	_ = s.navigateToDirWithISOImages()

	modifiedImagePath := filepath.Join(s.VMName + "-vm-disk.qcow2")
	log.Printf("modified Image Path %s", modifiedImagePath)
	mountPath := "/mnt/" + s.VMName

	log.Printf("Mount Path Setup VM %s", mountPath)

	if err := utils.MountImage(modifiedImagePath, mountPath); err != nil {
		slog.Error("Failed Mounting Image", "error", err)
		return err
	}

	_ = s.navigateToRoot()

	// If Boot Files Present Copy Them
	if s.BootFilesDir != "" {
		utils.LogStep("COPYING SCRIPTS AND SYSTEMD SERVICES")
		if err := s.CopyVMSetupFiles(); err != nil {
			slog.Error("Failed Copying Boot Script and Service", "error", err)
			return err
		}
		log.Printf("Files Copied Successfully")
	}

	// If SystemD scripts defined - enable them
	if s.SystemdScript != "" {
		utils.LogStep("ENABLING SYSTEMD SERVICE AND UNMOUNTING")
		if err := s.EnableSystemdServices(); err != nil {
			slog.Error("Failed Enabling Systemd Services", "error", err)
			return err
		}
		log.Printf("Systemd services on Image enabled successfully")
	}

	log.Printf("Unmounting Image and Clearing Temp Mount Path")

	if err := utils.UnmountImage(mountPath); err != nil {
		slog.Error("Failed Unmounting Image", "error", err)
		return err
	}

	if err := utils.ClearMountPath(s.VMName); err != nil {
		slog.Error("Failed Unmounting Image", "error", err)
		return err
	}

	_ = s.navigateToRoot()

	return nil
}

// CreateVM() uses libvirtd to create the VM and boot it. The state will change to Running and the boot scripts will run followed by systemd services
func (s *VMConfig) CreateVM() error {
	err := s.navigateToRoot()
	if err != nil {
		log.Printf("Failed to Navigate to Root Dir. Virt-install must be ran with relative pathing. :%s", err)
	}

	modifiedImagePath := filepath.Join(s.ImagesDir, s.VMName+"-vm-disk.qcow2")
	vm_userdata_img := filepath.Join("data", "artifacts", s.VMName, "userdata", "user-data.img")

	cmd := exec.Command("virt-install", "--name", s.VMName,
		"--virt-type", "kvm",
		"--memory", fmt.Sprint(s.Memory),
		"--vcpus", fmt.Sprint(s.CPUCores),
		"--disk", "path="+modifiedImagePath+",device=disk",
		"--disk", "path="+vm_userdata_img+",format=raw",
		"--graphics", "none",
		"--boot", "hd,menu=on",
		"--network", "network=default",
		"--os-variant", "ubuntu18.04", "--noautoconsole",
		// "--print-xml", // to test the XML structure
	)

	log.Printf("%sCreating Virtual Machine%s %s%s%s%s: %s\n", utils.BOLD, utils.NC, utils.BOLD, utils.COOLBLUE, s.VMName, utils.NC, cmd.String())

	var stderr bytes.Buffer // Capture stderr
	cmd.Stderr = &stderr
	// cmd.Stdout = &stderr if we need to print XML

	err = cmd.Run()
	if err != nil {
		log.Printf("ERROR Failed to Create VM error=%q", stderr.String())
		return err
	}
	//	log.Printf("VM XML Description:\n%s", stderr.String())

	return nil
}

func (s *VMConfig) EnableSystemdServices() error {
	mountPath := "/mnt/" + s.VMName

	for _, serviceName := range s.EnableServices {

		if err := utils.EnableSystemdService(mountPath, serviceName); err != nil {
			slog.Error("Failed to enable systemd service", "service", serviceName, "error", err)
			return err
		}
		log.Printf("Systemd service %s enabled successfully", serviceName)
	}

	return nil
}

/*
Pulls the defined artifacts from the VM - such as boot outputs required by other Virtual Machines.

Wait until the VM is in the running state.

Once the VM is running - wait for the artifacts to exist before pulling them.

Behind the hood, we are simply using virt-ls to check the contents of a folder
& parsing the output of virt-ls to check for the desired files.

Once the files exist - we can proceed to pull the artifacts.

Usage:

	config := vm.NewVMConfig("kubecontrol").
							SetArtifacts([]string{
								"/home/ubuntu/kubeadm-init.log"
							})

	if err := config.PullArtifacts(); err != nil {
		log.Printf("Failed to pull Artifacts from VM ERROR:%s,", err)
		return err
	}
*/
func (s *VMConfig) PullArtifacts() error {
	if len(s.Artifacts) == 0 {
		log.Printf("No Artifacts Specified...")
		return nil
	}
	log.Printf("Waiting for VM to complete initialization before pulling artifacts")
	time.Sleep(3 * time.Second)

	timeout := time.After(3 * time.Minute)

	tick := time.NewTicker(15 * time.Second)

	vmChecked := false
	checkCount := 0

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for VM and artifacts to be ready")
		case <-tick.C:

			log.Printf("%sChecking for VM and Artifacts readiness... %s%s", utils.SAND, utils.NC, strings.Repeat(utils.BOLD+utils.YELLOW+"."+utils.NC, checkCount))
			checkCount++

			if !vmChecked {
				if running, _ := utils.IsVMRunning(s.VMName); running {
					log.Printf("VM is now running.")
					vmChecked = true
				}
			} else {
				allArtifactsExist := true
				for _, artifact := range s.Artifacts {
					if !utils.FileExistsInVM(s.VMName, artifact) {
						allArtifactsExist = false
						log.Printf("Checking for artifact: %s ... Not found", artifact)
						break
					}
					log.Printf("Artifact %s%s%s found", utils.BOLD, artifact, utils.NC)
				}

				if allArtifactsExist {
					log.Printf("%s %sAll Artifacts Exist%s - Pulling from VM %s", utils.TICK_GREEN, utils.BOLD, utils.NC, s.VMName)
					for _, artifact := range s.Artifacts {
						if err := s.PullFromVM(artifact); err != nil {
							return err
						}
					}
					return nil
				}
			}
		}
	}
}

func CheckTimeout() error {
	timeout := time.After(30 * time.Minute)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-timeout:
			log.Printf("Timed out - returning")
			return fmt.Errorf("Timed Out")
		case <-ticker.C:
			log.Printf("Ticked, check something")
			count += 10
			if count > 20 {
				return nil
			}
		}
	}
}

func (s *VMConfig) PullArtifactsOg() error {
	if len(s.Artifacts) == 0 {
		log.Printf("No Artifacts Specified...")
		return nil
	}
	log.Printf("Waiting for VM to complete initialization before pulling artifacts")
	time.Sleep(3 * time.Second)

	timeout := time.After(30 * time.Minute)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:

			return fmt.Errorf("timeout waiting for VM and artifacts to be ready")
		case <-ticker.C:

			if running, _ := utils.IsVMRunning(s.VMName); running {
				allArtifactsExist := true
				for _, artifact := range s.Artifacts {
					if !utils.FileExistsInVM(s.VMName, artifact) {
						allArtifactsExist = false
						break
					}
				}

				if allArtifactsExist {
					log.Printf("All Artifacts Exist - pulling")
					for _, artifact := range s.Artifacts {
						if err := s.PullFromVM(artifact); err != nil {
							return err
						}
					}
					return nil
				}
			}
		}
	}
}

func (s *VMConfig) CopyVMSetupFiles() error {
	mountPath := "/mnt/" + s.VMName
	log.Printf("Navigating to boot directory for VM setup files")
	if err := s.navigateToBootupDir(); err != nil {
		log.Printf("Error navigating to boot directory: %v", err)
		return err
	}

	setupDir := filepath.Join(s.RootDir, s.BootFilesDir)
	ubuntuUserPath := filepath.Join(mountPath, "home", "ubuntu")
	systemdPath := filepath.Join(mountPath, "etc", "systemd", "system")

	log.Printf("Reading files from setup directory: %s", setupDir)
	files, err := os.ReadDir(setupDir)
	if err != nil {
		log.Printf("Error reading setup directory: %v", err)
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		sourceFilePath := filepath.Join(setupDir, file.Name())
		var destDir string
		var fileMode fs.FileMode

		// Check if file is a service file
		if strings.HasSuffix(file.Name(), ".service") {
			destDir = systemdPath
			fileMode = 0o644 // Service files usually have 0644 permissions
		} else {
			destDir = ubuntuUserPath
			fileMode = 0o755 // Script files are typically executable
		}

		destFilePath := filepath.Join(destDir, file.Name())
		log.Printf("Copying file %s to %s", sourceFilePath, destFilePath)

		if err := exec.Command("sudo", "mkdir", "-p", destDir).Run(); err != nil {
			log.Printf("Error creating directory: %v", err)
			return err
		}
		if err := exec.Command("sudo", "cp", sourceFilePath, destFilePath).Run(); err != nil {
			log.Printf("Error copying file: %v", err)
			return err
		}
		if err := exec.Command("sudo", "chmod", fmt.Sprintf("%o", fileMode), destFilePath).Run(); err != nil {
			log.Printf("Error setting file permissions: %v", err)
			return err
		}
	}

	log.Printf("VM setup files copied successfully")
	return nil
}

func (s *VMConfig) PullFromVM(path string) error {
	local := filepath.Join(s.artifactPath, s.VMName)

	if err := utils.CreateDirIfNotExist(local); err != nil {
		log.Printf("Failed 	utils.CreateDirIfNotExist(local) ERROR:%s,", err)
		return err
	}

	log.Printf("artifact path:%s VMName:%s local:%s", s.artifactPath, s.VMName, local)

	cmd := exec.Command("sudo", "virt-copy-out", "-d", s.VMName, path, local)

	log.Printf("Running sudo virt-copy-out -d %s %s %s", s.VMName, path, local)

	log.Printf("Running command: %s\n", cmd.String())

	if err := cmd.Run(); err != nil {
		log.Printf("Failed to extract artifact %s ERROR:%s,", path, err)
		return err
	}

	//	sudo virt-copy-out -d kubecontrol /home/ubuntu/kubeadm-init.log .

	return nil
}

/*
	Creates an SSH Client to Interact with the Node

	For Secure Connection consider using a SSH Key if the Bastion Host is not Strengthened

Usage:

	client, err := config.GetSSHClient()
	// err handling
	defer client.Close()

	output, _ := client.RunCommand("uptime")
	log.Printf("VM Uptime:%s", output)
*/
func (s *VMConfig) GetSSHClient() (*network.VMClient, error) {
	vm_running, err := utils.IsVMRunning(s.VMName)
	if err != nil {
		log.Printf("Failed to check VM Running Status. Are you sure VM %s is running?", s.VMName)
		return nil, err
	}

	if !vm_running {
		log.Printf("VM %s is not running - an SSH Client can only be created for an active booted VM.", s.VMName)
	}

	utils.LogWarning("WARNING: Using the Insecure SSH Client. Prefer using SSH Key file authentication. Remember to call defer client.Close() after creating a client.")

	username := "ubuntu"
	password := "password"
	ip, _ := network.GetVMIPAddr(s.VMName)

	client, err := network.NewInsecureSSHClientVM(s.VMName, ip.String(), username, password)
	if err != nil {
		log.Printf("Error creating SSH client:%s", err)
		return nil, err
	}

	return client, nil
}

func (s *VMConfig) PullFromVMToPath(path, local string) error {
	if err := utils.CreateDirIfNotExist(local); err != nil {
		log.Printf("Failed 	utils.CreateDirIfNotExist(local) ERROR:%s,", err)
		return err
	}

	cmd := exec.Command("sudo", "virt-copy-out", "-d", s.VMName, path, local)

	log.Printf("Running sudo virt-copy-out -d %s %s %s", s.VMName, path, local)

	log.Printf("Running command: %s\n", cmd.String())

	_ = cmd.Run()

	//	sudo virt-copy-out -d kubecontrol /home/ubuntu/kubeadm-init.log .

	return nil
}

func (s *VMConfig) navigateToBootupDir() error {
	if err := os.Chdir(s.BootFilesDir); err != nil {

		log.Printf("Failed to change directory: %v", err)

		return err
	}

	return nil
}

// Navigates to the Path where we cache all the Base OS Images - so we can extend it to create an Image for the VM (data/images)
func (s *VMConfig) navigateToDirWithISOImages() error {
	if err := os.Chdir(s.ImagesDir); err != nil {

		log.Printf("Failed to change directory: %v", err)

		return err
	}

	return nil
}

// Navigate back to Root Dir as Commands are Path Relevant for QEMU/KVM
func (s *VMConfig) navigateToRoot() error {
	if err := os.Chdir(s.RootDir); err != nil {

		log.Printf("Failed to change directory: %v", err)

		return err
	}

	return nil
}

func (s *VMConfig) SetArtifactDir(path string) {
	s.artifactPath = path
}

func (config *VMConfig) GetImageUserDataPath() (string, error) {
	outputImgPath := filepath.Join(config.artifactPath, config.VMName, "userdata", "user-data.img")
	absoluteOutputImgPath, err := filepath.Abs(outputImgPath)
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
		return "", err
	}
	return absoluteOutputImgPath, nil
}

func (config *VMConfig) CreateUserDataDir() error {
	userDataDir := filepath.Join(config.artifactPath, config.VMName, "userdata")

	if err := utils.CreateDirIfNotExist(userDataDir); err != nil {
		log.Printf("Error creating user data directory: %v", err)
		return err
	}

	return nil
}
