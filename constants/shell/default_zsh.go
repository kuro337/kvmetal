package shell

const ZSH_UBUNTU_RUNCMD_NOINDENT = `
sudo -u ubuntu sh -c 'RUNZSH=no sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"'
sudo -u ubuntu git clone https://github.com/zsh-users/zsh-autosuggestions /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions
sudo -u ubuntu git clone https://github.com/zsh-users/zsh-syntax-highlighting /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting
echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
echo 'plugins=(git zsh-autosuggestions zsh-syntax-highlighting)' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
sudo sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
sudo systemctl restart sshd
`

const ZSH_UBUNTU_RUNCMD = ` 
  # Clone zsh-autosuggestions
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-/home/ubuntu/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
  # Clone zsh-syntax-highlighting
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-syntax-highlighting ${ZSH_CUSTOM:-/home/ubuntu/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting
  # Append plugins to .zshrc
  - echo "plugins=(git zsh-autosuggestions zsh-syntax-highlighting)" | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  # Source the files or set any other configurations you need
  - echo "source ${ZSH_CUSTOM:-/home/ubuntu/.oh-my-zsh/custom}/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh" | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo "source ${ZSH_CUSTOM:-/home/ubuntu/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  # comment out below in case causing issues
  - sudo chsh -s $(which zsh)
`

const Userdata_Literal_zsh_kernelupgrade = `#cloud-config

users:
  - name: ubuntu
    #hostname: _HOSTNAME_
    shell: /usr/bin/zsh
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    groups: sudo
    passwd: password 
    lock_passwd: false

package_upgrade: true

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

ssh_pwauth: true
chpasswd:
  list: |
     ubuntu:password
  expire: False
`
