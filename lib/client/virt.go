package client

import (
	"fmt"
	"log"
	"slices"
	"time"

	"libvirt.org/go/libvirt"
	libvirtxml "libvirt.org/libvirt-go-xml"

	. "kvmgo/lib/domain"
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

	return NewDomainWrapper(domain, v.conn, dom), nil
	// return &Domain{Name: domain, domain: dom}, nil
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
