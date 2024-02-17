package network

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"

	"kvmgo/utils"
)

/*
CHECK docs/network/hostname.md for Hostname not showing up Solutions:

sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml hadoop | awk -F"'" '/mac address/{print $2}') | awk '{print $1}'

sudo iptables -t nat -L PREROUTING -n -v --line-number
sudo iptables -L FORWARD -nv --line-number

# Leases stay active until the Expiry-time
sudo virsh net-dhcp-leases default

virsh net-dumpxml default

# from the VM run this TO make sure the hostname is set to the right one
sudo hostnamectl set-hostname hadoop

sudo systemctl restart network-manager
sudo systemctl restart libvirtd

# To Clear Forwarding Rules we dont need anymore

sudo iptables -D FORWARD 1
sudo iptables -D FORWARD 1  # Note: After deleting the first rule, the next rule becomes the first.

# Rules with Active Traffic such as these do not touch
7     565K   13G LIBVIRT_FWX



*/

const sampleQemuHooksFile = `#!/bin/bash
	
v=$(/sbin/iptables -L FORWARD -n -v | /usr/bin/grep 192.168.122.0/24 | /usr/bin/wc -l)

[ $v -le 2 ] && /sbin/iptables -I FORWARD 1 -o virbr0 -m state -s 192.168.1.0/24 -d 192.168.122.0/24 --state NEW,RELATED,ESTABLISHED -j ACCEPT
`

func CreateQemuHooksFile() string {
	hostIP, _ := GetHostIP()

	libvirtIpSubnet, err := GetLibvirtIpSubnet()
	if err != nil {
		log.Printf("%s", err)
	}

	log.Printf("Libvirt IP Subnet %s", libvirtIpSubnet)

	qemuHooksFile := ConstructQemuHooksFile(hostIP.String(), libvirtIpSubnet)

	utils.LogDottedLineDelimitedText(qemuHooksFile)

	return qemuHooksFile
}

func ConstructQemuHooksFile(hostIpWithSubnet, libvirtIpSubnet string) string {
	return fmt.Sprintf(`#!/bin/bash	
v=$(/sbin/iptables -L FORWARD -n -v | /usr/bin/grep %s | /usr/bin/wc -l)

[ $v -le 2 ] && /sbin/iptables -I FORWARD 1 -o virbr0 -m state -s %s -d %s --state NEW,RELATED,ESTABLISHED -j ACCEPT
	`, libvirtIpSubnet, hostIpWithSubnet, libvirtIpSubnet)
}

/*
CreateUfwBeforeRule creates the Port Forwarding Rule to expose the VM to external guests within the same network

		// Below Rule at the top of the file will expose the VM at 192.168.122.109:9999 on the network bridge through the host machine's port 8888

	  /etc/ufw/before.rules

		*nat
		:PREROUTING ACCEPT [0:0]
		-A PREROUTING -p tcp --dport 8888 -j DNAT --to-destination 192.168.122.109:9999 -m comment --comment "Testing port 9999 of vm from ubuntu host 8888"
		COMMIT

- After adding this to the file we must reload the firewall

	sudo ufw status # check if uncomplicated firewall is active
	sudo ufw enable

	bash /etc/libvirt/hooks/qemu // load new rule
	ufw reload

	reboot // or reboot the host
*/
func CreateUfwBeforeRule(vmIpAddr, vmExposePort, hostPort, comment string) string {
	return fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s -m comment --comment \"%s\"", hostPort, vmIpAddr, vmExposePort, comment)
}

// PrivateIPAddrAllVMs parses libvirtd output for DHCP leases and gets the IP Subnet
func PrivateIPAddrAllVMs(print bool) []IPAddressWithSubnet {
	output, _ := utils.ExecCmd("virsh net-dhcp-leases default", false)

	ipAddresses := ParseIpAddrWithSubnet(output)

	if print {
		for _, addr := range ipAddresses {
			log.Printf("%s/%d\n", addr.IP, addr.Subnet)
		}
	}

	return ipAddresses
}

// VMIpAddrInfoList returns the info for all Virtual Machines managed by the host
func VMIpAddrInfoList(print bool) []VMLeaseInfo {
	output, _ := utils.ExecCmd("virsh net-dhcp-leases default", false)

	// Use GetVMLeaseInfo to find and return all lease information in the output.
	leaseInfo := GetVMLeaseInfo(output)

	if print {
		for _, info := range leaseInfo {
			log.Printf("IP: %s/%s, Hostname: %s, MAC: %s, Protocol: %s\n",
				info.IP, info.Subnet, info.Hostname, info.MAC, info.Protocol)
		}
	}

	return leaseInfo
}

func GetVMLeaseInfo(output string) []VMLeaseInfo {
	var leaseInfo []VMLeaseInfo
	leaseRegex := regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/(\d+)\s+(\S+)`)

	matches := leaseRegex.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		ip := net.ParseIP(match[4])
		if ip != nil {
			leaseInfo = append(leaseInfo, VMLeaseInfo{
				IP:       ip,
				Subnet:   match[5],
				Hostname: match[6],
				MAC:      match[1],
				Protocol: match[3],
			})
		}
	}

	return leaseInfo
}

// ParseIpAddrWithSubnet parses IP addr + Subnet from a string
func ParseIpAddrWithSubnet(output string) []IPAddressWithSubnet {
	var ipAddresses []IPAddressWithSubnet
	ipRegex := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/(\d+)\b`)

	matches := ipRegex.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		ip := net.ParseIP(match[1])
		subnet, err := strconv.Atoi(match[2])
		if err != nil {
			log.Printf("Failed to convert Subnet to valid int")
			continue
		}
		if ip != nil {
			ipAddresses = append(ipAddresses, IPAddressWithSubnet{
				IP:     ip,
				Subnet: subnet,
			})
		}
	}

	return ipAddresses
}

// GetHostIP finds the host's primary IP address in CIDR notation and optionally prints it.
// It returns the first non-loopback IPv4 address found with its subnet mask, which is often used by the default network interface.
func GetHostIP() (*IPAddressWithSubnet, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			// Interface is down or it is a loopback; skip it.
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP
				if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
					ones, _ := v.Mask.Size() // Correctly handle the multiple return values here

					return &IPAddressWithSubnet{
						IP:     ip,
						Subnet: ones,
					}, nil

					// return cidr, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no suitable IP address with subnet found")
}

/*

These are marked Unsafe so SSH is preferable!


Meant to be run from guest VM

sudo apt-get install qemu-guest-agent

systemctl start qemu-guest-agent
systemctl enable qemu-guest-agent

sudo virsh -c qemu:///system qemu-agent-command kubecontrol \
  '{"execute": "guest-exec", "arguments": { "path": "/usr/bin/ls", "arg": [ "/" ], "capture-output": true }}'

{"return":{"pid":14925}}


virsh -c qemu:///system qemu-agent-command kubecontrol \
  '{"execute": "guest-exec-status", "arguments": { "pid": 14925 }}'

will return {"return":{"exitcode":0,"out-data":"YmluCmJvb3QKZGVhZC5sZXR0ZXIKZGV2CmV0Ywpob21lCmxpYgpsaWI2NApsb3N0K2ZvdW5kCm1lZGlhCm1udApvcHQKcHJvYwpyb290CnJ1bgpzYmluCnNlbGludXgKc3J2CnN5cwp0bXAKdXNyCnZhcgo=","exited":true}}

base64 decode the out-data

> command output decode


*/
