# upgrades

```bash
###### Step 0. Setup and Upgrade System Packages
echo -e "Updating Packages"

# run if as a user
sudo apt update && sudo apt upgrade -y

# we can run this in scripts to fully automate it and be nonblocking


sudo DEBIAN_FRONTEND=noninteractive apt-get update && sudo DEBIAN_FRONTEND=noninteractive apt-get -y upgrade
echo "\$nrconf{restart} = 'a';" | sudo tee -a /etc/needrestart/needrestart.conf > /dev/null


export EDITOR='/home/kuro/.vscode-server/cli/servers/Stable-863d2581ecda6849923a2118d93a088b0745d9d6/server/bin/remote-cli/code --wait'

# Finding the correct dir
fd 'vscode' $HOME -d 1 -H

# Look in $HOME for .vscode-server
/home/kuro/.vscode-server/cli/servers/

# Find bin folder
/home/kuro/.vscode-server/cli/servers/Stable-863d2581ecda6849923a2118d93a088b0745d9d6/server/bin/remote-cli

code



/home/kuro/.vscode-server/cli/servers/Stable-863d2581ecda6849923a2118d93a088b07
45d9d6/server/bin/remote-cli/code packages.md


/home/kuro/.vscode-server/cli/servers/Stable-863d2581ecda6849923a2118d93a088b07
45d9d6/server/bin/remote-cli/code

```
