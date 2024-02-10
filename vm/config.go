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

	"kvmgo/network"
	"kvmgo/utils"
)

type VMConfig struct {
	VMName         string
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

func (config *VMConfig) SetMemory(memory_mb int) *VMConfig {
	config.Memory = memory_mb
	return config
}

func (config *VMConfig) SetBootServices(services []string) *VMConfig {
	config.EnableServices = services
	return config
}

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
	return config
}

func (config *VMConfig) DefaultUserData() *VMConfig {
	config.UserData = "data/userdata/default/user_data.txt"
	return config
}

func (s *VMConfig) PullImage() {
	err := utils.PullImage(s.ImageURL, s.ImagesDir)
	if err != nil {
		slog.Error("Failed HTTP GET", "error", err)
		os.Exit(1)
	}
}

func (config *VMConfig) GenerateCustomUserDataImg() error {
	// Create the directory for userdata if it doesn't exist
	userdataDirPath := filepath.Join(config.artifactPath, config.VMName, "userdata")
	if err := os.MkdirAll(userdataDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create userdata directory: %v", err)
	}

	// Define the content of the user-data file
	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
password: password
chpasswd: { expire: False }
ssh_pwauth: True
`, config.VMName)

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

func (s *VMConfig) CreateBaseImage() error {
	log.Print("Creating Base Image")

	s.navigateToDirWithISOImages()

	modifiedImageOutputPath, err := utils.CreateBaseImage(s.ImageURL, s.VMName)
	if err != nil {
		log.Printf("Failed to create base image ERROR:%s", err)
		return err
	}

	log.Printf("Created new Base Image at %s/%s", s.ImagesDir, modifiedImageOutputPath)

	s.navigateToRoot()
	return nil
}

func (s *VMConfig) SetupVM() error {
	utils.LogStep("MOUNTING IMAGE")

	s.navigateToDirWithISOImages()

	modifiedImagePath := filepath.Join(s.VMName + "-vm-disk.qcow2")
	log.Printf("modified Image Path %s", modifiedImagePath)
	mountPath := "/mnt/" + s.VMName

	log.Printf("Mount Path Setup VM %s", mountPath)

	if err := utils.MountImage(modifiedImagePath, mountPath); err != nil {
		slog.Error("Failed Mounting Image", "error", err)
		return err
	}

	s.navigateToRoot()

	utils.LogStep("COPYING SCRIPTS AND SYSTEMD SERVICES")

	if err := s.CopyVMSetupFiles(); err != nil {
		slog.Error("Failed Copying Boot Script and Service", "error", err)
		return err
	}

	log.Printf("Files Copied Successfully")

	utils.LogStep("ENABLING SYSTEMD SERVICE AND UNMOUNTING")

	if err := s.EnableSystemdServices(); err != nil {
		slog.Error("Failed Enabling Systemd Services", "error", err)
		return err
	}

	log.Printf("Systemd services on Image enabled successfully")

	if err := utils.UnmountImage(mountPath); err != nil {
		slog.Error("Failed Unmounting Image", "error", err)
		return err
	}

	log.Printf("Unmounting Image")

	s.navigateToRoot()

	return nil
}

func (s *VMConfig) CreateVM() error {
	s.navigateToRoot()

	modifiedImagePath := filepath.Join(s.ImagesDir, s.VMName+"-vm-disk.qcow2")
	vm_userdata_img := filepath.Join("data", "artifacts", s.VMName, "userdata", "user-data.img")

	cmd := exec.Command("virt-install", "--name", s.VMName, "--virt-type", "kvm", "--memory", fmt.Sprint(s.Memory), "--vcpus", fmt.Sprint(s.CPUCores), "--boot", "hd,menu=on", "--disk", "path="+modifiedImagePath+",device=disk", "--disk", "path="+vm_userdata_img+",format=raw", "--graphics", "none", "--network", "network=default", "--os-type", "Linux", "--os-variant", "ubuntu18.04", "--noautoconsole")

	log.Printf("%sCreating Virtual Machine%s %s%s%s: %s\n", utils.BOLD, utils.NC, utils.PURP_HI, s.VMName, utils.NC, cmd.String())

	// Capture standard error
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("ERROR Failed to Create VM error=%q", stderr.String())
		return err
	}

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

	timeout := time.After(30 * time.Minute)
	tick := time.Tick(15 * time.Second)

	vmChecked := false
	checkCount := 0

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for VM and artifacts to be ready")
		case <-tick:

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

func (s *VMConfig) PullArtifactsOg() error {
	if len(s.Artifacts) == 0 {
		log.Printf("No Artifacts Specified...")
		return nil
	}
	log.Printf("Waiting for VM to complete initialization before pulling artifacts")
	time.Sleep(3 * time.Second)

	timeout := time.After(30 * time.Minute)
	tick := time.Tick(15 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for VM and artifacts to be ready")
		case <-tick:
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

	if vm_running == false {
		log.Printf("VM %s is not running - an SSH Client can only be created for an active booted VM.", s.VMName)
	}

	utils.LogWarning("WARNING: Using the Insecure SSH Client. Prefer using SSH Key file authentication. Remember to call defer client.Close() after creating a client.")

	username := "ubuntu"
	password := "password"
	ip, _ := network.GetVMIPAddr(s.VMName)

	client, err := network.NewInsecureSSHClientVM(s.VMName, ip, username, password)
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

	cmd.Run()

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

func (s *VMConfig) navigateToDirWithISOImages() error {
	if err := os.Chdir(s.ImagesDir); err != nil {

		log.Printf("Failed to change directory: %v", err)

		return err
	}

	return nil
}

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
