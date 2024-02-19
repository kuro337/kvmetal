package vm

import (
	"log"

	"libvirt.org/go/libvirt"
)

// ShutdownVM shuts off a Virtual Machine
func ShutdownVM(conn *libvirt.Connect, domain libvirt.Domain) {
	err := domain.Shutdown()
	name, _ := domain.GetName()

	if err == nil {
		log.Printf("Successfully shutdown VM %s ERROR:%s", name, err)
	}
	if err != nil {
		log.Printf("Failed to shutdown VM %s ERROR:%s", name, err)
	}
}

// ClearVM removes a VM and all the associated Metadata and Snapshots
func ClearVM(domain *libvirt.Domain) error {
	// bitwise to combine all flags
	flags := libvirt.DomainUndefineFlagsValues(
		libvirt.DOMAIN_UNDEFINE_MANAGED_SAVE |
			libvirt.DOMAIN_UNDEFINE_SNAPSHOTS_METADATA |
			libvirt.DOMAIN_UNDEFINE_NVRAM |
			libvirt.DOMAIN_UNDEFINE_CHECKPOINTS_METADATA,
	)

	if err := domain.UndefineFlags(flags); err != nil {
		log.Printf("Failed to undefine VM and remove all storage. Error: %v", err)
		return err
	}

	log.Printf("VM has been successfully undefined and all storage removed.")
	return nil
}
