#!/bin/bash

# Check if KVM modules are loaded
lsmod | grep kvm

# Check libvirtd service status
systemctl status libvirtd

# Check QEMU version
qemu-system-x86_64 --version

# Check libvirt version
virsh --version

sudo apt update
sudo apt install qemu-kvm libvirt-daemon-system
