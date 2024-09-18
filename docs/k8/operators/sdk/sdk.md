# Operator SDK


```bash
# Ensure GOPROXY is set to "https://proxy.golang.org|direct"
echo 'export GOPROXY="https://proxy.golang.org|direct"' >> ~/.profile

source ~/.profile

# for make
sudo apt update && sudo apt install build-essential

git clone https://github.com/operator-framework/operator-sdk
cd operator-sdk
git checkout master
make install



```
