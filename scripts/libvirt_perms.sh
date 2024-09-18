#!/bin/bash

# Run this with sudo
KVMGO_DIR="/home/kuro/Documents/Code/Go/kvmgo"

# Grant read and execute permissions to libvirt-qemu for the entire directory
sudo chmod -R 755 "$KVMGO_DIR"
sudo setfacl -R -m u:libvirt-qemu:rX "$KVMGO_DIR"

echo "Permissions updated for libvirt-qemu on $KVMGO_DIR"

# sudo vim /etc/libvirt/qemu.conf
# Uncomment user='root' and group='root'

sudo virt-install \
  --name ubuntu-cloud-vm \
  --memory 2048 \
  --vcpus 2 \
  --disk path=/home/kuro/Documents/Code/Go/kvmgo/data/images/ubuntu-22.04-server-cloudimg-amd64.img,format=qcow2 \
  --import \
  --os-variant ubuntu22.04 \
  --network bridge=virbr0 \
  --noautoconsole


sudo virt-install \
  --name ubuntu-vm \
  --memory 2048 \
  --vcpus 2 \
  --disk path=/path/to/your/image.qcow2,format=qcow2 \
  --cdrom /path/to/ubuntu-20.04.4-live-server-amd64.iso \
  --network bridge=virbr0 \
  --graphics vnc,listen=0.0.0.0 \
  --noautoconsole \
  --os-type linux \
  --os-variant ubuntu20.04
