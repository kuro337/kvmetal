package vm

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"kvmgo/constants"
	"kvmgo/lib"
	"kvmgo/network"
	"kvmgo/types/fpath"
	"kvmgo/utils"

	"gopkg.in/yaml.v2"
)

type VMConfig struct {
	VMName         string `json:"vm_name" yaml:"vm_name"`
	InlineUserdata string `json:"inline_userdata" yaml:"inline_userdata"`
	ImageURL       string `json:"image_url" yaml:"image_url"`

	// Central Images Dir
	ImagesDir       string       `json:"images_dir" yaml:"images_dir"`
	BootFilesDir    string       `json:"boot_files_dir" yaml:"boot_files_dir"`
	ScriptsDir      string       `json:"scripts_dir" yaml:"scripts_dir"`
	BootScript      string       `json:"boot_script" yaml:"boot_script"`
	SystemdScript   string       `json:"systemd_script" yaml:"systemd_script"`
	UserData        string       `json:"user_data" yaml:"user_data"`
	RootDir         string       `json:"root_dir" yaml:"root_dir"`
	CPUCores        int          `json:"cpu_cores" yaml:"cpu_cores"`
	Memory          int          `json:"memory" yaml:"memory"`
	EnableServices  []string     `json:"enable_services" yaml:"enable_services"`
	Artifacts       []string     `json:"artifacts" yaml:"artifacts"`
	Disks           []DiskConfig `json:"disks" yaml:"disks"`
	sshPub          string
	ArtifactPath    string         `json:"artifact_path" yaml:"artifact_path"`
	ArtifactsPathFP fpath.FilePath `json:"artifacts_path_fp" yaml:"artifacts_path_fp"`
	ImagesPathFP    fpath.FilePath `json:"images_path_fp" yaml:"images_path_fp"`
	DisksPathFP     fpath.FilePath `json:"disks_path_fp" yaml:"disks_path_fp"`
	CreateDirsInit  bool           `json:"create_dirs_init" yaml:"create_dirs_init"`

	// kvm img manager
	imgManager *lib.ImageManager

	baseImgName string
}

func (vm *VMConfig) YAML() (string, error) { // Convert to YAML
	yamlData, err := yaml.Marshal(&vm)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func (vm *VMConfig) GetImagesPath() (string, error) { // Convert to YAML

	str := fmt.Sprintf(`1. Generating the Base Image

1. Navigate to vm.ImagesDir() to generate the Base Qemu Image for VM

Requirements : OS Backing Image must be Present & Reachable from cwd 

cmd : qemu-img create -b <backing_img>.img -F qcow2 -f qcow2 <desired_vm_img_nam>.qcow2 20G 

    _ = s.navigateToAbsPath(s.ImagesDir) 
    backing_img := filepath.Base(imageURL) // extracts os name from URL of ubuntu image
    desiredImgName := vmName + "-vm-disk.qcow2"

This option specifies the backing file. 
The new image will be created as a copy-on-write (COW) image based on this backing file. 
The backing file is typically a base image that contains a standard installation of an operating system.
    `, vm.ImagesDir)

	log.Println(str)

	// 1. Calls vm.CreateBaseImage()

	log.Printf("Navigate to %s and use base OS image to generate VM Image.\n", vm.ImagesDir)

	// 1.A - Calls utils.CreateBaseImage(vm.ImageURL, vm.VMName)
	log.Printf("vm.CBI calls utils.CBI(%s,%s). [createBaseImage]", vm.ImageURL, vm.VMName)

	// utils.CBI calls  	modifiedImageOutputPath, err := utils.CreateBaseImage(s.ImageURL, s.VMName)

	backingImgFile := filepath.Base(vm.ImageURL)
	desiredVMImg := vm.VMName + "-vm-disk.qcow2"

	createVMImage := fmt.Sprintf("utils.CBI() expects to find %s and generates the VM Img: %s\n", backingImgFile, desiredVMImg)

	log.Println("Step 1A. %s", createVMImage)

	// Step 2. CreateDisks

	log.Printf("2. vm.CreateDisks() - calls utils.CreateDirIfNotExist(vm.DisksPath())")

	disksPath := vm.DisksPath()

	log.Printf("Navigate to Disks Path: %s\n", &disksPath)

	var disksPaths []string
	var diskPathCmds []string

	log.Println("For each path in vm.Disks - creates the disk by running utils.CreateDiskQcow(diskPath)")

	for _, disk := range vm.Disks {
		diskPathQemu, _ := disk.DiskPathFP.Relative()
		disksPaths = append(disksPaths, diskPathQemu)

		diskCmd := fmt.Sprintf("qemu-img create -f qcow2 %s <diskSize>G", diskPathQemu)
		diskPathCmds = append(diskPathCmds, diskCmd)

		log.Println("utils.CDQCow runs: DiskCmd: %s\n", diskCmd)

	}

	log.Println("Navigate back to Root")

	// 3.

	userdataDirPath := filepath.Join(vm.ArtifactPath, "userdata")

	log.Printf("Step 3: Generating User Data: %s\n", userdataDirPath)

	log.Println("Calls vmc.GenerateCloudInitImgFromPath()")

	log.Println("If vm.InlineUserdata empty - uses default Userdata with zsh installed")

	if vm.InlineUserdata == "" {
		log.Println("will use Default User Data using:  configuration.SubstituteHostNameAndFqdnUserdataSSHPublicKey")
	}

	log.Println("Generates user-data.txt, metadata , user-data.img at: %s/\n", userdataDirPath)

	log.Println("Step 4 : Runs virt-install from Root")

	//
	//    log.Println("If

	generatedVmImg := filepath.Join(vm.ImagesDir, vm.VMName+"-vm-disk.qcow2")
	vm_userdata_img := filepath.Join("data", "artifacts", vm.VMName, "userdata", "user-data.img")

	if f, _ := fpath.FileExists(generatedVmImg); !f {
		log.Println(utils.TurnError(fmt.Sprintf("VM OS Image not found: %s", generatedVmImg)))
	}

	if f, _ := fpath.FileExists(vm_userdata_img); !f {
		log.Println(utils.TurnError(fmt.Sprintf("VM Userdata Image not found: %s", generatedVmImg)))
	}

	log.Println("If either cannot be found - will fail: %s %s\n", generatedVmImg, vm_userdata_img)

	cmdArgs := []string{
		"--name", vm.VMName,
		"--virt-type", "kvm",
		"--memory", fmt.Sprint(vm.Memory),
		"--vcpus", fmt.Sprint(vm.CPUCores),
		"--disk", "path=" + generatedVmImg + ",device=disk",
		"--disk", "path=" + vm_userdata_img + ",format=raw",
		"--graphics", "none",
		"--boot", "hd,menu=on",
		"--network", "network=default",
		"--os-variant", "ubuntu18.04",
		"--noautoconsole",
	}

	virtInstallCmd := strings.Join(cmdArgs, "")

	log.Println("Runs virt-install cmd : %s\n", virtInstallCmd)

	fpath.LogCwd()

	yamlData, err := yaml.Marshal(&vm)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func (fp FilePathWrapper) MarshalYAML() (interface{}, error) {
	return fp.Get(), nil
}

func (fp FilePathWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(fp.Get())
}

type FilePathWrapper struct {
	fpath.FilePath
}

// DiskConfig used to manage disks for a VM - methods to add and backup Disks.
// qemu-img create -f qcow2 /var/lib/libvirt/images/myvm-openebs-disk.qcow2 50G
type DiskConfig struct {
	DiskName   string // uses for
	Size       int
	Persistent bool
	DiskPathFP fpath.FilePath
}

/* Methods to use FilePath type */
func NewDiskConfig(diskPath string, size int) (*DiskConfig, error) {
	fullPath, err := fpath.NewPath(diskPath, false)
	if err != nil {
		log.Printf("Failed to Create Qualified Abs Path from Base Path %s ERROR:%s", diskPath, err)
	}
	diskConf := DiskConfig{
		Size:       size,
		DiskPathFP: *fullPath,
	}
	return &diskConf, nil
}

/* Methods to use FilePath type */
func (config *VMConfig) SetArtifactPath(filePath fpath.FilePath) *VMConfig {
	config.ArtifactsPathFP = filePath
	return config
}

func (config *VMConfig) SetImagePath(filePath fpath.FilePath) *VMConfig {
	config.ImagesPathFP = filePath
	return config
}

// Base Image name for imgManager
func (config *VMConfig) SetBaseImageName(name string) *VMConfig {
	config.baseImgName = name
	return config
}

// WriteConfigYAML saves the YAML Config for the VM
func (config *VMConfig) WriteConfigYaml() error {
	if config.ArtifactPath == "" || config.UserData == "" {
		return fmt.Errorf("Do not call WriteConfig preemptively")
	}

	yaml, err := config.YAML()
	if err != nil {
		return fmt.Errorf("Error Marshalling: %s\n", err)
	}

	userdataDirPath := filepath.Join(config.ArtifactPath, "userdata", fmt.Sprintf("%s-vmconfig.yaml", config.VMName))

	log.Printf("Saving YAML Configuration at: %s\n", userdataDirPath)

	os.WriteFile(userdataDirPath, []byte(yaml), 0o644)

	return nil
}

func (config *VMConfig) SetDisksPath(filePath fpath.FilePath) *VMConfig {
	config.DisksPathFP = filePath
	return config
}

func (config *VMConfig) SetImgManager(manager string) error {
	imgManager, err := lib.NewImageMgr(manager, "")
	if err != nil {
		return fmt.Errorf("failed to create imgMgr image, %s\n", err)
	}
	config.imgManager = imgManager
	return nil
}

func (d DiskConfig) QcowName() string {
	return d.DiskName + ".qcow2"
}

// Call to get Paths of qcow Disks - at data/artifacts/vmname/disks/diskn.qcow etc.
func (config *VMConfig) GetDiskPaths() []string {
	var diskPaths []string
	for _, disk := range config.Disks {
		diskPaths = append(diskPaths, filepath.Join(config.DisksPath(), disk.QcowName()))
	}
	return diskPaths
}

// Create the Disk Paths for VM's Disk Images
func (config *VMConfig) GetDiskPathsFP() ([]fpath.FPath, error) {
	var diskPaths []fpath.FPath
	for _, disk := range config.Disks {
		fp, err := fpath.NewPath(filepath.Join(config.DisksPath(), disk.QcowName()), false)
		if err != nil {
			log.Printf("Failed to Create Abs Path for VM Secondary Disks. ERROR:%s", err)
			return nil, err
		}
		diskPaths = append(diskPaths, fp)
	}

	return diskPaths, nil
}

// DisksPath() returns the path where the VM Specific Disk should be created as an artifact
// data/artifacts/vm/disks/ (currently created at d/a/vm/ ) 1 level higher
func (config *VMConfig) DisksPath() string {
	return filepath.Join(config.ArtifactPath, "disks")
}

func (config *VMConfig) UserdataPath() string {
	return filepath.Join(config.ArtifactPath, "userdata")
}

func NewVMConfig(vmName string) *VMConfig {
	pwd, _ := os.Getwd()

	return &VMConfig{
		VMName:  vmName,
		RootDir: pwd,
		//	artifactPath: "data/artifacts",
	}
}

func NewKVM(vmName string) *VMConfig {
	config := &VMConfig{
		VMName: vmName,
	}
	// config.artifactPath = "data/artifacts"
	return config
}

func (config *VMConfig) SetImageURL(url string) *VMConfig {
	config.ImageURL = url
	return config
}

func (config *VMConfig) SetArtifactsDir(vmArtifactsPath string) *VMConfig {
	config.ArtifactPath = vmArtifactsPath
	return config
}

// Initializes and Validates the Dirs upon Creation of the Config
func (config *VMConfig) InitDirs(diskConfig DiskConfig) *VMConfig {
	// Userdata Dir
	// Disks Dir data/artifacts/vmname/disks

	userdataDirPath := filepath.Join(config.ArtifactPath, "userdata")
	if err := os.MkdirAll(userdataDirPath, 0o755); err != nil {
		log.Fatalf("failed to create userdata directory: %v", err)
	}

	disksDirPath := filepath.Join(config.ArtifactPath, "disks")
	if err := os.MkdirAll(disksDirPath, 0o755); err != nil {
		log.Fatalf("failed to create VM Disks directory: %v", err)
	}

	return config
}

// Add another Disk for the VM - such as a Persistent vdb disk for OpenEBS
func (config *VMConfig) AddDisk(diskConfig DiskConfig) *VMConfig {
	config.Disks = append(config.Disks, diskConfig)
	return config
}

// Sets the Public Key to be used for secure SSH access
func (config *VMConfig) SetPubkey(sshpubpath string) *VMConfig {
	config.sshPub = utils.ReadFileFatal(sshpubpath)
	return config
}

func (config *VMConfig) SetCores(vcpus int) *VMConfig {
	config.CPUCores = vcpus
	if vcpus == 0 {
		config.CPUCores = 1
	}

	return config
}

// Sets CloudInitUserData dynamically from Presets
func (config *VMConfig) SetCloudInitDataInline(cloudInitUserData string) *VMConfig {
	if cloudInitUserData != "" {
		utils.LogStep("Using Dynamic Preset Config for Userdata")
		config.InlineUserdata = cloudInitUserData
	}
	return config
}

func (config *VMConfig) SetMemory(memory_mb int) *VMConfig {
	config.Memory = memory_mb
	if memory_mb == 0 {
		config.Memory = 2048
	}
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
		// log.Print("No User Data Passed - Setting Default CloudInit UserData")
		config.DefaultUserData()
	}
	return config
}

/*
Generate Metadata File to Resolve VM behavior of setting the FQDN Properly on Boot
Responsible for sending DHCP Request and Boot Scripts

	// meta-data (data/artifacts/<vm>)
	instance-id: ubuntu-vm
	local-hostname: ubuntu-vm
*/
func (config *VMConfig) SmbiosMetadata() string {
	return fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n",
		config.VMName, config.VMName)
}

func (config *VMConfig) DefaultUserData() *VMConfig {
	config.UserData = "/home/kuro/Documents/Code/Go/kvmgo/data/userdata/default/user_data.txt"
	return config
}

func (s *VMConfig) PullImage() {
	log.Print(utils.TurnSuccess(fmt.Sprintf("Old s.ImagesDir:%s | New ImgsDir %s | Images URL: %s",
		s.ImagesDir, s.ImagesPathFP.Get(), s.ImageURL)))

	// err := utils.PullImage(s.ImageURL, s.ImagesDir)
	err := utils.PullImage(s.ImageURL, s.ImagesPathFP.Get())
	if err != nil {
		slog.Error("Failed HTTP GET", "error", err)
		os.Exit(1)
	}
}

// Create Image (user-data.img) from UserData for VM
//
//  1. Creates user-data & metedata temp files
//  2. Runs cloud-localds user-data.img user-data meta-data to create the UserData Disk
//  3. This is the persistent Disk required to access the VM
//
// Artifacts :  user-data.txt, meta-data ,  userdata.img
//
// Dest : data/artifacts/<vmname>/userdata/
func (config *VMConfig) GenerateCloudInitImgFromPath() error {
	// Create the directory for userdata if it doesn't exist fpr VM
	// data/artifacts/<vmname>/userdata

	// userdataDirPath := filepath.Join(config.artifactPath, config.VMName, "userdata")
	userdataDirPath := filepath.Join(config.ArtifactPath, "userdata")
	if err := os.MkdirAll(userdataDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create userdata directory: %v", err)
	}

	var userDataContent string

	if config.InlineUserdata != "" {
		userDataContent = config.InlineUserdata
	} else {

		log.Print("Using Default userdata with ZSH Shell. Optionally use DefaultUserdata to launch with Bash.")

		userDataContent = configuration.SubstituteHostNameAndFqdnUserdataSSHPublicKey(
			//			constants.DefaultUserdata,
			constants.DefaultUserDataShellZsh,
			config.VMName,
			config.sshPub)
	}

	log.Print(utils.StructureResultWithHeadingAndColoredMsg(
		"CloudInit UserData Set To", utils.PEACH,
		userDataContent,
	))

	/// 1. Creates user-data & metedata temp files
	//  2. Runs cloud-localds user-data.img user-data meta-data to create the UserData Disk
	//  3. This is the persistent Disk required to access the VM

	// Create a temporary user-data file
	userDataFilePath := filepath.Join(userdataDirPath, "user-data.txt")
	err := os.WriteFile(userDataFilePath, []byte(userDataContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write user-data file: %v", err)
	}

	// Path for the meta-data file
	metaDataFilePath := filepath.Join(userdataDirPath, "meta-data")
	metaDataContent := config.SmbiosMetadata()

	// Write the meta-data content to a file
	err = os.WriteFile(metaDataFilePath, []byte(metaDataContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write meta-data file: %v", err)
	}

	// Now, use both user-data and meta-data to generate the cloud-init disk
	outputImgPath := filepath.Join(userdataDirPath, "user-data.img")
	cmd := exec.Command("cloud-localds", outputImgPath, userDataFilePath, metaDataFilePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run cloud-localds with meta-data: %v", err)
	}

	log.Printf("Successfully created cloud-init disk with user-data and meta-data: %s", outputImgPath)

	config.WriteConfigYaml()

	return nil
}

// GetUserData for the VM from the Config
func GetUserData(config *VMConfig) string {
	if config.InlineUserdata != "" {
		return config.InlineUserdata
	}
	log.Print("Using Default userdata with ZSH Shell. Optionally use DefaultUserdata to launch with Bash.")
	return configuration.SubstituteHostNameAndFqdnUserdataSSHPublicKey(
		constants.DefaultUserDataShellZsh,
		config.VMName,
		config.sshPub)
}

// GetMetaData for the VM using the Name
func GetMetaData(vmName string) string {
	return fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n",
		vmName, vmName)
}

func CreateCloudInitData(userdata, metadata, path string) error {
	userDataFilePath := filepath.Join(path, "user-data.txt")
	metaDataFilePath := filepath.Join(path, "meta-data")
	outputImgPath := filepath.Join(path, "user-data.img")

	// write the userdata
	if err := os.WriteFile(userDataFilePath, []byte(userdata), 0o644); err != nil {
		return fmt.Errorf("failed to write user-data file: %v", err)
	}

	// Write the meta-data content to a file
	if err := os.WriteFile(metaDataFilePath, []byte(metadata), 0o644); err != nil {
		return fmt.Errorf("failed to write meta-data file: %v", err)
	}

	// combine both for the cloud init disk
	cmd := exec.Command("cloud-localds", outputImgPath, userDataFilePath, metaDataFilePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run cloud-localds with meta-data: %v", err)
	}

	log.Printf("Successfully created cloud-init disk with user-data and meta-data: %s", outputImgPath)

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
// Note : Navgates + Creates VM Image in CENTRAL IMAGES dir
func (s *VMConfig) CreateBaseImage() error {
	log.Print("Creating Base Image. Navigating to dir with Base OS Image: %s\n", s.ImagesDir)

	// 	_ = s.navigateToDirWithISOImages()
	_ = s.navigateToAbsPath(s.ImagesDir) // navigate to central images dir

	log.Print(utils.TurnSuccess(fmt.Sprintf("CREATBASEIMAGE() Old s.ImagesDir:%s | New ImgsDir %s",
		s.ImagesDir, s.ImagesPathFP.Get())))

	modifiedImageOutputPath, err := utils.CreateBaseImage(s.ImageURL, s.VMName)
	if err != nil {
		log.Printf("Failed to create base image ERROR:%s", err)
		return err
	}

	utils.TurnSuccess(fmt.Sprintf("Successfully Created new Base Image at %s/%s",
		s.ImagesDir, modifiedImageOutputPath))

	_ = s.navigateToRoot()

	return nil
}

// Navigates to Dir for VM and creates the base image using qemu-img create -b
// Disk created in data/artifacts/vm/
func (s *VMConfig) CreateDisks() error {
	// uses artifacts dir and hcoded + "disks"
	utils.CreateDirIfNotExist(s.DisksPath())

	fpath.LogCwd()
	log.Print("Creating VM Disks")

	err := s.navigateToAbsPath(s.DisksPath())
	if err != nil {
		log.Fatalf("FAILURE Generating Secondary Disks : %s", err)
	}

	fmt.Println(utils.LogMainAction("DEBUG PATH ISSUE"))

	diskPath := s.DisksPath()

	a := s.DisksPathFP.Abs()
	b, _ := s.DisksPathFP.Relative()
	c := s.DisksPathFP.Base()
	d := s.DisksPathFP.Get()

	log.Printf("Paths VM: Abs:%s,Relative:%s, Base:%s, Get:%s\n", a, b, c, d)

	if utils.PathResolvable(a) {
		log.Printf("Abs Path Resolvable: %s\n", a)
	}

	if utils.PathResolvable(b) {
		log.Printf("Relative Path Resolvable: %s\n", a)
	}
	if utils.PathResolvable(c) {
		log.Printf("Base Path Resolvable: %s\n", a)
	}

	if utils.PathResolvable(d) {
		log.Printf("Get Path Resolvable: %s\n", a)
	}

	log.Printf("VM Disks Path: %s\n", diskPath)

	for _, disk := range s.Disks {

		diskPathQemu, err := disk.DiskPathFP.Relative()
		absPath := disk.DiskPathFP.Abs()
		base := disk.DiskPathFP.Base()

		if utils.PathResolvable(diskPathQemu) {
			log.Printf("InnerRelative Path Resolvable: %s\n", a)
		}

		if utils.PathResolvable(absPath) {
			log.Printf("InnerAbs Path Resolvable: %s\n", a)
		}
		if utils.PathResolvable(base) {
			log.Printf("InnerBase Path Resolvable: %s\n", a)
		}

		log.Printf("Relative Path returned for disk creation:%s", diskPathQemu)

		log.Printf("Paths: Relative:%s , Abs:%s, Base:%s\n", diskPathQemu, absPath, base)

		/*
					   Wrong : Relative Path returned for disk creation:../../../images/data/artifacts/worker/worker-openebs-disk.qcow2
					   Running qemu-img to create disk : qemu-img create -f qcow2 ../../../images/data/artifacts/worker/worker-openebs-disk.qcow2 10G
			Relative Path returned for disk creation:../control-openebs-disk.qcow2
			2024/07/03 01:57:49 image.go:158: Running q
				fpath.LogCwd()
		*/
		if err != nil {
			log.Fatalf("Failed to Get Relative Disk Path for QEMU Create. ERROR:%s", err)
		}

		if err := utils.CreateDiskQCow(diskPathQemu, disk.Size); err != nil {
			log.Printf(utils.TurnError(fmt.Sprintf("Failed to Create Disk for VM. ERROR:%s,", err)))
			return err
		}
	}

	_ = s.navigateToRoot()

	return nil
}

/*
Uses virt-customize to truncate the Cloud Image
Patches the Hostname FQDN not being set during Boot

	sudo virt-customize -a myvm-disk.qcow2 --truncate /etc/machine-id

See: https://bugs.launchpad.net/cloud-init/+bug/1739516
*/
func (s *VMConfig) ResolveFQDNBootBehaviorImg() error {
	log.Print("Creating Base Image")

	_ = s.navigateToDirWithISOImages()

	if err := exec.Command(
		"sudo",
		"virt-customize",
		"-a",
		utils.ModifiedImageName(s.VMName),
		"--truncate",
		"/etc/machine-id",
	).Run(); err != nil {
		log.Printf("Error creating directory: %v", err)
		return err
	}

	return nil
}

// SetupVM() creates a Mount Path to Copy Boot scripts into the VM,
// Curr main logic - uses primary disk and mounts at /mnt/vmname
// Copies Dynamic Data into the VM, and then clears the Mount Data.
// Uses the generated base image in data/images/control-vm-disk.qcow2
func (s *VMConfig) SetupVM() error {
	utils.LogStep("MOUNTING IMAGE")

	if err := s.navigateToDirWithISOImages(); err != nil {
		errStr := utils.TurnError(fmt.Sprintf("Failed to navigate to Dir with VM Base Image (central repo) to modify the VM Base Image"))
		log.Println(errStr)
		return errors.New(errStr)

	}

	modifiedImagePath := filepath.Join(s.VMName + "-vm-disk.qcow2")
	log.Printf("modified Image Path %s", modifiedImagePath)
	mountPath := "/mnt/" + s.VMName

	log.Printf("Mount Path Setup VM %s", mountPath)

	// exit early - in case no files defined to copy into the Disk
	if s.BootFilesDir == "" && s.SystemdScript == "" {
		log.Println("No files required to copy into primary VM image, proceeding with user-data.img creation followed by virt-install.")
		return nil
	}
	// Mount vm image
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

	log.Printf("Unmounting Image and Clearing Temp Mount Path %s", mountPath)

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

// CreateVM() runs virt-install - needs paths for the VM Image and the user-data.img files
// uses libvirtd to create the VM and boot it.
// The state will change to Running and the boot scripts will run followed by systemd services
// Uses the image from data/images/control-vm-disk.qcow2
// Adds any extra disks defined on the Struct
func (s *VMConfig) CreateVM() error {
	err := s.navigateToRoot()
	if err != nil {
		log.Printf("Failed to Navigate to Root Dir. Virt-install must be ran with relative pathing. :%s", err)
	}

	// Currently vm images generated in data/images/<vmaame> itself
	generatedVmImg := filepath.Join(s.ImagesDir, s.VMName+"-vm-disk.qcow2")
	vm_userdata_img := filepath.Join("data", "artifacts", s.VMName, "userdata", "user-data.img")

	cmdArgs := []string{
		"--name", s.VMName,
		"--virt-type", "kvm",
		"--memory", fmt.Sprint(s.Memory),
		"--vcpus", fmt.Sprint(s.CPUCores),
		"--disk", "path=" + generatedVmImg + ",device=disk",
		"--disk", "path=" + vm_userdata_img + ",format=raw",
		"--graphics", "none",
		"--boot", "hd,menu=on",
		"--network", "network=default",
		"--os-variant", "ubuntu18.04",
		"--noautoconsole",
	}

	// Dynamically add disks to the command

	// Add Disks to the VM
	// Using FilePath type to get the Relative Command Path
	for _, diskPath := range s.Disks {
		relativePath, err := diskPath.DiskPathFP.Relative()
		if err != nil {
			log.Fatalf("ERROR:%s", err)
		}
		fmt.Println(utils.TurnBoldBlueDelimited(relativePath))

		cmdArgs = append(cmdArgs, "--disk", "path="+relativePath+",device=disk")

	}

	// for _, diskPath := range s.GetDiskPaths() {
	// 	cmdArgs = append(cmdArgs, "--disk", "path="+diskPath+",device=disk")
	// }

	log.Printf("%sCreating Virtual Machine%s %s%s%s%s:\nvirt-install %s\n", utils.BOLD, utils.NC, utils.BOLD, utils.COOLBLUE, s.VMName, utils.NC, strings.Join(cmdArgs, " "))

	/*
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
	*/

	cmd := exec.Command("virt-install", cmdArgs...)

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
	// local := filepath.Join(s.artifactPath, s.VMName)

	// if err := utils.CreateDirIfNotExist(local); err != nil {
	if err := utils.CreateDirIfNotExist(s.ArtifactPath); err != nil {
		log.Printf("Failed 	utils.CreateDirIfNotExist(local) ERROR:%s,", err)
		return err
	}

	cmd := exec.Command("sudo", "virt-copy-out", "-d", s.VMName, path, s.ArtifactPath)

	log.Printf("Running sudo virt-copy-out -d %s %s %s", s.VMName, path, s.ArtifactPath)

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

	client, err := network.NewInsecureSSHClientVM(s.VMName, ip.StringWithSubnet(), username, password)
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
		log.Printf(utils.TurnError(fmt.Sprintf("Failed to change directory: %v", err)))
		return err
	}
	return nil
}

/* Naviagate to a Path */
func (s *VMConfig) navigateToAbsPath(absPath string) error {
	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("Failed to Navigate to %s. ERROR:%s", absPath, err.Error())
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

// func (s *VMConfig) SetArtifactDir(path string) {
// 	s.artifactPath = path
// }

func (config *VMConfig) GetImageUserDataPath() (string, error) {
	// outputImgPath := filepath.Join(config.artifactPath, config.VMName, "userdata", "user-data.img")
	outputImgPath := filepath.Join(config.ArtifactPath, "userdata", "user-data.img")
	absoluteOutputImgPath, err := filepath.Abs(outputImgPath)
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
		return "", err
	}
	return absoluteOutputImgPath, nil
}

func (config *VMConfig) CreateUserDataDir() error {
	// userDataDir := filepath.Join(config.artifactPath, config.VMName, "userdata")
	userDataDir := filepath.Join(config.ArtifactPath, "userdata")

	if err := utils.CreateDirIfNotExist(userDataDir); err != nil {
		log.Printf("Error creating user data directory: %v", err)
		return err
	}

	return nil
}
