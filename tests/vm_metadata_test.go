package tests

import (
	"fmt"
	"log"
	"net"
	"strings"
	"testing"

	"kvmgo/network"
	"kvmgo/network/qemu_hooks"
	"kvmgo/utils"
)

// This will generate the Forwarding Config - and also Write it to our Location
func TestFwdConfigReadWrite(t *testing.T) {
	err := qemu_hooks.ClearVMConfig("spark")
	if err != nil {
		log.Fatalf("Error clearing VM config: %v", err)
	}

	externalIp := net.ParseIP("192.168.1.225")

	fwdingConfig := network.GeneratePortForwardingConfig("spark", externalIp, []network.PortMapping{
		{Protocol: network.TCP, HostPort: 1100, VMPort: 3000},
		{Protocol: network.TCP, HostPort: 443, VMPort: 443},
		{Protocol: network.UDP, HostPort: 27016, VMPort: 27016},
	},
		[]network.PortRange{
			{
				VMStartPort: 8888, VMEndPortNum: 8890,
				HostStartPortNum: 8888, HostEndPortNum: 8890, Protocol: network.TCP,
			},
			{
				VMStartPort: 30000, VMEndPortNum: 30100,
				HostStartPortNum: 30000, HostEndPortNum: 30100, Protocol: network.UDP,
			},
		})

	if err := qemu_hooks.WriteConfigToFile(fwdingConfig); err != nil {
		fmt.Println("Error writing config:", err)
		return
	}

	readConfig, err := qemu_hooks.ReadVMConfigFromFile("spark")
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	fmt.Printf("Read config: %+v\n", readConfig)

	t.Errorf("Trigger")
}

/*
Sample Correct IPTables Forwarding Rules

iptables -t nat -A DNAT-spark -p TCP -d 192.168.1.194 --dport 1100 -j DNAT --to 192.168.122.101:3000
iptables -t nat -A DNAT-spark -p TCP -d 192.168.1.194 --dport 8888:8890 -j DNAT --to 192.168.122.101:8888-8890
iptables -t nat -A SNAT-spark -p UDP -s 192.168.122.101 --dport 30000:30100 -j SNAT --to-source 192.168.1.225
*/
func TestSimulatedDomainConfig(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	domainConfig := network.SimulateVMConfig()

	startCmds := qemu_hooks.HandleForwardingEvent(domainConfig, qemu_hooks.Start)

	var populateChain strings.Builder
	for _, chain := range startCmds {
		populateChain.Write([]byte(chain + "\n"))
	}
	fmt.Println(utils.TurnBlueDelimited(populateChain.String()))

	stopCmds := qemu_hooks.HandleForwardingEvent(domainConfig, qemu_hooks.Stopped)

	reconnectedCmds := qemu_hooks.HandleForwardingEvent(domainConfig, qemu_hooks.Reconnect)

	result := utils.GetResultBlock("Forwarding Command Results",
		"Start Event", startCmds,
		"Stop Event", stopCmds,
		"Reconnect Event", reconnectedCmds)

	fmt.Printf(result)

	t.Errorf("Trigger")
}

func TestVMNetworkMetadata(t *testing.T) {
	hostIP, err := network.GetHostIP()
	if err != nil {
		t.Errorf("failed to get host IP")
	}
	libvirtIpSubnet, err := network.GetLibvirtIpSubnet()
	if err != nil {
		t.Errorf("Failed to get Nat Subnet")
	}

	vmIpAddr, err := network.GetVMIPAddr("spark")
	if err != nil {
		t.Errorf("Failed to get Nat Subnet")
	}

	log.Printf("Host IP:%s\nLibvirt Subnet:%s\nVM IP Addr:%s\n", hostIP, libvirtIpSubnet, vmIpAddr)

	t.Errorf("trigger")
}

// // go test -v
// // go test
// // go test circle_test.go
// // go test -v ./mypackage -run TestMyFunction
