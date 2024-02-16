# AUTOMATE EXPOSING VM

```bash
Complete

network/ExposeVM

# curr we get IP
# now figure out qemu hooks aspect

# we have the Logic to accept a VM name and generate the Rule

# now need to figure out how to update/keep qemu hooks file in Sync - this is crucial


# File 1 : /etc/libvirt/hooks/qemu

We need to define here the Host IP & the NAT libvirt IP+Subnet - when we want Port Forwarding ON for VMs

# virsh net-dumpxml default run this and find the line with bridge name=virbr0 and then get the IP from next block start

 <bridge name='virbr0' stp='on' delay='0'/>
 <mac address='52:54:00:ac:91:9f'/>
 <ip address='192.168.122.1' netmask='255.255.255.0'> # our IP is 192.168.122.1 - 1 => 192.168.122.0

# we already know how to get our Host IP with Subnet
Host IP:192.168.1.194/24
# hostIP, _ := GetHostIP(false)
# log.Printf("Host IP:%s", hostIP)


# from above we construct this

v=$(/sbin/iptables -L FORWARD -n -v | /usr/bin/grep 192.168.122.0/24 | /usr/bin/wc -l)
# avoid duplicate as this hook get called for each VM
[ $v -le 2 ] && /sbin/iptables -I FORWARD 1 -o virbr0 -m state -s 192.168.1.0/24 -d 192.168.122.0/24 --state NEW,RELATED,ESTABLISHED -j ACCEPT

# so create

func CreateQemuHooksFile() string {

}


##### next

sudo cat /etc/ufw/before.rules
We already have

# CreateUfwBeforeRule(vmIP, vmPort, hostPort, "Rule to expose Yarn UI")
# gives us the rule
# -A PREROUTING -p tcp --dport 5555 -j DNAT --to-destination 192.168.122.135:8088 -m comment --comment "Rule to expose Yarn UI"

TASK: Create a function - AddUfwBeforeRule() - that reads the content - and for now - if its commented simply adds our String as a Comment (in memory)

If it its not commented - adds our string normally ABOVE KVM_GO_END and below the line above it - and prints the result



```
