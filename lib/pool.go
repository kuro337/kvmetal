package lib

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"libvirt.org/go/libvirt"
)

type Pool struct {
	name string
	path string
	pool *libvirt.StoragePool

	//	volumes map[string]*libvirt.StorageVol
	volumes  map[string]*Volume
	preserve bool
}

func GetPool(conn *libvirt.Connect, poolName string) (*Pool, error) {
	// Check if the pool exists
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code == libvirt.ERR_NO_STORAGE_POOL {
			return nil, fmt.Errorf("storage pool %s does not exist", poolName)
		}
		return nil, fmt.Errorf("failed to lookup storage pool: %v", err)
	}

	// Get the XML description of the pool to extract the path
	xmlDesc, err := pool.GetXMLDesc(0)
	if err != nil {
		pool.Free()
		return nil, fmt.Errorf("failed to get pool XML description: %v", err)
	}

	pathStart := strings.Index(xmlDesc, "<path>")
	pathEnd := strings.Index(xmlDesc, "</path>")
	if pathStart == -1 || pathEnd == -1 {
		pool.Free()
		return nil, fmt.Errorf("path not found in pool XML")
	}
	path := xmlDesc[pathStart+6 : pathEnd]

	return &Pool{
		name:    poolName,
		path:    path,
		pool:    pool,
		volumes: map[string]*Volume{},
	}, nil
}

// Delete deletes a Pool entirely clearing all the volumes associated with it
func (p *Pool) Delete() error {
	if p.pool == nil {
		return fmt.Errorf("storage pool is nil")
	}

	if err := p.UpdateVolumes(); err != nil {
		return fmt.Errorf("failed updateVolumes:%s\n", err)
	}
	// delete volumes pulled from UpdateVolumes()
	if err := p.DeleteVolumes(); err != nil {
		return fmt.Errorf("failed delete volumes:%s\n", err)
	}

	// Destroy the pool if it is active
	if err := p.destroy(); err != nil {
		return err
	}

	// Undefine an Inactive Storage Pool - call after destroy
	if err := p.pool.Undefine(); err != nil {
		return fmt.Errorf("failed undefine pool: %v", err)
	}

	// Delete the pool - final irreversible operation
	if err := p.pool.Delete(libvirt.STORAGE_POOL_DELETE_NORMAL); err != nil {
		return fmt.Errorf("failed to delete pool data: %v", err)
	}

	return nil
}

func (p *Pool) DeleteImage(image string) error {
	if ex, _ := ImageExists(p.pool, image); ex {
		log.Printf("Volume %s already exists", image)
		if err := DeleteImage(p.pool, image); err != nil {
			log.Fatalf("failed to delete volume: %s\n", err)
		}
		log.Printf("Deleted volume %s", image)
	}

	return nil
}

func InitPool(pool *libvirt.StoragePool, name, path string) *Pool {
	return &Pool{
		name:    name,
		path:    path,
		pool:    pool,
		volumes: map[string]*Volume{},
	}
}

// NewPool creates and returns a new storage pool
func NewPool(conn *libvirt.Connect, name, path string) (*Pool, error) {
	poolXML := fmt.Sprintf(`<pool type='dir'>
                                <name>%s</name>
                                <target>
                                    <path>%s</path>
                                </target>
                            </pool>`, name, path)

	pool, err := conn.StoragePoolDefineXML(poolXML, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to define storage pool: %v", err)
	}

	log.Printf("Creating pool %s", name)
	err = pool.Create(libvirt.STORAGE_POOL_CREATE_WITH_BUILD)
	if err != nil {
		pool.Undefine() // clean up the defined pool if creation fails
		return nil, fmt.Errorf("failed to create storage pool: %v", err)
	}

	log.Printf("Setting autostart for pool %s", name)
	err = pool.SetAutostart(true)
	if err != nil {
		pool.Destroy()
		pool.Undefine()
		return nil, fmt.Errorf("failed to set autostart for storage pool: %v", err)
	}

	log.Printf("Pool %s created and autostart set", name)
	return &Pool{
		name:    name,
		path:    path,
		pool:    pool,
		volumes: map[string]*Volume{},
	}, nil
}

func (p *Pool) ImageExists(imageName string) bool {
	ex, err := ImageExists(p.pool, imageName)
	if err != nil || !ex {
		return false
	}
	return true
}

func ImageExists(pool *libvirt.StoragePool, volumeName string) (bool, error) {
	vol, err := pool.LookupStorageVolByName(volumeName)
	if err != nil {
		if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code == libvirt.ERR_NO_STORAGE_VOL {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if volume exists: %v", err)
	}
	defer vol.Free()
	return true, nil
}

func DeleteImage(pool *libvirt.StoragePool, volumeName string) error {
	vol, err := pool.LookupStorageVolByName(volumeName)
	if err != nil {
		if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code == libvirt.ERR_NO_STORAGE_VOL {
			return fmt.Errorf("storage volume %s does not exist", volumeName)
		}
		return fmt.Errorf("failed to look up storage volume by name: %v", err)
	}
	defer vol.Free()

	log.Printf("Deleting volume %s", volumeName)
	if err := vol.Delete(0); err != nil {
		return fmt.Errorf("failed to delete storage volume: %v", err)
	}

	log.Printf("Deleted volume %s", volumeName)
	return nil
}

func PoolExists(conn *libvirt.Connect, poolName string) (bool, error) {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		if err.(libvirt.Error).Code == libvirt.ERR_NO_STORAGE_POOL {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if storage pool exists: %v", err)
	}
	defer pool.Free()
	return true, nil
}

func DeletePool(conn *libvirt.Connect, poolName string) error {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code == libvirt.ERR_NO_STORAGE_POOL {
			return fmt.Errorf("storage pool %s does not exist", poolName)
		}
		return fmt.Errorf("failed to look up storage pool by name: %v", err)
	}
	defer pool.Free()

	log.Printf("Destroying pool %s", poolName)

	if err := pool.Destroy(); err != nil {
		if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code != libvirt.ERR_OPERATION_INVALID {
			return fmt.Errorf("failed to destroy storage pool: %v", err)
		}
	}
	log.Printf("Undefining pool %s", poolName)
	if err := pool.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine storage pool: %v", err)
	}
	log.Printf("Deleted pool %s", poolName)
	return nil
}

// destroy destroys an active Storage Pool. Destroy after deleting the Volumes
func (p *Pool) destroy() error {
	if active, err := p.Active(); err == nil && active {
		if err := p.pool.Destroy(); err != nil {
			if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code != libvirt.ERR_OPERATION_INVALID {
				return fmt.Errorf("failed to destroy storage pool: %v", err)
			}
		}
	}
	return nil
}

// Active() checks whether the Pool is Active or not
func (p *Pool) Active() (bool, error) {
	pool := p.pool

	active, err := pool.IsActive()
	if err != nil {
		return false, fmt.Errorf("error checking IsActive:%s", err)
	}
	return active, nil

	//	if err := pool.Create(0); err != nil && err.(libvirt.Error).Code != libvirt.ERR_OPERATION_INVALID {
	//		fmt.Printf("Failed to activate storage pool: %v\n", err)
	//		return false, fmt.Errorf("Storage Pool not active for %s", p.name)
	//	}
	//
	// return true, nil
}

func (p *Pool) Refresh() error {
	return p.pool.Refresh(0)
}

// UpdateVolumes refreshes the Pool and gets the Volumes/Images associated with it and resolves the Name and Path
// Updates the pool struct to have the Path as the key for the Volumes to &Volume pointers
func (p *Pool) UpdateVolumes() error {
	pool := p.pool
	// Refresh the pool to get the latest state
	if err := p.Refresh(); err != nil {
		return fmt.Errorf("failed to refresh pool: %v", err)
	}

	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		return fmt.Errorf("failed to list volumes: %v", err)
	}

	// libvirt.StorageVol
	for _, vol := range volumes {
		volume, err := NewVolume(&vol)
		if err != nil {
			vol.Free()
			return fmt.Errorf("failed to get volume:%s", err.Error())
		}

		p.volumes[volume.Path] = volume
	}
	return nil
}

// GetImages updates the current images/volumes and returns the Volumes
func (p *Pool) GetImages() (map[string]*Volume, error) {
	if p.volumes == nil {
		if err := p.UpdateVolumes(); err != nil {
			return nil, err
		}
	}
	return p.volumes, nil
}

// GetVolumes returns the Paths of the Volumes - convenience fn over GetImages() which returns the full Volume
// Refresh whether to pull latest volumes or use existing
func (p *Pool) GetVolumes(refresh bool) ([]string, error) {
	if refresh {
		if err := p.UpdateVolumes(); err != nil {
			return nil, fmt.Errorf("Failed to update volumes, %s\n", err)
		}
		pool := p.pool
		// Refresh the pool to get the latest state
		if err := pool.Refresh(0); err != nil {
			return nil, fmt.Errorf("failed to refresh pool: %v", err)
		}
	}

	var volumePaths []string
	for _, vol := range p.volumes {
		path := vol.Path
		volumePaths = append(volumePaths, path)
		//		vol.Free()
	}
	return volumePaths, nil
}

func (p *Pool) DeleteVolume(path string) error {
	vol, ok := p.volumes[path]
	if !ok {
		return fmt.Errorf("Volume at path %s not found\n", path)
	}
	if err := vol.Delete(); err != nil {
		vol.Free()
		return fmt.Errorf("failed to delete volume: %v", err)
	}
	return nil
}

// Delete the Volumes pulled on the struct
func (p *Pool) DeleteVolumes() error {
	for _, vol := range p.volumes {
		//  fmt.Printf("Deleting Volume: %s\n",vol.GetName()
		if err := vol.Delete(); err != nil {
			vol.Free()
			return fmt.Errorf("failed to delete volume: %v", err)
		}
		// vol.Free()
	}
	return nil
}

func (p *Pool) GetVolume(name string) (string, error) {
	pool := p.pool

	// Check if the volume already exists
	vol, err := pool.LookupStorageVolByName(name)
	if err != nil {
		return "", fmt.Errorf("Volume not found for Pool")
	}

	defer vol.Free()

	volPath, err := vol.GetPath()
	if err != nil {
		return "", fmt.Errorf("failed to get the path of the existing volume: %v", err)
	}
	return volPath, nil
}

func (p *Pool) GetXmlFromUrl(url, volume string, capacityGB int) string {
	return fmt.Sprintf(`<volume>
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
                           </volume>`, volume, capacityGB, url)
}

// ImageConfig returns XML required to generate a new Image/volume from an Existing Image
func (p *Pool) GetXMLFromPath(fromPath, name string, capacityGB int) string {
	log.Printf("Getting XML , passed path: %s , name:%s, cap:%d\n", fromPath, name, capacityGB)
	if strings.Contains(name, "/") {
		log.Fatalf("Invalid volumeName passed: %s\n", name)
	}

	return fmt.Sprintf(`<volume>
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
               </volume>`, name, capacityGB, fromPath)
}

// CreateImage creates a new Volume from the pool's Base Image
func (p *Pool) CreateImagePath(name, fromPath string, capacityGB int) error {
	// xml := p.GetXMLFromPath(name, fromPath, capacityGB)

	ThrowIfInvalid(name)

	xml := p.GetXMLFromPath(fromPath, name, capacityGB)
	return p.CreateImageXML(xml)

	// return p.CreateImageXML(xml)
}

func ThrowIfInvalid(name string) {
	if err := InvalidName(name); err != nil {
		log.Fatalf("Invalid volumeName passed: %s\n", name)
	}
}

func InvalidName(name string) error {
	if strings.Contains(name, "/") {
		return fmt.Errorf("invalid volume name: %s", name)
	}
	return nil
}

func (p *Pool) CreateImageURL(name, url string, capacityGB int) error {
	ThrowIfInvalid(name)
	// Download the image to a temporary directory

	log.Printf("passed : name %s , url:%s\n", name, url)
	err := Downloadfile(url)

	log.Printf("completed download")
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}

	log.Printf("no errors")

	// Use the temporary path in the XML configuration

	tempPath := "/home/kuro/kvm/test/imgfile.img"
	xml := p.GetXMLFromPath(tempPath, name, capacityGB)

	log.Printf("returned xml: from GetXML(): %s\n", xml)
	return p.CreateImageXML(xml)
}

// CreateImageXML generates the XML required to create a new Image/Volume
func (p *Pool) CreateImageXML(xml string) error {
	pool := p.pool

	log.Printf("StorageVolCreateXML()\n")
	vol, err := pool.StorageVolCreateXML(xml, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return err
	}

	log.Println("returning for createImageXML")

	defer vol.Free() // defer should come after the error check
	return nil
}

// DownloadFileTemp downloads a file from the given URL and saves it to the specified temporary directory.
// It returns the path to the downloaded file, a cleanup function to delete the file, and an error if any.
func DownloadFileTemp(url, tempDir string) (string, func(), error) {
	if url == "" {
		return "", nil, errors.New("Invalid URL passed")
	}

	if tempDir == "" {
		return "", nil, errors.New("Invalid tempDir passed")
	}

	log.Printf("Downloading File - in Progress")
	resp, err := http.Get(url)
	log.Printf("Downloading File - DONE")
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %v", err)
	}

	err = os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create directory: %v", err)
	}

	err = os.Chmod(tempDir, 0o755)
	if err != nil {
		return "", nil, fmt.Errorf("failed to set directory permissions: %v", err)
	}

	tempFilePath := tempDir + filepath.Base(url)

	log.Printf("Writing and returning data")
	err = os.WriteFile(tempFilePath, body, 0o644)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write file: %v", err)
	}

	cleanup := func() {
		os.Remove(tempFilePath)
		log.Printf("Deleted temp file: %s", tempFilePath)
	}

	return tempFilePath, cleanup, nil
}

func Downloadfile(url string) error {
	// Create a temporary file

	log.Printf("Downloading File - in Proress")

	resp, err := http.Get(url)
	log.Printf("Downloading File - DONE")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	err = os.MkdirAll("/home/kuro/kvm/test", os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	err = os.Chmod("/home/kuro/kvm/test", 0o777)
	if err != nil {
		return fmt.Errorf("failed to set directory permissions: %v", err)
	}

	log.Printf("Writing and returning data")
	return os.WriteFile("/home/kuro/kvm/test/imgfile.img", body, 0o644)
}

func DeletePoolIfExists(conn *libvirt.Connect, poolName string) error {
	if ex, _ := PoolExists(conn, poolName); ex {
		log.Print("Pool exists")
		if err := DeletePool(conn, poolName); err != nil {
			return fmt.Errorf("failed to delete: %s\n", err)
		}
	}
	return nil
}
