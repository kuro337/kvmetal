package lib

import (
	"fmt"
	"log"
	"slices"
	"time"

	"libvirt.org/go/libvirt"
	libvirtxml "libvirt.org/libvirt-go-xml"
)

type VirtClient struct {
	conn    *libvirt.Connect
	domains map[string]*Domain
}

/* Connect to Libvirt and Return the Client */
func ConnectLibvirt() (*VirtClient, error) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		return nil, err
	}

	return &VirtClient{conn: conn, domains: make(map[string]*Domain)}, nil
}

func (v *VirtClient) Close() {
	v.Close()
}

func (v *VirtClient) GetDomSlice() []*Domain {
	var doms []*Domain

	for _, d := range v.domains {
		doms = append(doms, d)
	}

	return doms
}

// AwaitDomains will wait until Domains are Ready and if they do not become Ready, returns an Error
func AwaitDomains(domains []string) (*VirtClient, map[string]*Domain, error) {
	l, err := ConnectLibvirt()
	if err != nil {
		return nil, nil, fmt.Errorf("Error Connecting %s", err)
	}

	// defer l.conn.Close()

	for _, d := range domains {
		if err := l.AddDomain(d); err != nil {
			return nil, nil, err
		}
	}

	log.Printf("Domains Added: %d\n", len(l.domains))

	if err := l.Running(); err != nil {
		return nil, nil, err
	}

	return l, l.domains, nil
}

// Checks if all the Domains are running
func (v *VirtClient) Running() error {
	retries := 8
	doms := v.GetDomSlice()

	log.Printf("Doms Size: %d\n", len(doms))

	delay := 5

	i := 0

	for i < retries {
		j := 0
		for j < len(doms) {
			r, err := doms[j].IsRunning()
			if err != nil {
				log.Printf("Domain %s not running, retrying.\n", doms[j].Name)
				// return fmt.Errorf("Error: %s", err)
			}

			if r {
				doms = slices.Delete(doms, j, j+1)
			} else {
				j++
			}
		}

		if len(doms) == 0 {
			break
		}

		wait := delay + (1 << i)
		log.Printf("Attempt %d: Backoff: %ds\n", i, wait)
		time.Sleep(time.Duration(wait) * time.Second)

		i++
	}

	if i == retries && len(doms) > 0 {
		return fmt.Errorf("Not all domains are stopped after retries")
	}

	return nil
}

func (v *VirtClient) AddDomain(domain string) error {
	retries := 8
	delay := 5
	i := 0

	var ferr error

	for i < retries {

		dom, err := NewDomain(v.conn, domain)

		if err == nil {
			v.domains[domain] = dom
			ip, _ := dom.IP()
			log.Println("DOM IP: " + ip)
			return nil
		}

		wait := delay + (1 << i)
		log.Printf("Failed getting domain attempt %d - sleeping %d seconds. Error:%s", i, wait, err)

		time.Sleep(time.Duration(wait) * time.Second)
		i++
		ferr = err
	}
	return fmt.Errorf("Failed getting domain attempt %d Error:%s", i, ferr)
}

// ListInterfaces() lists all Active Network Interfaces
func (v *VirtClient) ListInterfaces() error {
	interfaces, err := v.conn.ListAllInterfaces(libvirt.CONNECT_LIST_INTERFACES_ACTIVE)
	if err != nil {
		log.Printf("Failed to List Network Interfaces. ERROR:%s", err)
		return err
	}

	for _, iface := range interfaces {
		// Fetch the XML description of the interface

		xmlDesc, err := iface.GetXMLDesc(0)
		if err != nil {
			log.Printf("Failed to get XML description for interface: %v", err)
			continue
		}

		log.Println(xmlDesc)
	}

	return nil
}

// Gets the IP Addresses associated with the domain. A Domain can have multiple IP addresses such as IPv4, IPv6, so it returns a List of all of them.
func (v *VirtClient) GetIPFromDHCPLeases(domainName string) ([]string, error) {
	var ips []string

	dom, err := v.conn.LookupDomainByName(domainName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup domain by name %s: %v", domainName, err)
	}
	defer dom.Free()

	leases, err := dom.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		return nil, fmt.Errorf("failed to list all interface addresses from DHCP leases: %v", err)
	}

	for _, iface := range leases {
		for _, addr := range iface.Addrs {
			ips = append(ips, addr.Addr)
		}
	}

	return ips, nil
}

// Get the Domain (VM)
func (v *VirtClient) GetDomain(domain string) (*Domain, error) {
	dom, err := v.conn.LookupDomainByName(domain)
	if err != nil {
		log.Printf("Failed Lookup Domain %s", domain)
		return nil, err
	}
	return &Domain{Name: domain, domain: dom}, nil
}

// Parses the XML for a Domain and Prints it
func (v *VirtClient) ParseXML(domain string) (*libvirtxml.Domain, error) {
	dom, err := v.conn.LookupDomainByName(domain)
	if err != nil {
		log.Printf("Failed Lookup Domain %s", domain)
		return nil, err
	}

	// info,_ := dom.GetInfo()

	xmldoc, err := dom.GetXMLDesc(0)
	if err != nil {
		log.Printf("Failed Pulling XML for Domain %s", domain)
		return nil, err
	}

	domcfg := &libvirtxml.Domain{}
	err = domcfg.Unmarshal(xmldoc)
	if err != nil {
		log.Printf("Failed Parsing XML for Domain %s", domain)
		return nil, err
	}

	fmt.Printf("Virt type %s\n", domcfg.Type)

	return domcfg, nil
}

/////////////////// VM Image Generation for KVM Images from Base Images

// CreateStoragePool creates the Storage pool if it doesnt exist
// @Usage
// err := CreateStoragePool("default" , "/var/lib/libvirt/images")
func (v *VirtClient) CreateStoragePool(poolName, poolPath string) error {
	// Check if the storage pool already exists
	pool, err := v.conn.LookupStoragePoolByName(poolName)

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

	pool, err = v.conn.StoragePoolCreateXML(poolXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage pool: %v\n", err)
		return err
	}

	defer pool.Free()

	return nil
}

func (v *VirtClient) GetStoragePool(poolName string) (*libvirt.StoragePool, error) {
	pool, err := v.conn.LookupStoragePoolByName(poolName)
	return pool, err
}

func (v *VirtClient) CreateQemuImg() {
	poolXML := `<pool type='dir'>
                    <name>default</name>
                    <target>
                        <path>/var/lib/libvirt/images</path>
                    </target>
                </pool>`

	pool, err := v.conn.StoragePoolCreateXML(poolXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage pool: %v\n", err)
		return
	}
	defer pool.Free()

	volXML := `<volume>
                   <name>new_img.qcow2</name>
                   <allocation>0</allocation>
                   <capacity unit="G">20</capacity>
                   <target>
                       <format type='qcow2'/>
                   </target>
               </volume>`

	vol, err := pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return
	}
	defer vol.Free()

	v.Close()
}

func (v *VirtClient) CreateImgVolume(poolName string) error {
	pool, err := v.GetStoragePool(poolName)
	if err != nil {
		return err
	}

	// Ensure the pool is active
	if err := pool.Create(0); err != nil && err.(libvirt.Error).Code != libvirt.ERR_OPERATION_INVALID {
		fmt.Printf("Failed to activate storage pool: %v\n", err)
		return fmt.Errorf("Storage Pool not active for %s", poolName)
	}

	// Create a new storage volume
	volXML := `<volume>
                   <name>new_img.qcow2</name>
                   <allocation>0</allocation>
                   <capacity unit="G">20</capacity>
                   <target>
                       <format type='qcow2'/>
                   </target>
               </volume>`

	vol, err := pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return err
	}
	defer vol.Free()

	v.Close()
	return nil
}

func (v *VirtClient) Conn() *libvirt.Connect {
	return v.conn
}

func (v *VirtClient) StoragePoolExists(poolName string) bool {
	if _, err := v.conn.LookupStoragePoolByName(poolName); err != nil {
		return false
	}
	return true
}

// To Delete an Image/Volume when we have the Pool it belongs to
func (v *VirtClient) DeleteImgVolume(poolName, volumeName string) error {
	pool, err := v.conn.LookupStoragePoolByName(poolName)
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

// To Delete an Image/Volume when we dont have the Pool it belongs to
func (v *VirtClient) DeleteImgVolumeByName(volumeName string) error {
	pools, err := v.conn.ListAllStoragePools(0)
	if err != nil {
		return fmt.Errorf("failed to list storage pools: %v", err)
	}

	for _, pool := range pools {
		vol, err := pool.LookupStorageVolByName(volumeName)
		if err != nil {
			if libvirtError, ok := err.(libvirt.Error); ok && libvirtError.Code == libvirt.ERR_NO_STORAGE_VOL {
				// Volume not found in this pool, continue to the next pool
				continue
			}
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

	return fmt.Errorf("storage volume '%s' not found in any pool", volumeName)
}
