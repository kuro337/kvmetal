package qemu_hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"kvmgo/network"
	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

const (
	configFileDir  = "/home/kuro/Documents/Code/Go/kvmgo/data/network/config/"
	configFileName = "kvmfwding_config.json"
)

// WriteConfigToFile updates or adds a new VM configuration.
func WriteConfigToFile(vmConfig network.ForwardingConfig) error {
	configs, err := ReadConfigsFromFile()
	if err != nil {
		log.Printf("Error during Writing Configs:%s", err)
		// File might not exist, create a new configs if so.
		if os.IsNotExist(err) {
			configs = network.ForwardingConfigs{Configs: []network.ForwardingConfig{}}
		} else {
			return err
		}
	}

	// Add or Update the VM configuration
	updated := false
	for i, config := range configs.Configs {
		if config.VMName == vmConfig.VMName {
			configs.Configs[i] = vmConfig
			updated = true
			break
		}
	}
	if !updated {
		configs.Configs = append(configs.Configs, vmConfig)
	}

	// Write back the updated configs
	return WriteConfigsToFile(configs)
}

// ReadConfigFromFile reads the forwarding configuration from a JSON file.
func ReadVMConfigFromFile(vmName string) (*network.ForwardingConfig, error) {
	var configs network.ForwardingConfigs
	filePath := configFileDir + configFileName

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist can be considered as no config for the VM
			return nil, nil
		}
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configs); err != nil {
		return nil, fmt.Errorf("reading configs from file: %w", err)
	}

	// Look for the VM's configuration by name
	for _, config := range configs.Configs {
		if config.VMName == vmName {
			return &config, nil
		}
	}

	// No configuration found for the VM
	return nil, nil
}

// ClearVMConfig clears the Forwarding Configuration for a specific VM.
func ClearVMForwardingConfig(vmName string) error {
	configs, err := ReadConfigsFromFile()
	if err != nil {
		return err
	}

	// Filter out the VM configuration to remove
	newConfigs := make([]network.ForwardingConfig, 0)
	for _, cfg := range configs.Configs {
		if cfg.VMName != vmName {
			newConfigs = append(newConfigs, cfg)
		}
	}

	configs.Configs = newConfigs

	return WriteConfigsToFile(configs)
}

// ReadConfigsFromFile reads the VM forwarding configurations from a JSON file.
func ReadConfigsFromFile() (network.ForwardingConfigs, error) {
	var configs network.ForwardingConfigs
	filePath := filepath.Join(configFileDir, configFileName)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return an empty configuration if the file does not exist.
			return network.ForwardingConfigs{Configs: []network.ForwardingConfig{}}, nil
		}
		return configs, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configs)
	if err != nil {
		if err == io.EOF {
			// Return an empty configuration for an empty file.
			return network.ForwardingConfigs{Configs: []network.ForwardingConfig{}}, nil
		}
		return configs, fmt.Errorf("reading config from file: %w", err)
	}

	return configs, nil
}

// WriteConfigsToFile writes the VM forwarding configurations to a JSON file.
func WriteConfigsToFile(configs network.ForwardingConfigs) error {
	filePath := filepath.Join(configFileDir, configFileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(configs); err != nil {
		return fmt.Errorf("writing config to file: %w", err)
	}

	return nil
}

// Updates the Config so we can incrementally expose VM's
func UpdateConfig(newConfig network.ForwardingConfig) error {
	configs, err := ReadConfigsFromFile()
	if err != nil {
		return err
	}

	found := false
	for i, config := range configs.Configs {
		if config.VMName == newConfig.VMName {
			mergeConfigs(&configs.Configs[i], newConfig)
			found = true
			break
		}
	}

	if !found {
		configs.Configs = append(configs.Configs, newConfig)
	}

	return WriteConfigsToFile(configs)
}

// Merges new configuration fields into the original configuration without duplicating port mappings or port ranges.
func mergeConfigs(original *network.ForwardingConfig, newConfig network.ForwardingConfig) {
	if newConfig.HostIP != nil {
		original.HostIP = newConfig.HostIP
	}
	if newConfig.PrivateIP != nil {
		original.PrivateIP = newConfig.PrivateIP
	}
	if newConfig.ExternalIP != nil {
		original.ExternalIP = newConfig.ExternalIP
	}

	// Update PortMapping by checking for duplicates
	for _, newPM := range newConfig.PortMap {
		exists := false
		for _, origPM := range original.PortMap {
			if newPM.HostPort == origPM.HostPort && newPM.VMPort == origPM.VMPort && newPM.Protocol == origPM.Protocol {
				exists = true
				break
			}
		}
		if !exists {
			original.PortMap = append(original.PortMap, newPM)
		}
	}

	// Update PortRange by checking for duplicates
	for _, newPR := range newConfig.PortRange {
		exists := false
		for _, origPR := range original.PortRange {
			if newPR.VMStartPort == origPR.VMStartPort && newPR.VMEndPortNum == origPR.VMEndPortNum &&
				newPR.HostStartPortNum == origPR.HostStartPortNum && newPR.HostEndPortNum == origPR.HostEndPortNum &&
				newPR.Protocol == origPR.Protocol {
				exists = true
				break
			}
		}
		if !exists {
			original.PortRange = append(original.PortRange, newPR)
		}
	}
}

// Uses Libvirt Client to get the Domain IP, Gets Host IP, and Writes Default Forwarding Config
func DomainAddForwardingConfigIfRunning(domain string) error {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
	}
	defer conn.Close()

	dom, err := conn.LookupDomainByName(domain)
	if err != nil {
		log.Printf("Failed to get Domain ERROR:%s", err)
		return err
	}
	info, err := dom.GetInfo()
	if err != nil {
		log.Printf("Failed to list domain info ERROR:%s", err)
		return err
	}

	if info.State != libvirt.DOMAIN_RUNNING {
		stateName, stateDesc := utils.ConvertDomainState(info.State)
		infoString := "Domain not detected as Running: " + domain +
			"\nState: " + stateName + " (" + stateDesc + ")\n"
		log.Print(infoString)
	}

	log.Print(utils.TurnSuccess(fmt.Sprintf("Domain %s is Running", domain)))

	domainIP, err := utils.ListDomainIP(conn, dom)
	if err != nil || domainIP == nil {
		log.Printf("Failed to Get IP ERROR:%s", err)
		return err
	}

	hostIP, err := network.GetHostIP()
	if err != nil {
		log.Printf("Failed Getting Host IP. ERROR:%s", err)
	}

	forwadingConfig, err := network.GenerateDefaultPortForwardingConfig(domain, domainIP, net.ParseIP("192.168.1.225"),
		hostIP.IP,
		[]network.PortMapping{
			{Protocol: network.TCP, HostPort: 8080, VMPort: 9999},
			{Protocol: network.TCP, HostPort: 8088, VMPort: 9988},
		}, nil)
	if err != nil {
		log.Printf("Failed to Generate Default Port Forwarding Config. ERROR:%s", err)
	}

	if err := UpdateConfig(*forwadingConfig); err != nil {
		log.Printf("Error writing config: %s", err)
		return err
	}

	cmds, err := HandleQemuHookEvent("start", domain)
	if err != nil {
		log.Printf("Error Creating Forward Hooks for Start Event ERROR:%s", err)
	}

	log.Printf("Successfully Created Default Forwarding Config")

	if err := utils.WriteArraytoFile(cmds, CmdsFilePath); err != nil {
		log.Printf("Failed writing generated forwarding commands to file %s ERROR:%s,", CmdsFilePath, err)
	}
	log.Printf("Successfully Generated Commands Logs file at %s", CmdsFilePath)
	return nil
}
