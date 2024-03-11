package tests

import (
	"fmt"
	"testing"

	"kvmgo/configuration/presets"
)

func TestRedPandaConfig(t *testing.T) {
	vmIP := "redpanda.kuro.com"
	vmPort := "9095"
	hostIP := "192.168.1.10"
	hostPort := "8090"

	config := presets.GenerateRedpandaUserdata(vmIP, vmPort, hostIP, hostPort)
	fmt.Println(config)

	fullConfig := presets.CreateRedpandaUserdata("ubuntu", "password", "redpanda", "ssheky12341413123",
		vmIP, vmPort, hostIP, hostPort)

	fmt.Println(fullConfig)

	t.Error("Trigger")

	/*
		go run main.go --expose-vm=redpanda \
		--port=9095 \
		--hostport=8090 \
		--external-ip=192.168.1.225 \
		--protocol=tcp

	*/
}
