package tests

import (
	"log"
	"testing"

	"kvmgo/network/qemu_hooks"
)

// This will generate the Forwarding Config - and also Write it to our Location
func TestFwdConfigReadWrite(t *testing.T) {
	err := qemu_hooks.ClearVMConfig("hadoop")
	if err != nil {
		log.Printf("Error clearing VM config: %v", err)
	}

	// externalIp := net.ParseIP("192.168.1.225")

	// fwdingConfig, err := network.GeneratePortForwardingConfig("hadoop",
	// 	externalIp,
	// 	[]network.PortMapping{
	// 		{Protocol: network.TCP, HostPort: 9999, VMPort: 8088},
	// 	},
	// 	[]network.PortRange{})
	// if err != nil {
	// 	log.Printf("Failed to Generate Config ERROR:%s", err)
	// 	os.Exit(1)
	// }

	// if err := qemu_hooks.WriteConfigToFile(*fwdingConfig); err != nil {
	// 	t.Errorf("Error writing config: %s", err)
	// 	return
	// }

	// readConfig, err := qemu_hooks.ReadVMConfigFromFile("hadoop")
	// if err != nil {
	// 	t.Errorf("Error reading config:%s", err)
	// 	return
	// }

	// fmt.Printf("Read config: %+v\n", readConfig)
	t.Errorf("trigger")
}

/*
Sample Correct IPTables Forwarding Rules

iptables -t nat -A DNAT-spark -p TCP -d 192.168.1.194 --dport 1100 -j DNAT --to 192.168.122.101:3000
iptables -t nat -A DNAT-spark -p TCP -d 192.168.1.194 --dport 8888:8890 -j DNAT --to 192.168.122.101:8888-8890
iptables -t nat -A SNAT-spark -p UDP -s 192.168.122.101 --dport 30000:30100 -j SNAT --to-source 192.168.1.225
*/

// func TestGeneratingForwardingActionsFromExistingConfig(t *testing.T) {
// 	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

// 	cmds, err := qemu_hooks.HandleQemuHookEvent("start", "spark")
// 	if err != nil {
// 		t.Errorf("Failed to Handle Qemu Start Event Hook ERROR:%s", err)
// 	}

// 	if err := utils.WriteArraytoFile(cmds, qemu_hooks.CmdsFilePath); err != nil {
// 		t.Errorf("Failed writing generated forwarding commands to file ERROR:%s,", err)
// 	}
// }

// func TestSimulatedDomainConfig(t *testing.T) {
// 	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

// 	domainConfig := network.SimulateVMConfig()

// 	startCmds := qemu_hooks.HandleForwardingEvent(qemu_hooks.Start, &domainConfig)

// 	var populateChain strings.Builder
// 	for _, chain := range startCmds {
// 		populateChain.Write([]byte(chain + "\n"))
// 	}
// 	fmt.Println(utils.TurnBlueDelimited(populateChain.String()))

// 	stopCmds := qemu_hooks.HandleForwardingEvent(qemu_hooks.Stopped, &domainConfig)

// 	reconnectedCmds := qemu_hooks.HandleForwardingEvent(qemu_hooks.Reconnect, &domainConfig)

// 	result := utils.GetResultBlock("Forwarding Command Results",
// 		"Start Event", startCmds,
// 		"Stop Event", stopCmds,
// 		"Reconnect Event", reconnectedCmds)

// 	fmt.Printf(result)

// 	// t.Errorf("Trigger")
// }

// func TestVMNetworkMetadata(t *testing.T) {
// 	hostIP, err := network.GetHostIP()
// 	if err != nil {
// 		t.Errorf("failed to get host IP")
// 	}
// 	libvirtIpSubnet, err := network.GetLibvirtIpSubnet()
// 	if err != nil {
// 		t.Errorf("Failed to get Nat Subnet")
// 	}

// 	vmIpAddr, err := network.GetVMIPAddr("spark")
// 	if err != nil {
// 		t.Errorf("Failed to get Nat Subnet")
// 	}

// 	log.Printf("Host IP:%s\nLibvirt Subnet:%s\nVM IP Addr:%s\n", hostIP, libvirtIpSubnet, vmIpAddr)

// 	t.Errorf("trigger")
// }

// go test -v
// go test
// go test circle_test.go
// go test -v ./mypackage -run TestMyFunction

/*
Python:

iptables -t nat -N DNAT-test
iptables -t nat -N SNAT-test
iptables -t filter -N FWD-test
iptables -t nat -A DNAT-test -p udp -d 192.168.1.1 --dport 53 -j DNAT --to 127.0.0.1:53
iptables -t nat -A SNAT-test -p udp -s 127.0.0.1 --dport 53 -j SNAT --to-source 192.168.1.1
iptables -t nat -A SNAT-test -p udp -s 127.0.0.1 -d 127.0.0.1 --dport 53 -j MASQUERADE
iptables -t filter -A FWD-test -p udp -d 127.0.0.1 --dport 53 -j ACCEPT -o virbr0
iptables -t nat -A DNAT-test -p tcp -d 192.168.1.1 --dport 80 -j DNAT --to 127.0.0.1:8080
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 --dport 8080 -j SNAT --to-source 192.168.1.1
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 -d 127.0.0.1 --dport 80 -j MASQUERADE
iptables -t filter -A FWD-test -p tcp -d 127.0.0.1 --dport 8080 -j ACCEPT -o virbr0
iptables -t nat -A DNAT-test -p tcp -d 192.168.1.1 --dport 443 -j DNAT --to 127.0.0.1:443
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 --dport 443 -j SNAT --to-source 192.168.1.1
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 -d 127.0.0.1 --dport 443 -j MASQUERADE
iptables -t filter -A FWD-test -p tcp -d 127.0.0.1 --dport 443 -j ACCEPT -o virbr0
iptables -t nat -I OUTPUT -d 192.168.1.1 -j DNAT-test
iptables -t nat -I PREROUTING -d 192.168.1.1 -j DNAT-test
iptables -t nat -I POSTROUTING -s 127.0.0.1 -d 127.0.0.1 -j SNAT-test
iptables -t filter -I FORWARD -d 127.0.0.1 -j FWD-test


Go:

iptables -t nat -N DNAT-test
iptables -t nat -N SNAT-test
iptables -t filter -N FWD-test
iptables -t nat -A DNAT-test -p tcp -d 192.168.1.1 --dport 80 -j DNAT --to 127.0.0.1:8080
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 --dport 8080 -j SNAT --to-source 192.168.1.1
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 -d 127.0.0.1 --dport 80 -j MASQUERADE
iptables -t filter -A FWD-test -p tcp -d 127.0.0.1 --dport 8080 -j ACCEPT -o virbr0
iptables -t nat -A DNAT-test -p tcp -d 192.168.1.1 --dport 443 -j DNAT --to 127.0.0.1:443
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 --dport 443 -j SNAT --to-source 192.168.1.1
iptables -t nat -A SNAT-test -p tcp -s 127.0.0.1 -d 127.0.0.1 --dport 443 -j MASQUERADE
iptables -t filter -A FWD-test -p tcp -d 127.0.0.1 --dport 443 -j ACCEPT -o virbr0
iptables -t nat -A DNAT-test -p udp -d 192.168.1.1 --dport 53 -j DNAT --to 127.0.0.1:53
iptables -t nat -A SNAT-test -p udp -s 127.0.0.1 --dport 53 -j SNAT --to-source 192.168.1.1
iptables -t nat -A SNAT-test -p udp -s 127.0.0.1 -d 127.0.0.1 --dport 53 -j MASQUERADE
iptables -t filter -A FWD-test -p udp -d 127.0.0.1 --dport 53 -j ACCEPT -o virbr0
iptables -t nat -I OUTPUT -d 192.168.1.1 -j DNAT-test
iptables -t nat -I PREROUTING -d 192.168.1.1 -j DNAT-test
iptables -t nat -I POSTROUTING -s 127.0.0.1 -d 127.0.0.1 -j SNAT-test
iptables -t filter -I FORWARD -d 127.0.0.1 -j FWD-test

All Match.

Stopping

Python:

iptables -t nat -D OUTPUT -d 192.168.1.1 -j DNAT-test
iptables -t nat -D PREROUTING -d 192.168.1.1 -j DNAT-test
iptables -t nat -D POSTROUTING -s 127.0.0.1 -d 127.0.0.1 -j SNAT-test
iptables -t filter -D FORWARD -d 127.0.0.1 -j FWD-test
iptables -t nat -F DNAT-test
iptables -t nat -X DNAT-test
iptables -t nat -F SNAT-test
iptables -t nat -X SNAT-test
iptables -t filter -F FWD-test
iptables -t filter -X FWD-test


iptables -t nat -D OUTPUT -d 192.168.1.1 -j DNAT-test
iptables -t nat -D PREROUTING -d 192.168.1.1 -j DNAT-test
iptables -t nat -D POSTROUTING -s 127.0.0.1 -d 127.0.0.1 -j SNAT-test
iptables -t filter -D FORWARD -d 127.0.0.1 -j FWD-test
iptables -t nat -F DNAT-test
iptables -t nat -X DNAT-test
iptables -t nat -F SNAT-test
iptables -t nat -X SNAT-test
iptables -t filter -F FWD-test
iptables -t filter -X FWD-test

*/
