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

var baseOsImgDir = "/kvm/images"

type ImageManager struct {
	name string
	path string

	client *VirtClient

	// "/var/lib/libvirt/images/base"
	poolPath string
	images   map[string]string
}

// NewImageMgr returns the Img Manager with a default Pool
func NewImageMgr(name, path string) (*ImageManager, error) {
	client, err := ConnectLibvirt()
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}

	log.Printf("Pre New")
	imgMgr := &ImageManager{name: name, client: client, images: make(map[string]string)}

	path = imgMgr.BasePath()
	log.Printf("got basePath: %s\n", path)

	imgMgr.path = path

	defaultPool := imgMgr.BasePool()

	log.Printf("default pool: %s\n", defaultPool)

	log.Printf("creating storage pool")

	if err := imgMgr.CreateStoragePool(name, defaultPool); err == nil {
		return nil, err
	}

	return imgMgr, nil
}

// image manager has basePath - images are stored here
func (im *ImageManager) BasePath() string {
	return fmt.Sprintf("/var/lib/libvirt/images/%s/base", im.name)
}

// image manager has basePath - images are stored here
func (im *ImageManager) BasePool() string {
	return fmt.Sprintf("/var/lib/libvirt/images/%s/pools", im.name)
}

// AddImage will add an Image
func (im *ImageManager) AddImage(url, imgName string) error {
	if err := PullImage(url, im.BasePath()); err != nil {
		return fmt.Errorf("failed to pull image, %s\n", err)
	}
	im.images[imgName] = im.BasePath() + imgName
	return nil
}

// AddImage will add an Image
func (im *ImageManager) CreateImageFromBase(baseImg, newImg string, capacitGB int) error {
	baseImgPath, err := im.GetImage(baseImg)
	if err != nil {
		return fmt.Errorf("Base img does not exist")
	}
	if err := im.CreateImgVolume(im.name, newImg, baseImgPath, capacitGB); err != nil {
		return err
	}

	return nil
}

// GetImage() returns the Image Path if it exists - or nothing
func (im *ImageManager) GetImage(imgName string) (string, error) {
	img, ok := im.images[imgName]

	if !ok {
		return "", fmt.Errorf("Image %s does not exist")
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

	vol, err := pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return err
	}
	defer vol.Free()

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

// Images stored in "/var/lib/libvirt/images/base"
type ImagePool struct {
	name string
	path string
}

/*
func main() {
	// Example usage
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		fmt.Printf("Failed to connect to hypervisor: %v\n", err)
		return
	}
	defer conn.Close()


	client := &VirtClient{conn: conn}
	imageManager := NewImageManager("default", "/var/lib/libvirt/images", client)


	err = imageManager.CreateBaseImageStoragePool()
	if err != nil {
		fmt.Printf("Failed to create base image storage pool: %v\n", err)
		return
	}


	err = imageManager.CreateGeneratedImage("new_img.qcow2", 20)
	if err != nil {
		fmt.Printf("Failed to create generated image: %v\n", err)
		return
	}


	fmt.Println("Base images are stored at:", imageManager.BaseImagePath())
	fmt.Println("Generated images are stored at:", imageManager.GeneratedImagePath())
}

*/

/// BUt your above example makes no sense -