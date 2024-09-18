#!/bin/bash


# Update and upgrade packages non-interactively
sudo DEBIAN_FRONTEND=noninteractive apt-get update && sudo DEBIAN_FRONTEND=noninteractive apt-get -y upgrade

# Configure needrestart to automatically decide on service restarts
echo "\$nrconf{restart} = 'a';" | sudo tee -a /etc/needrestart/needrestart.conf > /dev/null


sudo DEBIAN_FRONTEND=noninteractive apt-get install zsh -y

# Change the default shell to Zsh for the current user without a password prompt
sudo chsh -s $(which zsh) $USER


# Install Oh My Zsh without automatically changing the shell again
RUNZSH=no sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"


# Wait for Oh My Zsh installation to complete
sleep 5

# Install additional Zsh plugins
ZSH_CUSTOM="/home/ubuntu/.oh-my-zsh/custom"
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM}/plugins/zsh-autosuggestions
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM}/plugins/zsh-syntax-highlighting

# Directly modify .zshrc to include desired plugins without sourcing it
echo "source ${ZSH_CUSTOM}/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh" >> /home/ubuntu/.zshrc
echo "source ${ZSH_CUSTOM}/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" >> /home/ubuntu/.zshrc
echo "plugins=(git zsh-autosuggestions zsh-syntax-highlighting)" >> /home/ubuntu/.zshrc

