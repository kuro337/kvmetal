package lib

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"libvirt.org/go/libvirt"
)

type Pool struct {
	name string
	path string
	pool *libvirt.StoragePool
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
		name: name,
		path: path,
		pool: pool,
	}, nil
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

func (p *Pool) Delete() error {
	if p.pool == nil {
		return fmt.Errorf("storage pool is nil")
	}
	// Destroy the pool if it is active
	if active, err := p.Active(); err == nil && active {
		if err := p.pool.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy storage pool: %v", err)
		}
	}
	// Undefine the pool
	if err := p.pool.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine storage pool: %v", err)
	}
	return nil
}

// Active() checks whether the Pool is Active or not
func (p *Pool) Active() (bool, error) {
	pool := p.pool
	if err := pool.Create(0); err != nil && err.(libvirt.Error).Code != libvirt.ERR_OPERATION_INVALID {
		fmt.Printf("Failed to activate storage pool: %v\n", err)
		return false, fmt.Errorf("Storage Pool not active for %s", p.name)
	}
	return true, nil
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

	log.Printf("Writing and returning data")
	return os.WriteFile("/home/kuro/kvm/test/imgfile.img", body, 0o644)
}

/*
func DownloadImageToTemp(url string) (string, func(), error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "img-*.img")
	if err != nil {
		return "", nil, err
	}

	if url == "" {
		log.Fatalf("failure empty url")
	}

	log.Println("Temp File ", tempFile.Name())

	log.Printf("Downloading File - in Proress")

	// Get the image from the URL
	resp, err := http.Get(url)

	log.Printf("Downloading File - DONE")
	if err != nil {
		// tempFile.Close()
		// os.Remove(tempFile.Name())
		return "", nil, err
	}
	defer resp.Body.Close()

	// Copy the content to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		// tempFile.Close()
		// os.Remove(tempFile.Name())
		return "", nil, err
	}

	// Ensure the temporary file is removed after use
	cleanup := func() {
		// tempFile.Close()
		// os.Remove(tempFile.Name())
		log.Println("Temp File Removed ", tempFile.Name())
	}

	return tempFile.Name(), cleanup, nil
}
*/
