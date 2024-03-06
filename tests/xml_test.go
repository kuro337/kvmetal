package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"kvmgo/lib"
)

/* Recommended Way to Get a Domain IP */
func TestGetHostIpRecommended(t *testing.T) {
	domain := "kafka"

	conn, _ := lib.ConnectLibvirt()

	dom, err := conn.GetDomain(domain)
	if err != nil {
		t.Errorf("Failed to Get Domain .ERROR:%s", err)
	}

	ip, err := dom.GetIP()
	if err != nil {
		t.Errorf("Failed to Get Domain IP.ERROR:%s", err)
	}

	log.Printf("Domain IP : %s\n", ip)
}

func TestGetIPFromClient(t *testing.T) {
	conn, _ := lib.ConnectLibvirt()

	results, _ := conn.GetIPFromDHCPLeases("consul")

	for _, x := range results {
		log.Printf("%s\n", x)
	}

	t.Error("Trigger")
}

func TestDumpAllIPDomains(t *testing.T) {
	domain := "consul"

	conn, err := lib.ConnectLibvirt()
	if err != nil {
		t.Fatalf("Failed to connect to libvirt: %v", err)
	}

	dom, _ := conn.GetDomain(domain)

	domipv4, _ := dom.GetIP()

	found := false

	ips, err := dom.GetAllIPs()
	if err != nil {
		t.Fatalf("Failed to get all IPs: %v", err)
	}

	if ips != nil {
		for _, result := range ips.Results {
			log.Printf("Source: %s\n", result.Source)
			for _, ip := range result.IPs {
				if ip == domipv4 {
					found = true
				}
				log.Printf("  IP: %s\n", ip)
			}
		}
	} else {
		t.Log("No IP information available.")
	}
	if !found {
		t.Error("IP from GetIP not matched with IP from All")
	}

	currIp, _ := dom.GetIP()

	log.Print("curr IP cached " + currIp)

	t.Error("trigger")
}

func TestXMLParse(t *testing.T) {
	conn, _ := lib.ConnectLibvirt()

	xml, _ := conn.ParseXML("consul")

	// Convert domcfg to JSON for a readable printout
	prettyJSON, err := json.MarshalIndent(xml, "", "    ")
	if err != nil {
		log.Printf("Failed to generate JSON from domain config: %s", err)
		t.Errorf("Pretty Serialization Failed")
	}

	// Print the indented JSON
	fmt.Printf("%s\n", prettyJSON)

	t.Errorf("TC Failed")
}

// go test -v
// go test
// go test circle_test.go
// go test -v ./mypackage -run TestMyFunction
