package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"kvmgo/lib"
	"kvmgo/types/fpath"

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

// NewImageMgr returns the Img Manager with a default Pool
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

// path will be path + images
func (vm *VM) tempPath() string {
	return filepath.Join(vm.path, "tmp")
}

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
	if err := os.MkdirAll(imgs, os.ModePerm); err != nil {
		log.Printf("Failed to create folder: %v", err)
		return err
	}
	vm.path = path
	return nil
}
