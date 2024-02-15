# upgrades

```bash
###### Step 0. Setup and Upgrade System Packages
echo -e "Updating Packages"

# run if as a user
sudo apt update && sudo apt upgrade -y

# we can run this in scripts to fully automate it and be nonblocking

sudo DEBIAN_FRONTEND=noninteractive apt-get update && sudo DEBIAN_FRONTEND=noninteractive apt-get -y upgrade
echo "\$nrconf{restart} = 'a';" | sudo tee -a /etc/needrestart/needrestart.conf > /dev/null



```
