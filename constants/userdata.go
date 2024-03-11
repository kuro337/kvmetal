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

const DefaultUserDataShellZsh = `#cloud-config

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

packages:
  - zsh

runcmd:
  - sudo -u ubuntu sh -c 'RUNZSH=no sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"'
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-autosuggestions /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-syntax-highlighting /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting
  - echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo 'plugins=(git zsh-autosuggestions zsh-syntax-highlighting)' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - sudo sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
  - sudo systemctl restart sshd

`

const RebootCloudInit = `
power_state:
  mode: reboot
  message: Rebooting after cloud-init configuration
  timeout: 15 
  condition: True`
