package lib

import (
	"fmt"
	"log"

	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

// So this code will generate the Qemu Image? But if I am currently downloading the image file in a Directory
// where will it generate it ? where will the output be? And if I create multiple images - where will it be ?

func main() {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
	}
	defer conn.Close()

	utils.ListAllDomains(conn)

	_ = utils.GetDomainInfo(conn, "worker")

	// Define your storage pool and volume XML here
	poolXML := `<pool type='dir'>
                    <name>default</name>
                    <target>
                        <path>/var/lib/libvirt/images</path>
                    </target>
                </pool>`

	pool, err := conn.StoragePoolCreateXML(poolXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage pool: %v\n", err)
		return
	}
	defer pool.Free()

	volXML := `<volume>
                   <name>new_img.qcow2</name>
                   <allocation>0</allocation>
                   <capacity unit="G">20</capacity>
                   <target>
                       <format type='qcow2'/>
                   </target>
               </volume>`

	vol, err := pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		fmt.Printf("Failed to create storage volume: %v\n", err)
		return
	}
	defer vol.Free()

	fmt.Println("Image created successfully")

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

/*




 */
