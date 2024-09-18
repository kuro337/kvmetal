#!/bin/bash

HOST_IP_SUBNET=
LIBVIRT_SUBNET=

echo -e "If not set - make sure qemu hook is set to map Host Public Subnet with virbr0 bridge"

# v=$(/sbin/iptables -L FORWARD -n -v | /usr/bin/grep 192.168.122.0/24 | /usr/bin/wc -l)
# # avoid duplicate as this hook get called for each VM
# [ $v -le 2 ] && /sbin/iptables -I FORWARD 1 -o virbr0 -m state -s 192.168.2.0/24 -d 192.168.122.0/24 --state NEW,RELATED,ESTABLISHED -j ACCEPT


echo -e "Parse output of this to log VM Private IPs"

virsh net-dhcp-leases default



echo -e "For each VM - we can specify a mapping for HostIP:Port::vmIP:Port"

#
# *nat
# :PREROUTING ACCEPT [0:0]
# -A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:9999 -m comment --comment "Testing port 9999 of vm from ubuntu host 9999"
# COMMIT

echo -e "Enable using ufw and reload Qemu Hooks and firewall"
# sudo ufw reload




# below are only required first time - for changes simply run sudo ufw reload
# sudo ufw enable
# ufw reload
# bash /etc/libvirt/hooks/qemu

echo -e "In case of connectivity issues and if duplicate rules show up - lower number rules take precedence"

echo -e "Identify rules and delete them accordingly"
# sudo iptables -t nat -L PREROUTING -n -v --line-number
# sudo iptables -t nat -D PREROUTING 4
# sudo iptables -t nat -D PREROUTING 3

#### CLEANUP

echo -e "Rules need to be deleted manually using sudo iptables -t nat -D PREROUTING $RULE_NUMBER"

# sudo iptables -t nat -D PREROUTING $RULE_NUMBER
# sudo ufw reload

