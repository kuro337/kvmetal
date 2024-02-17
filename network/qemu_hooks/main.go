package qemu_hooks

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"kvmgo/network"
)

/*
go build -o qemuhookintercept main.go
sudo cp qemuhookintercept /etc/libvirt/hooks
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept
sudo ln -sf /etc/libvirt/hooks/qemuhookintercept /etc/libvirt/hooks/qemu
sudo ln -sf /etc/libvirt/hooks/qemuhookintercept /etc/libvirt/hooks/lxc
sudo service libvirtd restart

virsh start spark

1. example log
LIBVIRT_HOOK: 2024/02/16 19:38:44 Event received - Domain: spark, Action: prepare, Time: 2024-02-16T19:38:44-05:00
LIBVIRT_HOOK: 2024/02/16 19:38:45 Event received - Domain: spark, Action: start, Time: 2024-02-16T19:38:45-05:00
LIBVIRT_HOOK: 2024/02/16 19:38:45 Event received - Domain: spark, Action: started, Time: 2024-02-16T19:38:45-05:00

os.Args[1] is "spark"
os.Args[2] : Action = "prepare" , "start" , "started" etc.


*/

// https://www.libvirt.org/hooks.html

// !!!! When a VM is shutdown - make sure to call  qemu_hooks.ClearVMConfig("spark") !!!!

const logfileDir = "/home/kuro/Documents/Code/Go/kvmgo/data/network/logs/"

/* IMPORTANT: Do NOT call any Libvirt API within a Hook

This will cause DEADLOCKS */

type HookAction string

const (
	Start     HookAction = "start"
	Stopped   HookAction = "stopped"
	Reconnect HookAction = "reconnect"
	Started   HookAction = "started"
	Prepare   HookAction = "prepare"
	Restore   HookAction = "restore"
	Release   HookAction = "release"
	Migrate   HookAction = "migrate"
)

type ChainHook string

const (
	DNAT ChainHook = "DNAT"
	SNAT ChainHook = "SNAT"
	FWD  ChainHook = "FWD"
)

type ChainAction string

const (
	INSERT ChainAction = "-I"
	DELETE ChainAction = "-D"
)

// table = "nat"/"filter", name = dnat/snat/fwd chain
func (c *LibvirtChain) CreateChain(table string) string {
	chain := fmt.Sprintf("iptables -t %s -N %s", table, c.String())
	fmt.Println(chain)
	return chain
}

func (c *LibvirtChain) DeleteChain(table string) string {
	return fmt.Sprintf("iptables -t %s -F %s", table, c.String()) +
		"\n" +
		fmt.Sprintf("iptables -t %s -X %s", table, c.String())
}

func (c *LibvirtChain) String() string {
	return fmt.Sprintf("%s-%s", string(c.ChainType), c.VMName)
}

type LibvirtChain struct {
	VMName    string
	ChainType ChainHook
}

func NewChain(vmName string, chainType ChainHook) LibvirtChain {
	return LibvirtChain{VMName: vmName, ChainType: chainType}
}

func HandleForwardingEvent(forwardingConfig network.ForwardingConfig, action HookAction) []string {
	dnat_chain := NewChain(forwardingConfig.VMName, DNAT)
	snat_chain := NewChain(forwardingConfig.VMName, SNAT)
	fwd_chain := NewChain(forwardingConfig.VMName, FWD)

	switch action {
	case Start:
		return StartForwarding(
			dnat_chain, snat_chain, fwd_chain,
			forwardingConfig.HostIP, forwardingConfig.PrivateIP, forwardingConfig.ExternalIP,
			forwardingConfig.PortMap, forwardingConfig.PortRange,
		)

	case Stopped:
		return StopForwarding(
			dnat_chain, snat_chain, fwd_chain,
			forwardingConfig.HostIP, forwardingConfig.PrivateIP,
		)

	case Reconnect:
		stopFirst := StopForwarding(
			dnat_chain, snat_chain, fwd_chain,
			forwardingConfig.HostIP, forwardingConfig.PrivateIP,
		)

		return append(stopFirst,
			StartForwarding(
				dnat_chain, snat_chain, fwd_chain,
				forwardingConfig.HostIP, forwardingConfig.PrivateIP, forwardingConfig.ExternalIP,
				forwardingConfig.PortMap, forwardingConfig.PortRange,
			)...,
		)

	default:
		return []string{}
	}
}

func StartForwarding(
	dnatChain, snatChain, fwdChain LibvirtChain,
	hostIp, vmPrivateIp, externalIp net.IP,
	directPortMappings []network.PortMapping,
	rangePortMappings []network.PortRange,
) []string {
	dnatCmd := dnatChain.CreateChain("nat")
	snatCmd := snatChain.CreateChain("nat")
	fwdCmd := fwdChain.CreateChain("filter")

	populated := PopulateChains(dnatChain, snatChain, fwdChain,
		hostIp, vmPrivateIp, externalIp,
		directPortMappings, rangePortMappings)

	insertChains := InsertChains(INSERT, dnatChain, snatChain, fwdChain, hostIp, vmPrivateIp)

	var combinedCmds []string
	combinedCmds = append(combinedCmds, dnatCmd, snatCmd, fwdCmd)
	combinedCmds = append(combinedCmds, populated...)
	combinedCmds = append(combinedCmds, insertChains...)

	return combinedCmds
}

func PopulateChains(
	dnatChain, snatChain, fwdChain LibvirtChain,
	publicIP, privateIP, externalIP net.IP,
	directPortMappings []network.PortMapping,
	rangePortMappings []network.PortRange,
) []string {
	var commands []string

	// Handle individual port mappings
	for _, mapping := range directPortMappings {
		dnatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -d %s --dport %d -j DNAT --to %s:%d",
			dnatChain.String(),
			mapping.Protocol,
			publicIP.String(),
			mapping.HostPort,
			privateIP.String(),
			mapping.VMPort)

		snatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -s %s --dport %d -j SNAT --to-source %s",
			snatChain.String(),
			mapping.Protocol,
			privateIP.String(),
			mapping.VMPort,
			externalIP.String())

		commands = append(commands, dnatCmd, snatCmd)
	}

	// Handle port ranges
	for _, rangeMapping := range rangePortMappings {

		dnatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -d %s --dport %d:%d -j DNAT --to %s:%d-%d",
			dnatChain.String(), rangeMapping.Protocol, publicIP.String(),
			rangeMapping.HostStartPortNum, rangeMapping.HostEndPortNum,
			privateIP.String(), rangeMapping.VMStartPort, rangeMapping.VMEndPortNum)

		snatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -s %s --dport %d:%d -j SNAT --to-source %s",
			snatChain.String(), rangeMapping.Protocol, privateIP.String(),
			rangeMapping.VMStartPort, rangeMapping.VMEndPortNum, externalIP.String())

		commands = append(commands, dnatCmd, snatCmd)
	}

	return commands
}

func InsertChains(action ChainAction, dnatChain, snatChain, fwdChain LibvirtChain,
	publicIP, privateIP net.IP,
) []string {
	commands := []string{
		fmt.Sprintf("iptables -t nat %s OUTPUT -d %s -j %s",
			string(action), publicIP.String(), dnatChain.String()),
		fmt.Sprintf("iptables -t nat %s PREROUTING -d %s -j %s",
			string(action), publicIP.String(), dnatChain.String()),
		fmt.Sprintf("iptables -t nat %s POSTROUTING -s %s -d %s -j %s",
			string(action), privateIP.String(), privateIP.String(), snatChain.String()),
		fmt.Sprintf("iptables -t filter %s FORWARD -d %s -j %s", string(action),
			privateIP.String(), fwdChain.String()),
	}
	return commands
}

func StopForwarding(dnatChain, snatChain, fwdChain LibvirtChain,
	hostIp, vmPrivateIp net.IP,
) []string {
	insertChains := InsertChains(INSERT,
		dnatChain, snatChain, fwdChain, hostIp, vmPrivateIp)

	//	dnatCmd := dnatChain.DeleteChain("nat")
	//	snatCmd := snatChain.DeleteChain("nat")
	//	fwdCmd := fwdChain.DeleteChain("filter")

	return append(insertChains,
		dnatChain.DeleteChain("nat"),
		snatChain.DeleteChain("nat"),
		fwdChain.DeleteChain("filter"))
}

/*

def stop_forwarding(dnat_chain, snat_chain, fwd_chain, public_ip, private_ip):
    """ tears down the iptables port-forwarding rules. """
    insert_chains("-D", dnat_chain, snat_chain,
                  fwd_chain, public_ip, private_ip)
    delete_chain("nat", dnat_chain)
    delete_chain("nat", snat_chain)
    delete_chain("filter", fwd_chain)

*/

func main() {
	// Ensure there are enough arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <domain> <action>")
		os.Exit(1)
	}

	// Extract the domain name and action from the arguments
	virDomain := os.Args[1]
	action := os.Args[2]

	logfilePath := filepath.Join(logfileDir, "libvirtHookEvents.log")

	// Open or create a log file
	logFile, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create a logger that writes to the file
	logger := log.New(logFile, "LIBVIRT_HOOK: ", log.LstdFlags)

	// Log the event
	logger.Printf("Event received - Domain: %s, Action: %s, Time: %s\n", virDomain, action, time.Now().Format(time.RFC3339))

	// Optionally, also print the log to stdout
	//	fmt.Printf("Event received - Domain: %s, Action: %s, Time: %s\n", virDomain, action, time.Now().Format(time.RFC3339))
}
