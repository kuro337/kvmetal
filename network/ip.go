package network

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"kvmgo/utils"
)

/*
go get golang.org/x/crypto/ssh

WORKER_IP=$(sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml worker | awk -F"'" '/mac address/{print $2}') | awk '{print $1}')

Make sure SSH is active on Node
systemctl status ssh

sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml worker | awk -F"'" '/mac address/{print $2}') | awk '{print $1}'

sudo apt-get install openssh-server
*/

/*
Using arp-scan to get the IP of a VM

Usage:

	ip , err := GetVMIPAddr("kubecontrol")
	log.Printf("IP of Control Node is %s",ip)
*/
func GetVMIPAddr(vmName string) (*IPAddressWithSubnet, error) {
	cmdString := fmt.Sprintf("sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml %s | awk -F\"'\" '/mac address/{print $2}') | awk '{print $1}'", vmName)

	cmd := exec.Command("bash", "-c", cmdString)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(&out)
	if scanner.Scan() {
		ipAddr, err := NewIPAddressWithSubnet(strings.TrimSpace(scanner.Text()))

		if err == nil && ipAddr != nil {
			return ipAddr, nil
		}

		// return strings.TrimSpace(scanner.Text()), nil
	}

	log.Printf("Error Parsing Output into an IP and Subnet for VM %s  ERROR:%s", vmName, err)

	return nil, fmt.Errorf("no IP address found for VM: %s", vmName)
}

/*
	func getVMInterface(xmlFilePath string) (string, error) {
		xmlFile, err := os.ReadFile(xmlFilePath)
		if err != nil {
			return "", err
		}

		var iface NetworkInterfaceVM
		if err := xml.Unmarshal(xmlFile, &iface); err != nil {
			return "", err
		}

		return iface.Bridge.Name, nil
	}
*/
func SimulateVMConfig() ForwardingConfig {
	return ForwardingConfig{
		VMName:    "test",
		Interface: "virbr0",
		PrivateIP: net.ParseIP("127.0.0.1"),
		HostIP:    net.ParseIP("192.168.1.1"),
		// ExternalIP: net.ParseIP("8.8.8.8"),
		PortMap: []PortMapping{
			{Protocol: TCP, HostPort: 80, VMPort: 8080}, // Maps host port 1100 to VM port 3000 over TCP
			{Protocol: TCP, HostPort: 443, VMPort: 443}, // Maps host port 443 to VM port 443 over TCP
			{Protocol: UDP, HostPort: 53, VMPort: 53},   // Maps host port 27016 to VM port 27016 over UDP
		},
		// PortRange: []PortRange{
		// 	{VMStartPort: 8888, VMEndPortNum: 8890, HostStartPortNum: 8888, HostEndPortNum: 8890, Protocol: TCP},     // TCP range mapping
		// 	{VMStartPort: 30000, VMEndPortNum: 30100, HostStartPortNum: 30000, HostEndPortNum: 30100, Protocol: UDP}, // UDP range mapping
		// },
	}
}

/*
 */
func CreatePortForwardingConfig(
	vmName, vmNetInterface string,
	vmIP, hostIP, externalIP net.IP,
	directPortMappings []PortMapping,
	rangePortMappings []PortRange,
) ForwardingConfig {
	return ForwardingConfig{
		VMName:      vmName,
		HostIP:      hostIP,
		PrivateIP:   vmIP,
		ExternalIP:  externalIP,
		Interface:   vmNetInterface,
		PortMap:     directPortMappings,
		PortRange:   rangePortMappings,
		LastUpdated: time.Now().Format(time.RFC3339),
	}
}

// Creates the Port Forwarding Config - that is used by the application to Expose VM's
// Consider setting this to a default external IP and Host and call it upon VM Creation
func GeneratePortForwardingConfigExtractDomainIP(vmName string,
	externalIp net.IP,
	directPortMappings []PortMapping,
	rangePortMappings []PortRange,
) (*ForwardingConfig, error) {
	// get Host IP , VM IP, and VM Subnet Dynamically

	hostIP, err := GetHostIP()
	if err != nil {
		log.Print(utils.TurnError("Failed to get Host IP"))
	}

	vmIpAddr, err := GetVMIPAddr(vmName)
	if err != nil {
		log.Print(utils.TurnError("Failed to get VM Private IP"))
		return nil, err
	}

	log.Printf("Host IP:%s\nVM IP Addr:%s\n", hostIP, vmIpAddr)

	fwdConfig := CreatePortForwardingConfig(vmName, "virbr0",
		vmIpAddr.IP, hostIP.IP, externalIp, directPortMappings, rangePortMappings)

	return &fwdConfig, nil
}

// Creates the Port Forwarding Config - that is used by the application to Expose VM's
// Consider setting this to a default external IP and Host and call it upon VM Creation
func GenerateDefaultPortForwardingConfig(domain string, domainIP,
	externalIp,
	hostIP net.IP,
	directPortMappings []PortMapping,
	rangePortMappings []PortRange,
) (*ForwardingConfig, error) {
	fwdConfig := CreatePortForwardingConfig(domain, "virbr0", domainIP,
		hostIP, externalIp, directPortMappings, rangePortMappings)

	return &fwdConfig, nil
}
