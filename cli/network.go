package cli

import (
	"fmt"
	"log"
	"net"

	"kvmgo/network"
	"kvmgo/network/qemu_hooks"
	"kvmgo/utils"
)

type NetworkExposeConfig struct {
	VM          string
	ExternalIP  net.IP
	PortMapping network.PortMapping
	PortRange   network.PortRange
}

// Generates Networking Config for VM and checks for Existing one - then generates the Forwarding Rules and Writes it to the Artifact Location for the VM
func HandleVMNetworkExposure(vmName string, vmPort, hostPort int, externalIp string, protocol string) error {
	netConfig := ParseNetExposeFlags(vmName, vmPort, hostPort, externalIp, protocol)

	if netConfig != nil {
		if err := CreateAndSetNetExposeConfig(*netConfig); err != nil {
			log.Printf("Failed To Create Forwarding Config ERROR:%s,", err)
			return err
		}
	}
	return nil
}

func ParseNetExposeFlags(vmName string, vmPort, hostPort int, externalIp string, protocol string) *NetworkExposeConfig {
	externalIP := net.ParseIP(externalIp)

	if externalIP == nil {
		log.Print(utils.TurnError(fmt.Sprintf("Failed to Parse External IP %s", externalIP)))
		return nil
	}

	return &NetworkExposeConfig{
		VM:         vmName,
		ExternalIP: externalIP,
		PortMapping: network.PortMapping{
			Protocol: network.NetProtocol(protocol),
			HostPort: int(hostPort),
			VMPort:   int(vmPort),
		},
	}
}

/*
Generates the Port Forwarding Rules to Expose a VM on a Port - to the Host on a Port

Usage:

	config := NetworkExposeConfig{
		VM:         "myvm",
		ExternalIP: 191.58.123.44, // External IP ex. personal device
		PortMapping: network.PortMapping{
			Protocol: NetProtocol.TCP,
			HostPort: 9001, // Host Port ex. Port on the Machine running VMs
			VMPort:   8088, // Port of the actual VM
		},
	}

	err := CreateAndSetNetExposeConfig(config)

Additionally this will Cache the Config on Disk - in case we want to enable Automatic Port Management by QEMU/Libvirt System Hooks

To Expose the VM - run commands located in

	data/artifacts/<vmname>/iptables_expose
*/
func CreateAndSetNetExposeConfig(config NetworkExposeConfig) error {
	artifactPath, err := utils.CreateAbsPathFromRoot("data/artifacts/" + config.VM + "/networking/iptables_expose")
	if err != nil {
		log.Printf("Failed to Generate Artifact Path ERROR:%s", err)
		return fmt.Errorf("Failed Path Generation for Artifact")
	}

	log.Printf("Artifact Path - %s", artifactPath)

	fwdingConfig, err := network.GeneratePortForwardingConfigExtractDomainIP(config.VM,
		config.ExternalIP,
		[]network.PortMapping{config.PortMapping},
		nil)
	if err != nil {
		log.Printf("Failed to Generate Config ERROR:%s", err)
		return err
	}

	table := network.CreateTableFromConfig(*fwdingConfig)
	fmt.Println(table)

	exposeCommands := qemu_hooks.HandleForwardingEvent(qemu_hooks.Reconnect, fwdingConfig)

	err = utils.WriteArraytoFile(
		append([]string{table}, exposeCommands...),
		artifactPath)
	if err != nil {
		log.Printf("Failed to Write Generated Expose Commands to Artifact Path. ERROR:%s", err)
		return err
	}

	if err := qemu_hooks.UpdateConfig(*fwdingConfig); err != nil {
		log.Printf("Error writing config: %s", err)
		return err
	}
	return nil

	/*
		If we want to read from curr Config by passing only hostname use this

		cmdsFromConfig, err := qemu_hooks.HandleQemuHookEvent(string(qemu_hooks.Reconnect), config.VM)
		if err != nil {
			log.Printf("Failed to Generate from Current Config File for VM ERROR:%s", err)
			return err
		}

		err = utils.WriteArraytoFile(cmdsFromConfig, artifactPath+"_config")
		if err != nil {
			log.Printf("Failed to Write Generated Expose Commands to Artifact Path. ERROR:%s", err)
			return err
		}
	*/
}

// go run main.go --expose-vm=hadoop --port=8080 --hostport=8000 --external-ip=192.168.1.225

// external_ip defaults to 0.0.0.0
// optionally also accepts --portrange=8080:8088 --hostrange=8100 8108

// Parse into above struct
