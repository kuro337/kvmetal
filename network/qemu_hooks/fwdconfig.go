package qemu_hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"kvmgo/network"
)

const (
	configFileDir  = "/home/kuro/Documents/Code/Go/kvmgo/data/network/config/"
	configFileName = "kvmfwding_config.json"
)

// WriteConfigToFile updates or adds a new VM configuration.
func WriteConfigToFile(vmConfig network.ForwardingConfig) error {
	configs, err := ReadConfigsFromFile()
	if err != nil {
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

// ClearVMConfig clears the configuration for a specific VM.
func ClearVMConfig(vmName string) error {
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
		return configs, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configs); err != nil {
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
