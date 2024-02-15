#!/bin/bash

# Get the default network interface
default_interface=$(ip route | grep default | awk '{print $5}' | head -n 1)

# Get the IP address of the default network interface
host_ip=$(ip addr show $default_interface | grep 'inet ' | awk '{print $2}' | cut -d/ -f1)

echo -e "Host IP:\n$host_ip"
