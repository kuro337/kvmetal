# ensuring host is resolvable using hostname

## F F F

Current Setup:

```bash
# On VM -
hostname -> kafka
hostname -f -> kafka.kvm (fqdn set to kafka)

# 1. Add this to default virnet (sudo virsh net-dumpxml default)

 <domain name='kvm' localOnly='yes'/>

# 2. /etc/nsswitch.conf
hosts:          files mdns4_minimal [NOTFOUND=return] libvirt libvirt_guest dns

# 3. /etc/systemd/resolved.conf
[Resolve]
DNS=192.168.122.1
Domains=~kvm
ResolveUnicastSingleLabel=yes

sudo virsh net-dumpxml default
sudo virsh net-edit default
virsh net-destroy default
virsh net-start default
sudo systemctl restart libvirtd
sudo systemctl restart NetworkManager
sudo systemctl restart systemd-resolved

```

## External

```bash
# on the Host - add IP of VM
/etc/hosts

192.168.122.121 kafka.kuro.com


```

## libvirt-nss

```bash
https://libvirt.org/nss.html

sudo apt install libnss-libvirt

# or

https://liquidat.wordpress.com/2017/03/03/howto-automated-dns-resolution-for-kvmlibvirt-guests-with-a-local-domain/

https://m0dlx.com/blog/Automatic_DNS_updates_from_libvirt_guests.html

# Possible Relevant Issue of hostname bug
https://github.com/canonical/cloud-init/issues/3088

sudo virt-customize -a https://cloud-images.ubuntu.com/focal/current/focal-server-cloudimg-amd64.img --truncate /etc/machine-id

https://unix.stackexchange.com/questions/482348/how-single-label-dns-lookup-requests-are-handled-by-systemd-resolved


https://github.com/systemd/systemd/issues/18761


```

Update Libvirt bridge using `virsh net-edit default`

```bash

192.168.122.1 kafka.kuro.co

192.168.122.78

<domain name='karhu.xyz' localOnly='yes'/>


sudo virsh net-dumpxml default

# we want to add a Domain - to resolve the VM and Hosts (only locally)

 <domain name='kuro.com' localOnly='yes'/>

sudo virsh net-edit default

# After changing:

virsh net-destroy default
virsh net-start default


```

Their Config:

```xml

<network connections='1'>
  <name>default</name>
  <uuid>158880c3-9adb-4a44-ab51-d0bc1c18cddc</uuid>
  <forward mode='nat'>
    <nat>
      <port start='1024' end='65535'/>
    </nat>
  </forward>
  <bridge name='virbr0' stp='on' delay='0'/>
  <mac address='52:54:00:fa:cb:e5'/>
  <domain name='qxyz.de' localOnly='yes'/>
  <ip address='192.168.122.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.122.128' end='192.168.122.254'/>
    </dhcp>
  </ip>
</network>

```

Current:

```xml

<network>
  <name>default</name>
  <uuid>01fa705e-c047-4697-b2f7-76034c90bc74</uuid>
  <forward mode='nat'>
    <nat>
      <port start='1024' end='65535'/>
    </nat>
  </forward>
  <bridge name='virbr0' stp='on' delay='0'/>
  <mac address='52:54:00:ac:91:9f'/>
  ## add this <domain name='kuro.com' localOnly='yes'/>
  <ip address='192.168.122.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.122.2' end='192.168.122.254'/>
      ## remove this <host mac='52:54:00:C0:D4:5C' name='talos-worker' ip='192.168.122.100'/>
    </dhcp>
  </ip>
</network>

```

## Configuring the VM guests

When the domain is set, the guests inside the VMs need to be defined.

With recent Linux releases this is as simple as setting the host name:

```bash
sudo hostnamectl set-hostname neon.qxyz.de
sudo hostnamectl set-hostname kafka.kuro.com

192.168.122.78

```

## 3. Configuring NetworkManager

```bash
# libvirt ships with its own dnsmasq
# if not present, install it on our machine - no interference and safe to run if exists
sudo apt-get install dnsmasq-base

# | sudo tee pipes to stdout and writes to a file and
#   circumvents role issues related to piping sudo and permission changes

echo -e "[main]\ndns=dnsmasq" | sudo tee /etc/NetworkManager/conf.d/localdns.conf
echo "server=/kuro.com/192.168.122.1" | sudo tee /etc/NetworkManager/dnsmasq.d/libvirt_dnsmasq.conf

# 1. Add dnsmasq to NetworkManager
cat /etc/NetworkManager/conf.d/localdns.conf
[main]
dns=dnsmasq

# 2. Set this to be the Domain and IP from output of sudo virsh net-dumpxml default
cat /etc/NetworkManager/dnsmasq.d/libvirt_dnsmasq.conf
server=/kuro.com/192.168.122.1

hosts:          files mdns4_minimal [NOTFOUND=return] libvirt libvirt_guest dns

# 3. Add libvirt to /etc/nsswitch.conf

hosts:          files mdns4_minimal [NOTFOUND=return] libvirt libvirt_guest dns

# IN CASE WE NEED TO RESET THIS WAS ORIGINAL FOR /etc/nsswitch.conf
hosts:          files mdns4_minimal [NOTFOUND=return] dns mymachines

# done - restart network manager

sudo systemctl restart NetworkManager

```

```bash
# should be kafka
hostname

# we want this to be kafka.kuro.com
hostname -f

cat /etc/hosts

# Add to /etc/hosts
127.0.0.1 kafka.kuro.com kafka

# confirm resolution
getent hosts kafka.kuro.com


nslookup kafka.kuro.com

systemctl status systemd-resolved


hostname -f

sudo virsh net-list --all

resolvectl status

systemctl list-units --type=service | grep -E 'dnsmasq|systemd-resolved'

ls -l /etc/resolv.conf # should symlink to systemd/resolve/stub-resolv.conf

/etc/NetworkManager/NetworkManager.conf must have dns=dnsmasq

[main]
plugins=ifupdown,keyfile
dns=dnsmasq
....

sudo systemctl restart NetworkManager

# Verify its being used
pgrep -a dnsmasq

sudo vi /etc/systemd/resolved.conf

/etc/systemd/resolved.conf

[Resolve]
DNS=192.168.122.1
Domains=~kuro.com

sudo systemctl restart systemd-resolved

DNS=192.168.122.1
Domains=~kvm

sudo systemctl restart libvirtd
sudo systemctl restart NetworkManager
sudo systemctl restart systemd-resolved

sudo vi /etc/systemd/resolved.conf.d/libvirt.conf

```
