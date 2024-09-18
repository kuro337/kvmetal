package discovery

import (
	"encoding/json"
	"log"
)

type ConsulConfig struct {
	Datacenter string   `json:"datacenter"`
	DataDir    string   `json:"data_dir"`
	BindAddr   string   `json:"bind_addr"`
	ClientAddr string   `json:"client_addr"`
	RetryJoin  []string `json:"retry_join"`
}

/*
Creates the Discovery File for the VM to be placed in /etc/consul.d/consul-config.json

Usage:

	d := discovery.CreateDiscoveryConfig([]string{
		"123.444.111",
		"999.123.99.01",
		"99.12.00.92",
	})
*/
func CreateDiscoveryConfig(servers []string) string {
	config := ConsulConfig{
		Datacenter: "dc1",
		DataDir:    "/var/consul",
		BindAddr:   "0.0.0.0",
		ClientAddr: "0.0.0.0",
		RetryJoin:  servers,
	}

	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal Consul config: %v", err)
		return ""
	}

	return string(configBytes)
}
