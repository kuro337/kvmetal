# ufw and exposing VMs

ufw : uncomplicated firewall

goated article - https://www.cyberciti.biz/faq/kvm-forward-ports-to-guests-vm-with-ufw-on-linux/

```bash
# comes installed with ubuntu so install might not be required
sudo ufw status
sudo apt-get update
sudo apt-get install ufw

# from here - get the virbr0 NAT subnet :
virsh net-list
virsh net-info default
virsh net-dumpxml default

# e.g
  <bridge name='virbr0' stp='on' delay='0'/>
  <mac address='52:54:00:ac:91:9f'/>
  <ip address='192.168.122.1' netmask='255.255.255.0'>

# our NAT subnet will be 192.168.122.0 - (even if its 1 - we take the full subnet)
# then get our host IP+subnet by running docs/network/hostip.sh

# host ip subnet    : 192.168.1.194/24

# virbr0 nat subnet : 192.168.122.0/24

# step 1.
cd /etc/libvirt/hooks/
vi qemu

# get ip and add below bash script and then run chmod for the hook (replace with host ip+sub)
# 192.168.1.194/24

sudo chmod -v +x /etc/libvirt/hooks/qemu

```

script to add before chmod

```bash
#!/bin/bash
# Hook to insert NEW rule to allow connection for VMs
# 192.168.122.0/24 is NATed subnet
# virbr0 is networking interface for VM and host
# -----------------------------------------------------------------
# Written by Vivek Gite under GPL v3.x {https://www.cyberciti.biz}
# -----------------------------------------------------------------
# get count
#################################################################
## NOTE replace 192.168.2.0/24 with your public IPv4 sub/net   ##
#################################################################
v=$(/sbin/iptables -L FORWARD -n -v | /usr/bin/grep 192.168.122.0/24 | /usr/bin/wc -l)
# avoid duplicate as this hook get called for each VM
[ $v -le 2 ] && /sbin/iptables -I FORWARD 1 -o virbr0 -m state -s 192.168.2.0/24 -d 192.168.122.0/24 --state NEW,RELATED,ESTABLISHED -j ACCEPT

```

our script after getting

- host ip subnet : 192.168.1.194/24
- virbr0 nat subnet : 192.168.122.0/24

```bash
cd /etc/libvirt/hooks/
vi qemu
chmod -v +x /etc/libvirt/hooks/qemu

```

```bash
#!/bin/bash

v=$(/sbin/iptables -L FORWARD -n -v | /usr/bin/grep 192.168.122.0/24 | /usr/bin/wc -l)
# avoid duplicate as this hook get called for each VM
[ $v -le 2 ] && /sbin/iptables -I FORWARD 1 -o virbr0 -m state -s 192.168.1.0/24 -d 192.168.122.0/24 --state NEW,RELATED,ESTABLISHED -j ACCEPT

```

### port forwarding

```bash
# get the internal ips of vms
virsh net-dhcp-leases default

# we got this ip and subnet 192.168.122.109/24

```

```bash
sudo vi /etc/ufw/before.rules

# update file so that from our laptop -
# we can reach the VM through host at port 9999

# add below to the top of the file

# note that - we dont need to specify the host IP because we only have 1 IP

*nat
:PREROUTING ACCEPT [0:0]
-A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:9999 -m comment --comment "Testing port 9999 of vm from ubuntu host 9999"
COMMIT


##### after adding it to the top and saving
sudo ufw enable

sudo bash /etc/libvirt/hooks/qemu
sudo ufw reload

# or reboot the full server
reboot

# list rules now
sudo iptables -L FORWARD -nv --line-number
sudo iptables -t nat -L PREROUTING -n -v --line-number
sudo iptables -t nat -L -n -v

# NOTE IMPORTANT: 
# Running ufw enable or reload too many times can add Duplicate Rules

# identify dupe rules and delete them
sudo iptables -t nat -L PREROUTING -n -v --line-number

# for example if all these rules wrong delete them
sudo iptables -t nat -D PREROUTING 4
sudo iptables -t nat -D PREROUTING 3
sudo iptables -t nat -D PREROUTING 2

# after confirming /etc/ufw/before.rules looks good - run
sudo ufw reload


# run this on vm
nc -lk 9999

# from laptop run
nc 192.168.1.194 9999

```

```bash
# KVM/libvirt Forward Ports to guests with Iptables (UFW) #
*nat
:PREROUTING ACCEPT [0:0]
-A PREROUTING -d 202.54.1.4 -p tcp --dport 1:65535 -j DNAT --to-destination 192.168.122.253:1-65535 -m comment --comment "VM1/CentOS 7 ALL ports forwarding"
-A PREROUTING -d 202.54.1.5 -p tcp --dport 22 -j DNAT --to-destination 192.168.122.125:22 -m comment --comment "VM2/OpenBSD SSH port forwarding"
-A PREROUTING -d 202.54.1.6 -p tcp --dport 443 -j DNAT --to-destination 192.168.122.231:443 -m comment --comment "VM3/FreeBSD 443 port forwarding"
-A PREROUTING -d 202.54.1.7 -p tcp --dport 80 -j DNAT --to-destination 192.168.122.229:80 -m comment --comment "VM4/CentOS 80 port forwarding"
COMMIT
# run this on vm
nc -lk 9999

# from laptop run
nc 192.168.1.194 9999

sudo iptables -t nat -L -n -v

virsh reboot hadoop

# in case we ever lose access to a VM session and need to kill the console
ps aux | grep 'virsh console hadoop'
kill -9 pid
```
