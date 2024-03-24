package lib

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

type Domain struct {
	name        string
	domain      *libvirt.Domain
	ip          net.IP
	ipRetrieved time.Time
}

// GetIPLibvirtRetry attempts to retry GetIPLibvirt with specified retry logic
func GetIPLibvirtRetry(domain string) (string, error) {
	retryFunc := func() (string, error) {
		return GetIPLibvirt(domain) // Your function that gets the IP
	}

	condition := func(err error) bool {
		return err != nil && (strings.Contains(err.Error(), "Domain not found:") || strings.Contains(err.Error(), "no IPv4 address found for domain"))
	}

	// Attempt to retry up to 5 times with a 2-second fixed delay between retries
	return utils.RetryUntilString(retryFunc, condition, 5, 5*time.Second)
}

func GetIPLibvirt(domain string) (string, error) {
	conn, _ := ConnectLibvirt()

	dom, err := conn.GetDomain(domain)
	if err != nil {
		return "", err
	}

	ip, err := dom.GetIP()
	if err != nil {
		return "", err
	}

	return ip, nil
}

// Gets the IP or Pulls it using Libvirt if it is unknown
func (d *Domain) GetIP() (string, error) {
	if d.ip != nil {
		timeSinceRetrieved := time.Since(d.ipRetrieved)
		humanReadable := utils.HumanizeDuration(timeSinceRetrieved)
		log.Printf("Latest IP Retrieved %s ago", humanReadable)
		return d.ip.String(), nil
	}

	ip, err := d.PullIP()
	if err != nil {
		log.Printf("Failed to Pull IP for Domain. ERROR:%s", err)
		return "", err
	}
	d.ip = net.ParseIP(ip)
	d.ipRetrieved = time.Now()
	return d.ip.String(), nil
}

/* Pulls all IP Addresses associated with the Domain */
func (d *Domain) PullIP() (string, error) {
	ifaces, err := d.domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		return "", fmt.Errorf("failed to list interface addresses: %v", err)
	}

	for _, iface := range ifaces {
		for _, addr := range iface.Addrs {
			if addrType := addr.Type; addrType == libvirt.IP_ADDR_TYPE_IPV4 {
				return addr.Addr, nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found for domain")
}
