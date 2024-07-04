package lib

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"kvmgo/types/fpath"

	"libvirt.org/go/libvirt"
)

// ImgManager "ubuntu"
// Volumes (Images) will be created at the PoolPath we specify

// Deleting a volume : virsh vol-delete myvolume --pool default

// Either each ImageManager will have its' own Pool
// Or a Single Pool for OS' like Ubuntu

type ImageManager struct {
	name string
	path string

	client *VirtClient

	// "/var/lib/libvirt/images/base"
	poolPath string
	images   map[string]string

	// ubuntu
	baseImg string
	// only need to define 1st time, or never - for efficiency store it as a field on the struct so we always
	// dont need to call libvirt to get it
	basePool string
}

// NewImageMgr returns the Img Manager with a default Pool
func NewImageMgr(name, path string) (*ImageManager, error) {
	client, err := ConnectLibvirt()
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}

	imgMgr := &ImageManager{name: name, client: client, images: make(map[string]string)}

	if err := imgMgr.initDirs(); err != nil {
		return nil, fmt.Errorf("Could not generate startup dirs")
	}

	log.Printf("creating storage pool")

	if err := imgMgr.CreateStoragePool(name, imgMgr.poolPath); err != nil {
		return nil, err
	}

	return imgMgr, nil
}

// image manager has basePath - images are stored here
func (im *ImageManager) BasePath() string {
	return fmt.Sprintf("/home/kuro/kvm/images/%s/base/", im.name)
}

// image manager has basePath - images are stored here
func (im *ImageManager) BasePool() string {
	return fmt.Sprintf("/home/kuro/kvm/images/%s/pools/", im.name)
}

// AddImage will add an Image
func (im *ImageManager) AddImage(url, imgName string) error {
	if err := PullImage(url, im.BasePath()); err != nil {
		return fmt.Errorf("failed to pull image, %s\n", err)
	}

	log.Println("Pull Image done")

	im.images[imgName] = im.BasePath() + imgName

	log.Println("Set to Images Image done")

	return nil
}

// AddImage will add an Image from Base Image.
// Base Image should have the name specified by the AddImage(url,name) call
func (im *ImageManager) CreateImageFromBase(baseImg, newImg string, capacityGB int) error {
	baseImgPath, err := im.GetImage(baseImg)
	if err != nil {
		return fmt.Errorf("Base img does not exist")
	}

	if err := im.CreateImgVolume(im.name, newImg, baseImgPath, capacityGB); err != nil {
		return err
	}

	return nil
}

// AddImage will add an Image from Base Image.
// Base Image should have the name specified by the AddImage(url,name) call
func (im *ImageManager) CreateImgFromBaseName(baseImg, newImg string, capacityGB int) error {
	baseImgPath, err := im.GetImage(baseImg)
	if err != nil {
		return fmt.Errorf("Base img does not exist")
	}

	if err := im.CreateImgVolume(im.name, newImg, baseImgPath, capacityGB); err != nil {
		return err
	}

	// im[newImg]

	return nil
}

// GetImage() returns the Image Path if it exists - or nothing
func (im *ImageManager) GetImage(imgName string) (string, error) {
	img, ok := im.images[imgName]

	if !ok {
		return "", fmt.Errorf("Image %s does not exist.")
	}

	return img, nil
}

// PullImage pulls the URL and saves it directly as a File if dir is a Path or saves to Dir with download name
func PullImage(url, dir string) error {
	if url == "" {
		return fmt.Errorf("passed empty URL")
	}

	var imagePath string
	if filepath.Ext(dir) != "" {
		// If 'dir' has an extension, treat it as a full file path
		imagePath = dir
	} else {
		// Otherwise, treat 'dir' as a directory and append the image name
		imageName := filepath.Base(url)
		imagePath = filepath.Join(dir, imageName)
	}

	pullImgsStr := fmt.Sprintf("Pulling Base Image: URL:%s, Dir:%s, ImgPath: %s\n", url, dir, imagePath)
	log.Println(pullImgsStr)

	if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
		log.Printf("Image %s already exists", filepath.Base(imagePath))
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// CreateStoragePool creates the storage pool if it doesn't exist
func (im *ImageManager) CreateStoragePool(poolName, poolPath string) error {
	// Check if the storage pool already exists

	if err := fpath.CreateDirIfNotExists(poolPath); err != nil {
		return err
	}

	pool, err := im.client.GetStoragePool(poolName)
	if err == nil {
		return nil
	}

	// If the pool does not exist, create it
	poolXML := fmt.Sprintf(`<pool type='dir'>
                                    <name>%s</name>
                                    <target>
                                        <path>%s</path>
                                    </target>
                                </pool>`, poolName, poolPath)

	pool, err = im.client.Conn().StoragePoolCreateXML(poolXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage pool: %v\n", err)
		return err
	}

	defer pool.Free()
	return nil
}

// StoragePoolExists checks if the storage pool exists
func (v *ImageManager) StoragePoolExists(poolName string) bool {
	return v.client.StoragePoolExists(poolName)
}

// CreateImgVolume creates a new image volume in the specified storage pool

// image will belong to a Pool and have a Volume Name

// CreateImgVolume creates the .img Volume at the poolPath specified from the Base Image
func (im *ImageManager) CreateImgVolume(poolName, volumeName, baseImagePath string, capacityGB int) error {
	pool, err := im.client.conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return err
	}

	if err := pool.Create(0); err != nil && err.(libvirt.Error).Code != libvirt.ERR_OPERATION_INVALID {
		fmt.Printf("Failed to activate storage pool: %v\n", err)
		return fmt.Errorf("Storage Pool not active for %s", poolName)
	}

	volXML := fmt.Sprintf(`<volume>
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
               </volume>`, volumeName, capacityGB, baseImagePath)

	// Check if the volume already exists
	vol, err := pool.LookupStorageVolByName(volumeName)
	if err == nil {
		defer vol.Free()
		volPath, err := vol.GetPath()
		if err != nil {
			return fmt.Errorf("failed to get the path of the existing volume: %v", err)
		}
		fmt.Printf("Storage volume '%s' already exists at path: %s\n", volumeName, volPath)
		return nil
	}

	vol, err = pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return err
	}
	defer vol.Free()

	return nil
}

func (im *ImageManager) DeleteImgVolume(volumeName string) error {
	pool, err := im.client.conn.LookupStoragePoolByName(im.name)
	if err != nil {
		return fmt.Errorf("failed to look up storage pool by name: %v", err)
	}

	vol, err := pool.LookupStorageVolByName(volumeName)
	if err != nil {
		return fmt.Errorf("failed to look up storage volume by name: %v", err)
	}

	defer vol.Free()

	volPath, err := vol.GetPath()
	if err != nil {
		return fmt.Errorf("failed to get the path of the volume: %v", err)
	}
	fmt.Printf("Deleting storage volume '%s' at path: %s\n", volumeName, volPath)

	err = vol.Delete(0)
	if err != nil {
		return fmt.Errorf("failed to delete storage volume: %v", err)
	}

	fmt.Printf("Storage volume '%s' deleted successfully\n", volumeName)
	return nil
}

// NewImageManager creates a new ImageManager instance
func NewImageManager(name, path string, client *VirtClient) *ImageManager {
	return &ImageManager{
		name:   name,
		path:   path,
		client: client,
	}
}

// CreateBaseImageStoragePool creates the storage pool for base images
func (im *ImageManager) CreateBaseImageStoragePool() error {
	return im.client.CreateStoragePool(im.name, im.path)
}

// BaseImagePath returns the path where base images are stored
func (im *ImageManager) BaseImagePath() string {
	return fmt.Sprintf("%s/base", im.path)
}

// GeneratedImagePath returns the path where generated images are stored
func (im *ImageManager) GeneratedImagePath() string {
	return fmt.Sprintf("%s/generated", im.path)
}

func (im *ImageManager) initDirs() error {
	defaultPool := im.BasePool()
	basePath := im.BasePath()

	if err := fpath.CreateDirIfNotExists(basePath); err != nil {
		log.Fatalf("BASE IMGS FAILURE")
		return fmt.Errorf("failed to create base imgs path %s Error:%s", basePath, err)
	}

	if err := fpath.CreateDirIfNotExists(defaultPool); err != nil {
		log.Fatalf("BASE POOL FAILURE")
		return fmt.Errorf("failed to create base imgs path %s Error:%s", defaultPool, err)
	}

	im.path = basePath
	im.poolPath = defaultPool

	return nil
}

// CreateImage uses the Base Image w/ Pool to create the secondary .img Volume at the poolPath specified from the Base Image
func (im *ImageManager) CreateImage(poolName, volumeName, baseImagePoolName, baseImageVolumeName string, capacityGB int) error {
	pool, err := im.client.conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("failed to look up storage pool by name: %v", err)
	}

	if err := pool.Create(0); err != nil && err.(libvirt.Error).Code != libvirt.ERR_OPERATION_INVALID {
		fmt.Printf("Failed to activate storage pool: %v\n", err)
		return fmt.Errorf("Storage Pool not active for %s", poolName)
	}

	// ex: ubuntu Look up the base image in the specified pool
	baseImagePool, err := im.client.conn.LookupStoragePoolByName(baseImagePoolName)
	if err != nil {
		return fmt.Errorf("failed to look up base image pool by name: %v", err)
	}

	// ex: ubuntu-22.04.img or whatever we named it
	// good idea to use a Hash of the URL as Deterministic Volume Names
	baseImageVol, err := baseImagePool.LookupStorageVolByName(baseImageVolumeName)
	if err != nil {
		return fmt.Errorf("failed to look up base image volume by name: %v", err)
	}
	defer baseImageVol.Free()

	baseImagePath, err := baseImageVol.GetPath()
	if err != nil {
		return fmt.Errorf("failed to get the path of the base image volume: %v", err)
	}

	volXML := fmt.Sprintf(`<volume>
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
               </volume>`, volumeName, capacityGB, baseImagePath)

	// Check if the volume already exists
	vol, err := pool.LookupStorageVolByName(volumeName)
	if err == nil {
		defer vol.Free()
		volPath, err := vol.GetPath()
		if err != nil {
			return fmt.Errorf("failed to get the path of the existing volume: %v", err)
		}
		fmt.Printf("Storage volume '%s' already exists at path: %s\n", volumeName, volPath)
		return nil
	}

	vol, err = pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return err
	}
	defer vol.Free()

	volPath, err := vol.GetPath()
	if err != nil {
		return fmt.Errorf("failed to get the path of the newly created volume: %v", err)
	}
	fmt.Printf("New storage volume '%s' created at path: %s\n", volumeName, volPath)

	return nil
}
