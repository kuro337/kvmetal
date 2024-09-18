# vm hostname

```bash
# NOTE: the solution is to run a reboot on the VM - to fix it
# it is a bug with cloud-init: we can use

sudo virsh reboot $VM


#### below is optional - above sudo virsh reboot vm should work
sudo virsh net-dhcp-leases default
# make sure hostname and fqdn are set in userdata

hostname: hadoop
fqdn: hadoop.kuro.com

# from the VM run
sudo hostnamectl set-hostname hadoop

# confirm
hostname
hostnamectl

# if above doesnt work do this after - and check again

sudo dhclient -r  # Release the current lease
sudo dhclient     # explicitly acquire a new lease (above should be sufficient)

sudo virsh net-dhcp-leases default



# Should show hostname

# for persistence - across restarts
sudo echo "hadoop" | sudo tee /etc/hostname
sudo systemctl restart systemd-hostnamed

```
