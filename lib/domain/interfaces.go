package domain

import (
	"fmt"

	"libvirt.org/go/libvirt"
)

type IPResult struct {
	Source string   `json:"source"`
	IPs    []string `json:"ips"`
}

type DomainIPs struct {
	DomainName string     `json:"domainName"`
	Results    []IPResult `json:"results"`
}

func (d *Domain) GetAllIPs() (*DomainIPs, error) {
	sources := []struct {
		Name   string
		Source libvirt.DomainInterfaceAddressesSource
	}{
		{"LEASE", libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE},
		{"ARP", libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_ARP},
		// Requires qemu-guest-agent to be running on VM
		// {"AGENT", libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT},
	}

	results := []IPResult{}

	for _, src := range sources {
		var ips []string
		ifaces, err := d.domain.ListAllInterfaceAddresses(src.Source)
		if err != nil {
			fmt.Printf("Error listing addresses for source %s: %v\n", src.Name, err)
			continue
		}

		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				ips = append(ips, addr.Addr)
			}
		}

		results = append(results, IPResult{Source: src.Name, IPs: ips})
	}

	domainIPs := &DomainIPs{DomainName: d.Name, Results: results}

	return domainIPs, nil
}
