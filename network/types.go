package network

import (
	"encoding/xml"
	"fmt"
	"net"
)

type NetProtocol string

const (
	TCP NetProtocol = "tcp"
	UDP NetProtocol = "udp"
)

type Network struct {
	XMLName xml.Name `xml:"network"`
	IP      IP       `xml:"ip"`
}

type IP struct {
	Address string `xml:"address,attr"`
	Netmask string `xml:"netmask,attr"`
}

// IPAddressWithSubnet holds an IP address and its subnet mask.
type IPAddressWithSubnet struct {
	IP     net.IP
	Subnet int
}

func (ip IPAddressWithSubnet) String() string {
	return fmt.Sprintf("%s/%d", ip.IP.String(), ip.Subnet)
}

// NewIPAddressWithSubnet creates a new IPAddressWithSubnet from a CIDR string.
// / The input string can be in CIDR notation (e.g., "192.168.1.1/24") or a plain IP address (e.g., "192.168.1.1").
func NewIPAddressWithSubnet(input string) (*IPAddressWithSubnet, error) {
	// First, try parsing as CIDR to get both IP and subnet
	ip, ipNet, err := net.ParseCIDR(input)
	if err == nil {
		// Successfully parsed as CIDR
		ones, _ := ipNet.Mask.Size()
		return &IPAddressWithSubnet{
			IP:     ip,
			Subnet: ones,
		}, nil
	}

	// If CIDR parsing failed, try parsing as a plain IP address
	ip = net.ParseIP(input)
	if ip == nil {
		// Parsing as a plain IP also failed
		return nil, fmt.Errorf("parsing input failed: %s is neither a valid CIDR nor an IP address", input)
	}

	// Successfully parsed as a plain IP, but without subnet information
	return &IPAddressWithSubnet{
		IP:     ip,
		Subnet: 0, // Subnet is unknown or not applicable
	}, nil
}

type VMLeaseInfo struct {
	IP       net.IP
	Subnet   string
	Hostname string
	MAC      string
	Protocol string
}

/* ForwardingConfigs represents the State of our Virtual Machines - stored in a resilient way */
type ForwardingConfigs struct {
	Configs     []ForwardingConfig `json:"configs"`
	LastUpdated string             `json:"last_updated"`
}

/* ForwardingConfig Defines the Routing Rules by which Ports its exposes and which External IPs have access to it */
type ForwardingConfig struct {
	VMName      string        `json:"domain"`
	PortMap     []PortMapping `json:"port_map"`
	PortRange   []PortRange   `json:"port_range"`
	HostIP      net.IP        `json:"host_ip,omitempty"`
	PrivateIP   net.IP        `json:"private_ip,omitempty"`
	ExternalIP  net.IP        `json:"external_ip"`
	Interface   string        `json:"interface,omitempty"`
	LastUpdated string        `json:"last_updated"`
}

type PortRange struct {
	VMStartPort      int         `json:"vm_start_port"`
	VMEndPortNum     int         `json:"vm_end_port,omitempty"`
	HostStartPortNum int         `json:"host_start_port,omitempty"`
	HostEndPortNum   int         `json:"host_end_port,omitempty"`
	Protocol         NetProtocol `json:"protocol"`
}

type PortMapping struct {
	HostPort int         `json:"host_port"`
	VMPort   int         `json:"vm_port"`
	Protocol NetProtocol `json:"protocol"`
}

type PortDetail struct {
	PublicStart  int
	PublicEnd    int
	PrivateStart int
	PrivateEnd   int
}

// NetworkInterfaceVM stores the metadata at /etc/libvirt/qemu/vm_name for a VM/domain
type NetworkInterfaceVM struct {
	XMLName xml.Name `xml:"interface"`
	Bridge  struct {
		Name string `xml:"name,attr"`
	} `xml:"bridge"`
}
