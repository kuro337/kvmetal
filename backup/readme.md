# kernel updates

Cloud-init for setting latest kernel and security upgrades on boot

```bash
#cloud-config
password: password
chpasswd:
  expire: False
ssh_pwauth: True
package_update: true
package_upgrade: true

```
