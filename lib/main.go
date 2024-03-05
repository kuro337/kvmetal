package lib

import (
	"log"

	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

func main() {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
	}
	defer conn.Close()

	utils.ListAllDomains(conn)

	_ = utils.GetDomainInfo(conn, "worker")

	// doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_SHUTOFF)
	// if err != nil {
	// 	log.Printf("Error Listing %s", err)
	// }

	// fmt.Printf("%d running domains:\n", len(doms))
	// for _, dom := range doms {
	// 	name, err := dom.GetName()
	// 	if err == nil {
	// 		fmt.Printf("  %s\n", name)
	// 	}
	// 	dom.Free()
	// }
}
