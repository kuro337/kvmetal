package qemu_hooks

import "fmt"

// SubstituteChainName trims the name if the VM/Domain name is >28 chars. (IPTables Limit -> 28 chars)
func SubstituteChainName(vmName string, index int) string {
	const maxLen = 28
	prefix := "DNAT-"

	maxVmNameLen := maxLen - len(prefix) - len(fmt.Sprint(index)) - 1

	if len(vmName) > maxVmNameLen {
		vmName = vmName[:maxVmNameLen]
	}
	return fmt.Sprintf("%s%s-%d", prefix, vmName, index)
}
