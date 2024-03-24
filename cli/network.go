package cli

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"kvmgo/lib"
	"kvmgo/network"
	"kvmgo/network/qemu_hooks"
	"kvmgo/utils"
)

type NetworkExposeConfig struct {
	VM          string
	ExternalIP  net.IP
	PortMapping network.PortMapping
	PortRange   network.PortRange
}

/*
Quick End to End Flow:

	go run main.go --expose-vm=kafka \
	--port=9092 \
	--hostport=8071 \
	--external-ip=192.168.1.225 \
	--protocol=tcp

From the VM launch a nc srvr:

	nc -l 9092

Host <-> VM

	nc test.kuro.com 9092 (write to stdin)

External Device <-> Host

	nc 192.168.1.10 8071

!NOTE: Ensure the External Machine and Host are on the SAME NETWORK (eth,wifi,etc.)
*/
func PrintNetworkQuickHelp(vmName, vmIp string, vmPort, hostPort int, hostIp string) {
	log.Printf("%s\n    %s\n    %s\n    %s\n    %s\n     Checking Fwding Rules iptables:\nsudo iptables -t nat -L -n -v | grep %d\n",
		utils.TurnBoldColor("Port Forwarding Quick Help:", utils.COOLBLUE),
		utils.TurnUnderline("Launch a nc server on the VM")+utils.TurnBold(fmt.Sprintf("\n    nc -l %d", vmPort)),
		utils.TurnUnderline("Host <-> VM")+utils.TurnBold(fmt.Sprintf("\n    nc %s.kuro.com %d", vmName, vmPort)),
		utils.TurnUnderline("External Device <-> Host")+utils.TurnBold(fmt.Sprintf("\n    nc %s %d", hostIp, hostPort)),
		utils.TurnUnderline("Connect to the VM server from Host")+utils.TurnBold(fmt.Sprintf("\n    nc 192.168.122.x %d", vmPort)),
		vmPort,
	)
}

/*
Generates Networking Config for VM and checks for Existing one - then generates the Forwarding Rules and Writes it to the Artifact Location for the VM

Usage:

	go run main.go --expose-vm=kafka \

--port=9095 \
--hostport=9094 \
--external-ip=192.168.1.225 \
--protocol=tcp

err := HandleVMNetworkExposure("vmname",9095,9094,"192.168.1.225","tcp")
*/
func HandleVMNetworkExposure(
	vmName string,
	vmPort, hostPort int,
	externalIp string, protocol string,
) error {
	netConfig := ParseNetExposeFlags(vmName, vmPort, hostPort, externalIp, protocol)

	if netConfig != nil {
		if err := CreateAndSetNetExposeConfig(*netConfig); err != nil {
			log.Printf("Failed To Create Forwarding Config ERROR:%s,", err)
			return err
		}
	}
	return nil
}

func ParseNetExposeFlags(vmName string, vmPort, hostPort int, externalIp string, protocol string) *NetworkExposeConfig {
	externalIP := net.ParseIP(externalIp)

	if externalIP == nil {
		log.Print(utils.TurnError(fmt.Sprintf("Failed to Parse External IP %s", externalIP)))
		return nil
	}

	return &NetworkExposeConfig{
		VM:         vmName,
		ExternalIP: externalIP,
		PortMapping: network.PortMapping{
			Protocol: network.NetProtocol(protocol),
			HostPort: int(hostPort),
			VMPort:   int(vmPort),
		},
	}
}

/*
Generates the Port Forwarding Rules to Expose a VM on a Port - to the Host on a Port

Usage:

	config := NetworkExposeConfig{
		VM:         "myvm",
		ExternalIP: 191.58.123.44, // External IP ex. personal device
		PortMapping: network.PortMapping{
			Protocol: NetProtocol.TCP,
			HostPort: 9001, // Host Port ex. Port on the Machine running VMs
			VMPort:   8088, // Port of the actual VM
		},
	}

	err := CreateAndSetNetExposeConfig(config)

Additionally this will Cache the Config on Disk - in case we want to enable Automatic Port Management by QEMU/Libvirt System Hooks

To Expose the VM - run commands located in

	data/artifacts/<vmname>/iptables_expose
*/
func CreateAndSetNetExposeConfig(config NetworkExposeConfig) error {
	artifactPath, err := utils.CreateAbsPathFromRoot("data/artifacts/" + config.VM + "/networking/iptables_expose")
	if err != nil {
		log.Printf("Failed to Generate Artifact Path ERROR:%s", err)
		return fmt.Errorf("Failed Path Generation for Artifact")
	}

	log.Printf("Artifact Path - %s", artifactPath)

	fwdingConfig, err := network.GeneratePortForwardingConfigExtractDomainIP(config.VM,
		config.ExternalIP,
		[]network.PortMapping{config.PortMapping},
		nil)
	if err != nil {
		log.Printf("Failed to Generate Config ERROR:%s", err)
		return err
	}

	table := network.CreateTableFromConfig(*fwdingConfig)
	fmt.Println(table)

	exposeCommands := qemu_hooks.HandleForwardingEvent(qemu_hooks.Reconnect, fwdingConfig)

	err = utils.WriteArraytoFile(
		append([]string{table}, exposeCommands...),
		artifactPath)
	if err != nil {
		log.Printf("Failed to Write Generated Expose Commands to Artifact Path. ERROR:%s", err)
		return err
	}

	if err := qemu_hooks.UpdateConfig(*fwdingConfig); err != nil {
		log.Printf("Error writing config: %s", err)
		return err
	}

	PrintNetworkQuickHelp(fwdingConfig.VMName,
		fwdingConfig.PrivateIP.String(),
		config.PortMapping.VMPort,
		config.PortMapping.HostPort,
		fwdingConfig.HostIP.String())

	return nil

	/*
		If we want to read from curr Config by passing only hostname use this

		cmdsFromConfig, err := qemu_hooks.HandleQemuHookEvent(string(qemu_hooks.Reconnect), config.VM)
		if err != nil {
			log.Printf("Failed to Generate from Current Config File for VM ERROR:%s", err)
			return err
		}

		err = utils.WriteArraytoFile(cmdsFromConfig, artifactPath+"_config")
		if err != nil {
			log.Printf("Failed to Write Generated Expose Commands to Artifact Path. ERROR:%s", err)
			return err
		}
	*/
}

func CreateUpdateForwarding(
	domain string,
	vmPort, hostPort int,
	fwdingConfig network.ForwardingConfig,
) error {
	artifactPath, err := utils.CreateAbsPathFromRoot("data/artifacts/" + domain + "/networking/iptables_expose")
	if err != nil {
		log.Printf("Failed to Create Abs Artifact Path ERROR:%s", err)
		return err
	}

	table := network.CreateTableFromConfig(fwdingConfig)
	fmt.Println(table)

	exposeCommands := qemu_hooks.HandleForwardingEvent(qemu_hooks.Reconnect, &fwdingConfig)

	err = utils.WriteArraytoFile(
		append([]string{table}, exposeCommands...),
		artifactPath)
	if err != nil {
		log.Printf("Failed to Write Generated Expose Commands to Artifact Path. ERROR:%s", err)
		return err
	}

	if err := qemu_hooks.UpdateConfig(fwdingConfig); err != nil {
		log.Printf("Error writing config: %s", err)
		return err
	}

	PrintNetworkQuickHelp(fwdingConfig.VMName,
		fwdingConfig.PrivateIP.String(),
		vmPort,
		hostPort,
		fwdingConfig.HostIP.String())

	return nil
}

func WaitForVMThenGenerateFwdingConfig(
	ctx context.Context,
	wg *sync.WaitGroup,
	domain string,
	vmPort, hostPort int,
	externalIp string,
	protocol string,
) error {
	defer wg.Done()

	log.Printf("Sleeping for 5 seconds before attempting to get IP")
	time.Sleep(5 * time.Second)

	log.Printf("Trying to obtain Domain IP")
	ip, err := lib.GetIPLibvirtRetry(domain)
	if err != nil {
		log.Printf("Could not get IP for Domain using Retries. ERROR:%s %s", err, ip)
		return err
	}
	log.Printf("IP Successfully Obtained : %s", ip)

	extIp := net.ParseIP(externalIp)
	domainIp := net.ParseIP(ip)
	hostIP, err := network.GetHostIP()
	if err != nil {
		log.Print(utils.TurnError("Failed to get Host IP"))
	}

	portMapping := network.PortMapping{
		Protocol: network.NetProtocol(protocol),
		HostPort: int(hostPort),
		VMPort:   int(vmPort),
	}

	fwdConfig := network.CreatePortForwardingConfig(domain, "virbr0",
		domainIp, hostIP.IP, extIp, []network.PortMapping{portMapping}, nil)

	err = CreateUpdateForwarding(domain, vmPort, hostPort, fwdConfig)
	if err != nil {
		log.Printf("Failed to Create and Update Forwarding Config. ERROR:%s", err)
		return err
	}

	log.Printf("Successsfully Generated Forwarding Config")

	return nil
}

// go run main.go --expose-vm=hadoop --port=8080 --hostport=8000 --external-ip=192.168.1.225

// external_ip defaults to 0.0.0.0
// optionally also accepts --portrange=8080:8088 --hostrange=8100 8108

// Parse into above struct

/*
   Launch a nc server on the VM

   nc -l 9092

   Host <-> VM

   nc test.kuro.com 9092

   External Device <-> Host

   nc 192.168.1.10 9092

   Connect to the VM server from Host

   nc 192.168.122.x 9092

*/
