package network

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

/*
Get the IP Subnet that our Libvirt NAT bridge runs on

	virsh net-dumpxml default // gets network interfaces info for Libvirt

	<bridge name='virbr0' stp='on' delay='0'/>
	<mac address='52:54:00:ac:91:9f'/>
	<ip address='192.168.122.1' netmask='255.255.255.0'>

	// So our IP Subnet will be 192.168.122.1 - 1 => 192.168.122.0
*/
func GetLibvirtIpSubnet() (string, error) {
	// Execute the virsh command to get the XML output for the default network
	cmd := exec.Command("virsh", "net-dumpxml", "default")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute virsh net-dumpxml default: %v", err)
	}

	// Parse the XML to find the IP address and netmask
	var net Network
	if err := xml.Unmarshal(out.Bytes(), &net); err != nil {
		return "", fmt.Errorf("failed to parse XML: %v", err)
	}

	// Correctly calculate the subnet base address and convert netmask to CIDR notation
	cidr := netmaskToCIDR(net.IP.Netmask)
	if cidr == "" {
		return "", fmt.Errorf("failed to convert netmask to CIDR notation")
	}

	// Split the IP address to construct the base subnet address
	ipParts := strings.Split(net.IP.Address, ".")
	if len(ipParts) != 4 {
		return "", fmt.Errorf("unexpected IP address format: %s", net.IP.Address)
	}
	// Replace the last octet with "0" to denote the subnet base
	ipParts[3] = "0"
	subnetBase := strings.Join(ipParts, ".")

	// Return the subnet in CIDR notation
	subnet := fmt.Sprintf("%s/%s", subnetBase, cidr)
	return subnet, nil
}

// netmaskToCIDR converts a netmask to CIDR notation.
func netmaskToCIDR(netmask string) string {
	bits := strings.Split(netmask, ".")
	if len(bits) != 4 {
		fmt.Println("Unexpected netmask format:", netmask)
		return ""
	}

	cidr := 0
	for _, bit := range bits {
		b, _ := strconv.Atoi(bit) // Ignoring error for simplicity
		for b > 0 {
			cidr++
			b = b << 1 & 0xFF
		}
	}

	return fmt.Sprintf("%d", cidr)
}
