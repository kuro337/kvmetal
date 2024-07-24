package domain

import (
	"fmt"
	"log"
	"net"
	"time"

	"kvmgo/network"
	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

type Domain struct {
	Name        string
	domain      *libvirt.Domain
	ip          net.IP
	ipRetrieved time.Time
}

func NewDomainAwait(conn *libvirt.Connect, domain string) (*Domain, error) {
	retries := 8
	delay := 5
	i := 0

	dom, err := conn.LookupDomainByName(domain)
	if err != nil {
		log.Printf("Failed to get Domain ERROR:%s", err)
		return nil, err
	}
	d := &Domain{Name: domain, domain: dom}

	for i < retries {
		ip, err := d.IP()
		if err == nil {
			log.Printf("Domain IP: Retreived domain.go %s\n", ip)
			return d, nil
		}
		i++
		wait := delay + (1 << i)
		log.Printf("Failed to get Domain IP. Retry #%d : %d seconds. %s\n", i, wait, err)
		time.Sleep(time.Duration(wait) * time.Second)

		return d, nil

	}

	return nil, fmt.Errorf("Timed out getting ip for domain: %s", domain)
}

// func NewDomain(conn *libvirt.Connect, domain string) (*Domain, error) {

func NewDomainWrapper(name string, conn *libvirt.Connect, domain *libvirt.Domain) *Domain {
	return &Domain{Name: name, domain: domain}
}

func GetDomain(conn *libvirt.Connect, domain string) (*Domain, error) {
	return NewDomain(conn, domain)
}

func NewDomain(conn *libvirt.Connect, domain string) (*Domain, error) {
	dom, err := conn.LookupDomainByName(domain)
	// DOMAIN_GUEST_INFO_INTERFACES
	//  dom.GetGuestInfo()
	if err != nil {
		log.Printf("Failed to get Domain ERROR:%s", err)
		return nil, err
	}

	d := &Domain{Name: domain, domain: dom}

	ip, err := d.IP()
	if err != nil {
		// return nil, fmt.Errorf("Failed to get Domain IP: %s\n", err)
		return d, fmt.Errorf("Failed to get Domain IP: %s\n", err)
	}

	log.Printf("Domain IP: Retreived domain.go %s\n", ip)

	return d, nil
}

// IP gets the IP
func (d *Domain) IP() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (d *Domain) Delete() error {
	domain := d.domain

	active, err := domain.IsActive()
	if err != nil {
		return fmt.Errorf("failed to check if domain is active: %v", err)
	}

	// Check if the domain is persistent
	persistent, err := domain.IsPersistent()
	if err != nil {
		return fmt.Errorf("failed to check if domain is persistent: %v", err)
	}
	log.Printf("Domain %s is persistent: %v", d.Name, persistent)

	if active { // If the domain is running, stop it

		err = domain.Shutdown()
		if err != nil {
			return fmt.Errorf("failed to initiate shutdown: %v", err)
		}
		log.Printf("Initiated shutdown for domain %s", d.Name)

		err = domain.Destroy()
		if err != nil {
			return fmt.Errorf("failed to stop domain: %v", err)
		}
		log.Printf("Domain %s stopped", d.Name)
	}

	err = domain.Undefine()
	if err != nil {
		return fmt.Errorf("failed to undefine domain: %v", err)
	}
	log.Printf("Domain %s undefined", d.Name)

	return nil
}

/*
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


func GetIPLibvirt( 	conn    *libvirt.Connect,  domain string) (string, error) {
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
*/

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

/*
DOMAIN_NOSTATE     = DomainState(C.VIR_DOMAIN_NOSTATE)
DOMAIN_RUNNING     = DomainState(C.VIR_DOMAIN_RUNNING)
DOMAIN_BLOCKED     = DomainState(C.VIR_DOMAIN_BLOCKED)
DOMAIN_PAUSED      = DomainState(C.VIR_DOMAIN_PAUSED)
DOMAIN_SHUTDOWN    = DomainState(C.VIR_DOMAIN_SHUTDOWN)
DOMAIN_CRASHED     = DomainState(C.VIR_DOMAIN_CRASHED)
DOMAIN_PMSUSPENDED = DomainState(C.VIR_DOMAIN_PMSUSPENDED)
DOMAIN_SHUTOFF     = DomainState(C.VIR_DOMAIN_SHUTOFF)
*/

func (d *Domain) IsRunning() (bool, error) {
	info, err := d.domain.GetInfo()
	if err != nil {
		return false, fmt.Errorf("Error:%s", err)
	}

	log.Printf("Domain %s State: %+v\n", d.Name, info)

	return info.State == libvirt.DOMAIN_RUNNING, nil
}

func (d *Domain) GetInfo() error {
	info, err := d.domain.GetInfo()
	if err != nil {
		return fmt.Errorf("Error:%s", err)
	}

	log.Printf("Domain %s State: %+v\n", d.Name, info)
	return nil
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

// Use this everywhere for SSH Clients
func (d *Domain) NewSSHClient() (*network.VMClient, error) {
	log.Printf("Using ip for ssh client: %s\n", d.ip.String())
	client, err := network.NewInsecureSSHClientVM(d.Name, d.ip.String(), "ubuntu", "password")
	if err != nil {
		return nil, fmt.Errorf("Failed to create client Error:%s", err)
	}

	return client, nil
}
