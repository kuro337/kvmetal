package constants

/*
To Validate a User Data Schema for CloudInit

sudo cloud-init schema --config-file currRun.yaml  --annotate

# Run this on login to set default shell

chsh -s $(which zsh)
sudo chsh -s $(which zsh)
*/
const CloudInitUbuntu = `#cloud-config

#hostname: _HOSTNAME_
passwd: password  
lock_passwd: false
sudo: ALL=(ALL) NOPASSWD:ALL
package-update: true
package_upgrade: true
password: password
ssh_pwauth: true
chpasswd: { expire: False }



chpasswd: { expire: False }`
