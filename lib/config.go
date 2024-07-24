package lib

import (
	"fmt"
	"log"

	ldom "kvmgo/lib/domain"

	"libvirt.org/go/libvirt"
)

// Configuring the Virtual Machine
// vmConfig.SetName("testVM").
// SetMemory(2048).
// SetCores(2).
// SetBaseImage("/var/lib/libvirt/images/base.img").
// SetUserDataPath("/var/lib/libvirt/images/userdata.img").
// SetNetwork("default").
// SetOSVariant("ubuntu18.04")

type VMConfig struct {
	Name         string
	Memory       int
	CPUCores     int
	DiskPath     string
	UserDataPath string
	Network      string
	OSVariant    string
}

func NewVMConfig(name string) *VMConfig {
	return &VMConfig{Name: name}
}

func (c *VMConfig) SetName(name string) *VMConfig {
	c.Name = name
	return c
}

func (c *VMConfig) SetMemory(memory int) *VMConfig {
	c.Memory = memory
	return c
}

func (c *VMConfig) SetCores(cores int) *VMConfig {
	c.CPUCores = cores
	return c
}

// Set the Base Image Path
func (c *VMConfig) SetBaseImage(baseImgPath string) *VMConfig {
	c.DiskPath = baseImgPath
	return c
}

func (c *VMConfig) SetUserDataPath(userDataPath string) *VMConfig {
	c.UserDataPath = userDataPath
	return c
}

func (c *VMConfig) SetNetwork(network string) *VMConfig {
	c.Network = network
	return c
}

func (c *VMConfig) SetOSVariant(osVariant string) *VMConfig {
	c.OSVariant = osVariant
	return c
}

/*	cmdArgs := []string{
	"--name", vm.VMName,
	"--virt-type", "kvm",
	"--memory", fmt.Sprint(vm.Memory),
	"--vcpus", fmt.Sprint(vm.CPUCores),
	"--disk", "path=" + generatedVmImg + ",device=disk",
	"--disk", "path=" + vm_userdata_img + ",format=raw",
	"--graphics", "none",
	"--boot", "hd,menu=on",
	"--network", "network=default",
	"--os-variant", "ubuntu18.04",
	"--noautoconsole",
} */

func xmlName(name string) string {
	return fmt.Sprintf("<name>%s</name>", name)
}

func xmlCpuMem(vcpu, mem int) string {
	return fmt.Sprintf(`<memory unit='MiB'>%d</memory>
  <vcpu placement='static'>%d</vcpu>`, mem, vcpu)
}

func xmlPrimaryDisk(path string) string {
	return fmt.Sprintf(`<disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>`, path)
}

func xmlCloudInitDisk(path string) string {
	return fmt.Sprintf(`<disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='%s'/>
      <target dev='hda' bus='ide'/>
      <readonly/>
    </disk>`, path)
}

func xmlNetwork(iface string) string {
	return fmt.Sprintf(`<interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>`, iface)
}

func (c *VMConfig) GenerateDomainXML() string {
	return fmt.Sprintf(`
<domain type='kvm'>
  <name>%s</name>
  <memory unit='MiB'>%d</memory>
  <vcpu placement='static'>%d</vcpu>
  <os>
    <type arch='x86_64' machine='pc-i440fx-2.9'>hvm</type>
    <boot dev='hd'/>
  </os>
  <features>
    <acpi/>
    <apic/>
  </features>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='%s'/>
      <target dev='hda' bus='ide'/>
      <readonly/>
    </disk>
    <interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>
    <graphics type='vnc' port='-1'/>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
  </devices>
</domain>`, c.Name, c.Memory, c.CPUCores, c.DiskPath, c.UserDataPath, c.Network)
}

// Create the VM with the config
func (vm *VMConfig) CreateAndStartVM(client *libvirt.Connect) error {
	domainXML := vm.GenerateDomainXML()
	domain, err := client.DomainCreateXML(domainXML, 0)
	if err != nil {
		return fmt.Errorf("failed to create VM: %v", err)
	}
	defer domain.Free()

	// DomainCreateXML creates and starts it

	//	if err := domain.Create(); err != nil {
	//		return fmt.Errorf("failed to start VM: %v", err)
	//	}

	log.Printf("VM %s created and started", vm.Name)
	return nil
}

// Delete the VM by the Domain Name
func DeleteVM(client *libvirt.Connect, domain string) error {
	dom, err := ldom.GetDomain(client, domain)

	if dom == nil {
		return fmt.Errorf("failed to get domain: %v", err)
	}

	if err := dom.Delete(); err != nil {
		return fmt.Errorf("failed to Delete VM:%s", err.Error())
	}

	return nil
}
