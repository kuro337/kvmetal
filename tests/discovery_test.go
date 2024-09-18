package tests

import (
	"log"
	"testing"

	"kvmgo/discovery"
)

func TestConsulDiscoveryFile(t *testing.T) {
	d := discovery.CreateDiscoveryConfig([]string{
		"123.444.111",
		"999.123.99.01",
		"99.12.00.92",
	})

	log.Print(d)

	t.Error("Trigger")
}
