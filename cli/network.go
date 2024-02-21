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

func ParseNetExposeFlags(vmName string, vmPort, hostPort int, externalIp string, protocol string) *NetworkExposeConfig {
	externalIP := net.ParseIP(externalIp)

	if externalIP == nil {
		log.Printf(utils.TurnError(fmt.Sprintf("Failed to Parse External IP %s", externalIP)))
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

func CreateAndSetNetExposeConfig(config NetworkExposeConfig) error {
	fwdingConfig, err := network.GeneratePortForwardingConfigExtractDomainIP(config.VM,
		config.ExternalIP,
		[]network.PortMapping{config.PortMapping},
		nil)
	if err != nil {
		log.Printf("Failed to Generate Config ERROR:%s", err)
		return fmt.Errorf("Failed to Generate Config ERROR:%s", err)
	}

	if err := qemu_hooks.UpdateConfig(*fwdingConfig); err != nil {
		log.Printf("Error writing config: %s", err)
		return err
	}
	return nil
}

// go run main.go --expose-vm=hadoop --port=8080 --hostport=8000 --external-ip=192.168.1.225

// external_ip defaults to 0.0.0.0
// optionally also accepts --portrange=8080:8088 --hostrange=8100 8108

// Parse into above struct
