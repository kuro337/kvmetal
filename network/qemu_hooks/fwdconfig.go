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

// Merges new configuration fields into the original configuration.
func mergeConfigs(original *network.ForwardingConfig, newConfig network.ForwardingConfig) {
	// Only update fields that are explicitly set in newConfig.
	if newConfig.HostIP != nil {
		original.HostIP = newConfig.HostIP
	}
	if newConfig.PrivateIP != nil {
		original.PrivateIP = newConfig.PrivateIP
	}
	if newConfig.ExternalIP != nil {
		original.ExternalIP = newConfig.ExternalIP
	}

	if len(newConfig.PortMap) > 0 {
		original.PortMap = append(original.PortMap, newConfig.PortMap...)
	}
	if len(newConfig.PortRange) > 0 {
		original.PortRange = append(original.PortRange, newConfig.PortRange...)
	}
}

func GenerateDefForwardConf(domain string) error {
	external := net.ParseIP("192.168.1.225")

	conf, err := network.GeneratePortForwardingConfig(domain, external, []network.PortMapping{
		{Protocol: network.TCP, HostPort: 8080, VMPort: 9999},
		{Protocol: network.TCP, HostPort: 8088, VMPort: 9988},
	}, nil)
	if err != nil {
		log.Printf("Failed to Create Default Port Forwarding Config. ERROR:%s", err)
		return err
	}

	if err := UpdateConfig(*conf); err != nil {
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
