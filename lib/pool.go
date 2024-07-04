package lib

import (
	"fmt"

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

	err = pool.Create(libvirt.STORAGE_POOL_CREATE_WITH_BUILD)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage pool: %v", err)
	}

	err = pool.SetAutostart(true)
	if err != nil {
		return nil, fmt.Errorf("failed to set autostart for storage pool: %v", err)
	}

	defer pool.Free()

	return &Pool{
		name: name,
		path: path,
		pool: pool,
	}, nil
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
	xml := p.GetXMLFromPath(name, fromPath, capacityGB)
	return p.CreateImageXML(xml)
}

func (p *Pool) CreateImageURL(name, fromPath string, capacityGB int) error {
	xml := p.GetXMLFromPath(name, fromPath, capacityGB)
	return p.CreateImageXML(xml)
}

func (p *Pool) CreateImageXML(xml string) error {
	pool := p.pool
	vol, err := pool.StorageVolCreateXML(xml, 0)
	defer vol.Free()
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return err
	}
	return nil
}
