package constants

type InitSvc string

const (
	Restart InitSvc = "Restart"
)

/*
To Validate a User Data Schema for CloudInit

sudo cloud-init schema --config-file currRun.yaml  --annotate

# Run this on login to set default shell

chsh -s $(which zsh)
sudo chsh -s $(which zsh)
*/
const CloudInitUbuntu = `#cloud-config

#hostname: _HOSTNAME_
#fqdn: _FQDN_
passwd: password  
lock_passwd: false
sudo: ['ALL=(ALL) NOPASSWD:ALL']
# sudo: ALL=(ALL) NOPASSWD:ALL
package-update: true
package_upgrade: true
password: password
ssh_pwauth: true
chpasswd: { expire: False }

`

const DefaultUserdata = `#cloud-config

#hostname: _HOSTNAME_
#fqdn: _FQDN_
lock_passwd: false
sudo: ['ALL=(ALL) NOPASSWD:ALL']
package-update: true
package_upgrade: true
password: password
ssh_pwauth: true
chpasswd: { expire: False }
#ssh_authorized_keys:
#  - ssh-rsa $SSH_PUB

`

const RebootCloudInit = `
power_state:
  mode: reboot
  message: Rebooting after cloud-init configuration
  timeout: 15 
  condition: True`
