# Alpine Cloud Image

Reference: https://gitlab.alpinelinux.org/alpine/cloud/alpine-cloud-images/-/blob/main/configs/cloud/aws.conf

## validation

```bash
sudo cloud-init schema --config-file user-data --annotate
sudo cloud-init schema --config-file ./network-config --schema-type network-config --annotate
```

```bash
curl -O https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/cloud/nocloud_alpine-3.20.3-x86_64-bios-cloudinit-r0.qcow2

sudo apt update
sudo apt install genisoimage

genisoimage -output cloud-init.iso -volid cidata -joliet -rock meta-data user-data

sudo cloud-init schema --config-file user-data --annotate

virt-install --connect="qemu:///system" \
--name alpine-vm \
--ram 2048 \
--vcpus 2 \
--disk path=/var/lib/libvirt/images/nocloud_alpine-3.20.3-x86_64-bios-cloudinit-r0.qcow2,format=qcow2,size=10 \
--disk path=/var/lib/libvirt/images/cloud-init.iso,device=cdrom \
--os-variant alpinelinux3.19 \
--network bridge=virbr0 \
--import --graphics none 

echo "instance-id: alpine-vm" > meta-data
echo "local-hostname: alpine-vm" >> meta-data

```

## Cloud init reference

https://cloudinit.readthedocs.io/en/24.1/reference/examples.html

```bash


sudo rm cloud-init.iso
genisoimage -output cloud-init.iso -volid cidata -joliet -rock meta-data user-data network-config
sudo rm /var/lib/libvirt/images/alpine-vm-disk.qcow2
sudo rm /var/lib/libvirt/images/cloud-init.iso
sudo cp cloud-init.iso /var/lib/libvirt/images/cloud-init.iso


echo "instance-id: alpine-vm" > meta-data
echo "local-hostname: alpine-vm" > meta-data

cat > user-data <<EOF
#cloud-config
hostname: alpine-vm
package_update: true
package_upgrade: true
ssh_pwauth: true # sshd service will be configured to accept password authentication method
password: password

ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDte6rr8ubfGx9dQoH9hHrtKVYvRdf7Fzw9IXdZtI6VbRj4Bc8h3X/Wz9aQGHsbLxC3CGCOVTzoaQzoX+pNmd7oDcMHgjsiFlYce18QnDVorwjgeU1abtQNQdc97S0fS95tJDxXwRTJrX7hWQCAkrnS7E2wj+pWPuXc50gEXlkubSsCuN4C9P+Gd+ul4Sq/urz+FRnikTJXAioaKWge/ZQDp7HHCwNw497KE4VkntunExgg/ElJCGWTCSwlIpq007iRpxMngclPPTLqny2BvuHuzRmhMRZYZNcOkVe5G8fre0beDt3giwUtuVNAQqYSXmmBGQRtda4VwC7oQbe5uus9H/idrmTC7gKVdH5J1FMNMIXREF9TMiZTiAZeou5l+QMwM44r2SsvjALkB2z5NunC9sL8EycUZWu1G95xdhTl3iK7oF8gvsdhEjOYkwePzqWoqJHuUOBkis3jnJfKnPKtMECadBcE38ewSihI7/76KTt+nE9ecwceqMd+YP5jUqW6ufaJr7K7Wvb3M7i7IUDjDkSLqRuHX16sG/Rt2xVR2jNmcFf5jhU/DexlPh0tg1b4dEUx1hw4RwDVb/EU9iWwQMSLf9VCKBBQ3KNcrCI1HPAOqP39XjoYUHKCBKFYWiJpzN4i4kOgEPc4RRlObcpV/+RX8RAb/TtXJzJCXknarw== kuro@kuro-dev

users:
  - default 

packages:
  - zsh
  - openssh
runcmd:
  - apk update
  - apk add zsh
  - doas apk add sudo
EOF


cat > network-config <<EOF
version: 2
ethernets:
  eth0:
    addresses:
      - 192.168.122.150/24
    routes:
      - to: 0.0.0.0/0
        via: 192.168.122.1
    nameservers:
      addresses:
        - 8.8.8.8
        - 8.8.4.4
    dhcp4: false
    dhcp6: false
EOF

genisoimage -output cloud-init.iso -volid cidata -joliet -rock meta-data user-data

```
