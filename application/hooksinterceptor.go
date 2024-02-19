package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"kvmgo/network"
	"kvmgo/network/qemu_hooks"
	"kvmgo/utils"
)

/*
QEMU Hooks Port Forwarding Entry Point

When placed in /etc/libvirt/hooks/<APP>

And linked with /etc/libvirt/hooks/qemu and /etc/libvirt/hooks/lxc

The <APP> is ran in response to any qemu and lxc event

Reference: https://www.libvirt.org/hooks.html
*/
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <domain> <action>")
		os.Exit(1)
	}

	virDomain := os.Args[1]
	action := os.Args[2]

	logger, err := LogHookEvent(virDomain, action)

	if err != nil || logger == nil {
		os.Exit(1) // this means it wont log our actions
	}

	cmds, err := HandleQemuHookEvent(action, virDomain)
	if err != nil {
		logger.Printf("Error Handling Qemu Hooks Event for %s ERROR:%s", action, err)
	}
	if len(cmds) > 1 {
		if err := utils.WriteArraytoFile(cmds, CmdsFilePath); err != nil {
			logger.Printf("Failed writing generated forwarding commands to file %s ERROR:%s,", CmdsFilePath, err)
		}
	}
	logger.Printf("Successfully Generated Commands Logs file at %s", CmdsFilePath)
}

func HandleQemuHookEvent(action, domain string) ([]string, error) {
	readConfig, err := qemu_hooks.ReadVMConfigFromFile(domain)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return []string{}, err
	}

	if readConfig == nil {
		fmt.Printf("No current Config for Domain %s\n", domain)
		return []string{}, fmt.Errorf("No current Config for Domain %s\n", domain)
	}
	switch action {
	case "start":
		return HandleForwardingEvent(Start, readConfig), nil
	case "stopped":
		return HandleForwardingEvent(Stopped, readConfig), nil
	case "reconnect":
		return HandleForwardingEvent(Reconnect, readConfig), nil
	}
	return []string{}, nil
}

// https://www.libvirt.org/hooks.html

// !!!! When a VM is shutdown - make sure to call  qemu_hooks.ClearVMConfig("spark") !!!!

const logfileDir = "/home/kuro/Documents/Code/Go/kvmgo/data/network/logs/"

const CmdsFilePath = "/home/kuro/Documents/Code/Go/kvmgo/data/network/logs/cmds"

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

func HandleForwardingEvent(action HookAction, forwardingConfig *network.ForwardingConfig) []string {
	dnat_chain := NewChain(forwardingConfig.VMName, DNAT)
	snat_chain := NewChain(forwardingConfig.VMName, SNAT)
	fwd_chain := NewChain(forwardingConfig.VMName, FWD)

	switch action {
	case Start:
		return StartForwarding(
			dnat_chain, snat_chain, fwd_chain,
			forwardingConfig.HostIP, forwardingConfig.PrivateIP, forwardingConfig.ExternalIP,
			forwardingConfig.PortMap, forwardingConfig.PortRange, forwardingConfig.Interface,
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
				forwardingConfig.PortMap, forwardingConfig.PortRange, forwardingConfig.Interface,
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
	net_interface string,
) []string {
	// err := DisableBridgeFiltering()
	// if err != nil {
	// 	log.Printf("Failed Disabling Bridge ERROR:%s", err)
	// }

	dnatCmd := dnatChain.CreateChain("nat")
	snatCmd := snatChain.CreateChain("nat")
	fwdCmd := fwdChain.CreateChain("filter")

	populated := PopulateChains(dnatChain, snatChain, fwdChain,
		hostIp, vmPrivateIp, externalIp,
		directPortMappings, rangePortMappings, net_interface)

	insertChains := InsertChains(
		INSERT,
		dnatChain, snatChain, fwdChain,
		hostIp, vmPrivateIp)

	var combinedCmds []string
	combinedCmds = append(combinedCmds, dnatCmd, snatCmd, fwdCmd)
	combinedCmds = append(combinedCmds, populated...)
	combinedCmds = append(combinedCmds, insertChains...)

	return combinedCmds
}

func PopulateChains(
	dnatChain, snatChain, fwdChain LibvirtChain,
	hostIP, vmPrivateIP, externalIP net.IP,
	directPortMappings []network.PortMapping,
	rangePortMappings []network.PortRange,
	net_interface string,
) []string {
	var commands []string

	// Handle individual port mappings
	for _, mapping := range directPortMappings {
		vmIPandPort := fmt.Sprintf("%s:%d", vmPrivateIP.String(), mapping.VMPort)

		// Enable Forwarding of Traffic on Host Ports to VM Ports
		dnatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -d %s --dport %d -j DNAT --to %s",
			dnatChain.String(),
			mapping.Protocol,
			hostIP.String(),
			mapping.HostPort,
			vmIPandPort)

		// Only enable access from Specified Whitelisted External IP if specified - else open
		if externalIP != nil {
			dnatCmd += fmt.Sprintf(" -s %s", externalIP.String())
		}

		// Masquerade outgoing VM Traffic as coming from the Host to communicate with External Clients
		snatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -s %s --dport %d -j SNAT --to-source %s",
			snatChain.String(),
			mapping.Protocol,
			vmPrivateIP.String(),
			mapping.VMPort,
			hostIP.String())

		masqCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -s %s -d %s --dport %d -j MASQUERADE",
			snatChain.String(),
			mapping.Protocol,
			vmPrivateIP.String(),
			vmPrivateIP.String(), mapping.HostPort)

		fwdCmd := fmt.Sprintf("iptables -t filter -A %s -p %s -d %s --dport %d -j ACCEPT",
			fwdChain.String(),
			mapping.Protocol,
			vmPrivateIP.String(),
			mapping.VMPort)

		if net_interface != "" {
			fwdCmd += fmt.Sprintf(" -o %s", net_interface)
		}

		commands = append(commands, dnatCmd, snatCmd, masqCmd, fwdCmd)
	}

	// Handle port ranges
	for _, rangeMapping := range rangePortMappings {

		if rangeMapping.HostStartPortNum == 0 || rangeMapping.HostEndPortNum == 0 ||
			rangeMapping.VMStartPort == 0 || rangeMapping.VMEndPortNum == 0 {
			log.Printf("Invalid Range Port Mapping Passed. Skipping")
			continue
		}

		portRange := fmt.Sprintf("%d:%d", rangeMapping.HostStartPortNum, rangeMapping.HostEndPortNum)
		vmPortRange := fmt.Sprintf("%s:%d-%d", vmPrivateIP.String(), rangeMapping.VMStartPort, rangeMapping.VMEndPortNum)
		protocol := string(rangeMapping.Protocol)

		dnatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -d %s --dport %s -j DNAT --to %s",
			dnatChain.String(), protocol,
			hostIP.String(), portRange, vmPortRange)

		// SNAT command for outgoing traffic to be masqueraded as from the host
		snatCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -s %s --dport %s -j SNAT --to-source %s",
			snatChain.String(),
			rangeMapping.Protocol,
			vmPrivateIP.String(),
			portRange,
			hostIP.String())

		masqCmd := fmt.Sprintf("iptables -t nat -A %s -p %s -s %s -d %s --dport %s -j MASQUERADE",
			snatChain.String(),
			rangeMapping.Protocol,
			vmPrivateIP.String(),
			vmPrivateIP.String(),
			portRange)

		fwdCmd := fmt.Sprintf("iptables -t filter -A %s -p %s -d %s --dport %s -j ACCEPT",
			fwdChain.String(),
			rangeMapping.Protocol,
			vmPrivateIP.String(),
			portRange)

		// Conditionally add interface specification
		if net_interface != "" {
			fwdCmd += fmt.Sprintf(" -o %s", net_interface)
		}

		commands = append(commands, dnatCmd, snatCmd, masqCmd, fwdCmd)
	}

	return commands
}

func InsertChains(action ChainAction, dnatChain, snatChain, fwdChain LibvirtChain,
	publicIP, privateIP net.IP,
) []string {
	chainAction := string(action)

	return []string{
		fmt.Sprintf("iptables -t nat %s OUTPUT -d %s -j %s",
			chainAction,
			publicIP.String(),
			dnatChain.String()),

		fmt.Sprintf("iptables -t nat %s PREROUTING -d %s -j %s",
			chainAction,
			publicIP.String(),
			dnatChain.String()),

		fmt.Sprintf("iptables -t nat %s POSTROUTING -s %s -d %s -j %s",
			chainAction,
			privateIP.String(),
			privateIP.String(),
			snatChain.String()),

		fmt.Sprintf("iptables -t filter %s FORWARD -d %s -j %s",
			chainAction,
			privateIP.String(),
			fwdChain.String()),
	}
	// return commands
}

func StopForwarding(dnatChain, snatChain, fwdChain LibvirtChain,
	hostIp, vmPrivateIp net.IP,
) []string {
	insertChains := InsertChains(DELETE,
		dnatChain, snatChain, fwdChain, hostIp, vmPrivateIp)

	//	dnatCmd := dnatChain.DeleteChain("nat")
	//	snatCmd := snatChain.DeleteChain("nat")
	//	fwdCmd := fwdChain.DeleteChain("filter")

	return append(insertChains,
		dnatChain.DeleteChain("nat"),
		snatChain.DeleteChain("nat"),
		fwdChain.DeleteChain("filter"))
}

func LogHookEvent(domain, action string) (*log.Logger, error) {
	logfilePath := filepath.Join(logfileDir, "libvirtHookEvents.log")
	logFile, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return nil, err
	}
	defer logFile.Close()

	logger := log.New(logFile, "LIBVIRT_HOOK: ", log.LstdFlags)

	logger.Printf("Event received - Domain: %s, Action: %s, Time: %s\n",
		domain, action, time.Now().Format(time.RFC3339))
	return logger, nil
}

func ValidatingRules() {
	/*
		iptables -t nat -L DNAT-hadoop -v -n
		iptables -t nat -L SNAT-hadoop -v -n
		iptables -t filter -L FWD-hadoop -v -n
	*/
}

/*

sudo cp qemuhookintercept /etc/libvirt/hooks
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept
sudo ln -sf /etc/libvirt/hooks/qemuhookintercept /etc/libvirt/hooks/qemu
sudo ln -sf /etc/libvirt/hooks/qemuhookintercept /etc/libvirt/hooks/lxc
sudo service libvirtd restart

Cleanup/Disable

sudo rm -f /etc/libvirt/hooks/qemu
sudo rm -f /etc/libvirt/hooks/lxc
sudo rm -f /etc/libvirt/hooks/qemuhookintercept

go build -o qemuhookintercept hooksinterceptor.go
sudo cp qemuhookintercept /etc/libvirt/hooks/
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept
# confirm File was updated
ls -l /etc/libvirt/hooks/qemuhookintercept

virsh start spark

To compile whole dir

go build -o qemuhookintercept .
# optionally
sudo systemctl stop libvirtd & sudo systemctl stop libvirtd before and after copy

sudo cp qemuhookintercept /etc/libvirt/hooks/
sudo chmod +x /etc/libvirt/hooks/qemuhookintercept

# confirm File was updated
ls -l /etc/libvirt/hooks/qemuhookintercept

virsh start spark

go clean -i ./...
go build -o qemuhookintercept


1. example log
LIBVIRT_HOOK: 2024/02/16 19:38:44 Event received - Domain: spark, Action: prepare, Time: 2024-02-16T19:38:44-05:00
LIBVIRT_HOOK: 2024/02/16 19:38:45 Event received - Domain: spark, Action: start, Time: 2024-02-16T19:38:45-05:00
LIBVIRT_HOOK: 2024/02/16 19:38:45 Event received - Domain: spark, Action: started, Time: 2024-02-16T19:38:45-05:00

os.Args[1] is "spark"
os.Args[2] : Action = "prepare" , "start" , "started" etc.
*/
