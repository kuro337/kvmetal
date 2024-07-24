package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"kvmgo/lib"
	"kvmgo/types/fpath"
	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

type VM struct {
	Name   string
	Path   string
	client *lib.VirtClient

	// "/var/lib/libvirt/images/base"
	pool *lib.Pool

	// basePath *fpath.FilePath
	images map[string]*libvirt.StorageVol
}

// NewImageMgr returns the Img Manager with a default Pool with the name provided
// The storage pool is created at the path provided - and a /storage , /tmp and /images dir is created at the Path for the VM
// /storage is the libvirt managed Storage Pool
// The path defined here must be entirely managed by libvirt - i.e do not create any dirs/files manually within the Storage Pool
// @Usage
// NewVM("ubuntu","/home/kuro/testubuntu") - creates /testubuntu/tmp and /testubuntu/images
func NewVM(name, path string) (*VM, error) {
	log.Printf("VM for %s at %s\n", name, path)
	conn, err := lib.ConnectLibvirt()
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}

	vm := &VM{Name: name, Path: path, client: conn, images: map[string]*libvirt.StorageVol{}}

	// note: do not create dirs where we register the Storage Pool - it has to be managed by libvirt
	//	if err := vm.initPath(path); err != nil {
	//		return nil, err
	//	}

	pool, err := conn.GetOrCreatePool(name, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Storage Pool for VM. Error:%s\n", err)
	}

	vm.pool = lib.InitPool(pool, name, path)

	return vm, nil
}

// baseImageName will be the name of the Volume/Image created from the base OS
// <vm.Name>-base.img
func (vm *VM) baseImageName() string {
	// make sure it has an .img extension if not provided
	return utils.AddExtensionIfRequired(vm.Name+"-base", ".img")
}

// CreateBaseImage will use the vm name to generate a default <vm-name>-base.img file as the base backing image
func (vm *VM) CreateBaseImage(imgPath string, capacityGB int) error {
	baseImageName := vm.baseImageName()
	// <vm.Name>-base.img
	if vm.pool.ImageExists(baseImageName) {
		return nil // already exists
	} else {
		fmt.Printf("Pool %s does not exist, creating.\n", baseImageName)
	}

	return vm.CreateImage(imgPath, baseImageName, capacityGB)
}

// CreateImage creates a new Image for this VM from a backing image such as an Ubuntu base Image
//
//	vm.CreateImage("/ubuntu-22.04.img","kafka-base",20) note can also pass kafka-vm.img - the suffix is added if not present
func (vm *VM) CreateImage(imgPath, imageName string, capacityGB int) error {
	xml := getXMLFromBaseImage(imgPath, imageName, capacityGB)
	if err := vm.pool.CreateImageXML(xml); err != nil {
		return fmt.Errorf("failed to create volume: %v", err)
	}
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
	return filepath.Join(vm.Path, "tmp")
}

// Images path is path + images
func (vm *VM) Images() string {
	return filepath.Join(vm.Path, "images")
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

	vm.Path = path
	return nil
}
