package utils

import (
	"fmt"
	"log"
	"net"

	"libvirt.org/go/libvirt"
)

func ListAllDomains(conn *libvirt.Connect) {
	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		log.Printf("Error Listing %s", err)
	}

	fmt.Printf(TurnBold("%d running domains:\n"), len(doms))
	for _, dom := range doms {
		name, err := dom.GetName()
		if err == nil {
			fmt.Printf("  %s\n", name)
		}
		_ = dom.Free()
	}

	doms, err = conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_SHUTOFF)
	if err != nil {
		log.Printf("Error Listing %s", err)
	}

	fmt.Printf(TurnBold("%d shutdown domains:\n"), len(doms))
	for _, dom := range doms {

		_ = dom.Shutdown()

		name, err := dom.GetName()
		if err == nil {
			fmt.Printf("  %s\n", name)
		}
		_ = dom.Free()
	}
}

func ConvertDomainState(domainState libvirt.DomainState) (string, string) {
	switch domainState {
	case libvirt.DOMAIN_NOSTATE:
		return "No State", "State is unknown"
	case libvirt.DOMAIN_CRASHED:
		return "Crashed", "Domain has crashed"
	case libvirt.DOMAIN_BLOCKED:
		return "Blocked", "Domain currently blocked on a resource"
	case libvirt.DOMAIN_PAUSED:
		return "Paused", "Domain Paused by User"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "Susepnded", "Domain suspended by power management"
	case libvirt.DOMAIN_RUNNING:
		return "Running", "Domain is Running"
	case libvirt.DOMAIN_SHUTDOWN:
		return "Shutdown", "Domain is being shut down"
	case libvirt.DOMAIN_SHUTOFF:
		return "Shut Off", "Domain is Shut Off"
	default:
		return "Unknown", "State is unknown"
	}
}

func ListDomainIP(conn *libvirt.Connect, dom *libvirt.Domain) (net.IP, error) {
	netinfo, err := dom.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		log.Printf("Failed to get Net Info ERROR:%s", err)
		return nil, err
	}

	for _, ni := range netinfo {
		log.Printf("%+v", ni)
		for _, domnetinfo := range ni.Addrs {

			log.Printf("IP Addr:%s Prefix:%d", domnetinfo.Addr, domnetinfo.Prefix)
			ip := net.ParseIP(domnetinfo.Addr)
			if ip != nil {
				return ip, nil
			}

		}
	}
	return nil, fmt.Errorf("Failed to Obtain IP Addr from Libvirt Domain API")
}

func ListDomainXML(conn *libvirt.Connect, domain *libvirt.Domain) error {
	flags := libvirt.DOMAIN_XML_SECURE | libvirt.DOMAIN_XML_INACTIVE
	xmlDesc, err := domain.GetXMLDesc(flags)
	if err != nil {
		log.Printf("Failed to get domain XML description: %v", err)
		return err
	}

	fmt.Println("Domain XML Description:")
	fmt.Println(xmlDesc)

	return nil
}

func WaitUntilReady(domains []string) error {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return fmt.Errorf("Error Connecting %s", err)
	}
	defer conn.Close()

	ListAllDomains(conn)

	return nil
}

// GetDomainInfo prints the Domains Info such as current state and Memory and CPU allocated
func GetDomainInfo(conn *libvirt.Connect, domain string) error {
	dom, err := conn.LookupDomainByName(domain)
	if err != nil {
		log.Printf("Failed to get Domain ERROR:%s", err)
		return err
	}

	// ListDomainXML(conn, dom)

	// guestinfo, err := dom.GetGuestInfo(libvirt.DOMAIN_GUEST_INFO_OS|libvirt.DOMAIN_GUEST_INFO_TIMEZONE, 0) // Assuming 0 for flags; adjust as needed
	// if err != nil {
	// 	log.Printf("Failed to get guest info: %v", err)
	// 	return err
	// }

	// // Process info here

	// fmt.Println("Guest Info:", guestinfo)

	info, err := dom.GetInfo()
	if err != nil {
		log.Printf("Failed to list domain info ERROR:%s", err)
		return err
	}

	stateName, stateDesc := ConvertDomainState(info.State)

	infoString := "Domain Information: " + domain +
		"\nState: " + stateName + " (" + stateDesc + ")\n" +
		"Max Memory: " + fmt.Sprintf("%d", info.MaxMem) + " KiB\n" +
		"Memory: " + fmt.Sprintf("%d", info.Memory) + " KiB\n" +
		"Number of Virtual CPUs: " + fmt.Sprintf("%d", info.NrVirtCpu) + "\n" +
		"CPU Time: " + fmt.Sprintf("%d", info.CpuTime) + " ns\n"

	log.Print(infoString)
	return nil
}
