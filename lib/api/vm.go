package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"kvmgo/lib"
	"kvmgo/types/fpath"
	"kvmgo/utils"
)

type VM struct {
	Name           string // redpanda
	StoragePath    string // /home/kuro/kvm/rpanda
	baseImgName    string // <vm_name>-base.img
	backingImgPath string // /path/to/ubuntu-24.04.img

	client *lib.VirtClient
	pool   *lib.Pool

	images map[string]*lib.Volume

	config *lib.VMConfig

	baseImgInitialized bool
	userdataPath       string
	baseimgPath        string
}

// NewVM returns the Img Manager with a default Pool with the name provided
// The Path provided will be the path of the Storage Pool
// The path defined here must be entirely managed by libvirt - i.e do not create any dirs/files manually within the Storage Pool
// @Usage
// NewVM("ubuntu","/home/kuro/testubuntu") - creates /testubuntu/tmp and /testubuntu/images
// Note: To create and launch a VM - only a valid path to an Image File must be provided
func NewVM(name, path string) (*VM, error) {
	log.Printf("VM for %s at %s\n", name, path)
	conn, err := lib.ConnectLibvirt()
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}

	// Init the VmConfig with the Name
	vm := &VM{Name: name, StoragePath: path, client: conn, config: lib.NewVMConfig(name), images: map[string]*lib.Volume{}}

	pool, err := conn.GetOrCreatePool(name, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Storage Pool for VM. Error:%s\n", err)
	}

	vm.pool = lib.InitPool(pool, name, path)

	return vm, nil
}

// Launch the VM - once a base image is set. Apply defaults to uninitialized VmConfig params
// Userdata and Base Image must be set
func (vm *VM) Launch() error {
	if !vm.baseImgInitialized || vm.userdataPath == "" || vm.baseimgPath == "" {
		return fmt.Errorf("VM is not initialized with a Base Image and User Data")
	}
	vm.config.SetBaseImage(vm.BaseImagePath()) // set path to the base image
	return nil
}

// baseImageName will be the name of the Volume/Image created from the base OS
// <vm.Name>-base.img
func (vm *VM) BaseImageName() string {
	if vm.baseImgName == "" {
		vm.baseImgName = utils.AddExtensionIfRequired(vm.Name+"-base", ".img")
	}
	return vm.baseImgName
}

// Get the Image/Volume Path associated with the Base Image
// Adds Use this to configure the VM.
func (vm *VM) BaseImagePath() string {
	if vm.baseimgPath == "" {
		vm.baseimgPath = filepath.Join(vm.StoragePath, vm.BaseImageName())
	}
	return vm.baseimgPath
}

// CreateBaseImage will use the vm name to generate a default <vm-name>-base.img file as the base backing image
// The path of the VM's image is the Storage Pool Path + vm.BaseImageName()
func (vm *VM) CreateBaseImage(imgPath string, capacityGB int) error {
	if vm.baseImgInitialized {
		return nil // already setup
	}
	if imgPath == "" {
		if vm.backingImgPath == "" {
			return fmt.Errorf("Either a backing image must be set or a path to a base image must be provided")
		}
		imgPath = vm.backingImgPath
	}

	vmBaseImageName := vm.BaseImageName() // <vm.Name>-base.img
	if vm.pool.ImageExists(vmBaseImageName) {
		log.Printf("Image %s already exists\n", vmBaseImageName)
		vm.baseImgInitialized = true
		return nil
	}
	log.Printf("Creating VM Base Image at %s from Backing Image %s with capacity %d.\n", vm.BaseImagePath(), imgPath, capacityGB)
	return vm.CreateImage(imgPath, vmBaseImageName, capacityGB)
}

// CreateImage creates a new Image for this VM from a backing image such as an Ubuntu base Image
// The imageName defines the name of the image created where the VM's Storage Pool is configured
//
//	vm.CreateImage("/ubuntu-22.04.img","kafka-base",20) note can also pass kafka-vm.img - the suffix is added if not present
func (vm *VM) CreateImage(baseImgPath, imageName string, capacityGB int) error {
	if !utils.FileExists(baseImgPath) {
		return fmt.Errorf("image file %s does not exist", baseImgPath)
	}

	xml := getXMLFromBaseImage(baseImgPath, imageName, capacityGB)
	if err := vm.pool.CreateImageXML(xml); err != nil {
		return fmt.Errorf("failed to create volume: %v", err)
	}
	vm.baseImgInitialized = true

	return nil
}

// Define the XML for the new volume
// @Usage:
// getXMLFromBaseImage("/path/v.img", "kafka", 20) // or "kafka.img" - it is added if not specified
func getXMLFromBaseImage(baseImagePath, newName string, capacityGB int) string {
	return fmt.Sprintf(`
<volume>
	<name>%s</name>
	<allocation>0</allocation>
	<capacity unit="G">%d</capacity>
	<target>
		<format type='qcow2'/>
	</target>
	<backingStore>
		<path>%s</path>
		<format type='qcow2'/>
	</backingStore>
</volume>`, newName, capacityGB, baseImagePath)
}

///////////////////////

// Get and Set the config for the VM
func (vm *VM) GetConfig() *lib.VMConfig {
	return vm.config
}

func (vm *VM) SetConfig(config *lib.VMConfig) {
	vm.config = config
}

// Set the backing image used by the VM
func (vm *VM) SetBackingImage(path string) {
	vm.backingImgPath = path
}

// Set the userdata.img used by the VM
func (vm *VM) SetUserData(path string) {
	vm.userdataPath = path
}

// /////////////////////
func (vm *VM) ListImages() error {
	if vm.images == nil {
		imgs, err := vm.pool.GetImages()
		if err != nil {
			return fmt.Errorf("failed to get images. Err:%s")
		}
		vm.images = imgs
	}

	vols, err := vm.pool.GetVolumes(false)
	if err != nil {
		return fmt.Errorf("Failed to get Volumes/Images Error:%s", err)
	}

	for _, vol := range vols {
		fmt.Printf("Volume/Image: %s\n", vol)
	}

	return nil
}

// path will be path + tmp
func (vm *VM) tempPath() string {
	return filepath.Join(vm.StoragePath, "tmp")
}

// Images path is path + images
func (vm *VM) Images() string {
	return filepath.Join(vm.StoragePath, "images")
}

func (vm *VM) GetImage(name string) (string, error) {
	path, err := vm.pool.GetVolume(name)
	if err != nil {
		return "", fmt.Errorf("Image does not exist. Err:%s")
	}

	return path, nil
}

// Fetch the Base OS Image
func (vm *VM) AddImageHttp(url, name string) (string, error) {
	path, err := FetchImageUrl(url, vm.tempPath())
	if err != nil {
		return "", fmt.Errorf("Failed to pull image: %s\n", err)
	}

	log.Printf("Successfully pulled image to %s\n", path)

	return path, nil
}

// Creates images and tmp dir for the VM under its' storage pool path
func (vm *VM) initPath(path string) error {
	fpath := fpath.SecurePath(path)

	imgs := filepath.Join(fpath.Abs(), "images")
	tmp := filepath.Join(fpath.Abs(), "tmp")

	if err := os.MkdirAll(imgs, 0o755); err != nil {
		log.Printf("Failed to create folder: %v", err)
		return err
	}

	if err := os.MkdirAll(tmp, 0o755); err != nil {
		log.Printf("Failed to create folder: %v", err)
		return err
	}

	vm.StoragePath = path
	return nil
}
