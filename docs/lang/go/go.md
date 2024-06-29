# Install Go

```bash
sudo apt install golang-go

# Pull Source Files - Check https://go.dev/dl/ for latest releases 
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz -O go.tar.gz
sudo tar -xzvf go.tar.gz -C /usr/local

# echo export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH >> ~/.profile

echo 'export GOBIN=$HOME/go/bin' >> ~/.profile
echo 'export PATH=$PATH:$GOBIN' >> ~/.profile

source ~/.profile



```
