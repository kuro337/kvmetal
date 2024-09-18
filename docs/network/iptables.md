# iptables

Backing up iptables

```bash
# creating backups in curr dir

sudo iptables-save > iptables_backup_ipv4.txt
sudo ip6tables-save > iptables_backup_ipv6.txt




# restoring
sudo iptables-restore < iptables_backup_ipv4.txt
sudo ip6tables-restore < iptables_backup_ipv6.txt


```
