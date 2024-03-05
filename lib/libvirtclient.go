package lib

import (
	"fmt"
	"log"

	"libvirt.org/go/libvirt"
	libvirtxml "libvirt.org/libvirt-go-xml"
)

type VirtClient struct {
	conn libvirt.Connect
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
	return &Domain{name: domain, domain: dom}, nil
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

/* Connect to Libvirt and Return the Client */
func ConnectLibvirt() (*VirtClient, error) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		return nil, err
	}

	return &VirtClient{conn: *conn}, nil
}
