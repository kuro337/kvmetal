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
	name   string
	path   string
	client *lib.VirtClient

	// "/var/lib/libvirt/images/base"
	poolPath string
	pool     *lib.Pool

	// basePath *fpath.FilePath
	images map[string]*libvirt.StorageVol
}

// NewImageMgr returns the Img Manager with a default Pool with the name provided
// The storage pool is created at the path provided - and a /tmp and /images dir is created at the Path for the VM
// @Usage
// NewVM("ubuntu","/home/kuro/testubuntu") - creates /testubuntu/tmp and /testubuntu/images
func NewVM(name, path string) (*VM, error) {
	log.Printf("VM for %s at %s\n", name, path)
	conn, err := lib.ConnectLibvirt()
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}
	vm := &VM{name: name, client: conn, images: map[string]*libvirt.StorageVol{}}

	if err := vm.initPath(path); err != nil {
		return nil, err
	}

	pool, err := conn.GetOrCreatePool(name, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Storage Pool for VM. Error:%s\n", err)
	}

	vm.pool = lib.InitPool(pool, name, path)

	return vm, nil
}

// CreateBaseImage will use the vm name to generate a default <vm-name>-base.img file as the base backing image
func (vm *VM) CreateBaseImage(imgPath string, capacityGB int) error {
	return vm.CreateImage(imgPath, vm.name+"-base", capacityGB)
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
	// make sure it has an .img extension if not provided
	newName = utils.AddExtensionIfRequired(newName, ".img")

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

	return nil
}

// path will be path + tmp
func (vm *VM) tempPath() string {
	return filepath.Join(vm.path, "tmp")
}

// Images path is path + images
func (vm *VM) Images() string {
	return filepath.Join(vm.path, "images")
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

// NewVM done - now list all images for it
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

	vm.path = path
	return nil
}
