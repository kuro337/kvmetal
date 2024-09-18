# Containerd

## Prerequisites

```bash
sudo apt-get update
sudo apt-get install make
sudo apt-get install unzip

# install Go
wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
source ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc

# containerd
wget https://github.com/containerd/containerd/releases/download/v1.7.22/containerd-1.7.22-linux-amd64.tar.gz
sudo tar Cxzvf /usr/local containerd-1.7.22-linux-amd64.tar.gz

# get systemd agent
wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service

sudo mkdir -p /usr/local/lib/systemd/system/
sudo mv containerd.service /usr/local/lib/systemd/system/

# /usr/local/lib/systemd/system/containerd.service
sudo systemctl daemon-reload
sudo systemctl enable --now containerd

# install runc
wget https://github.com/opencontainers/runc/releases/download/v1.1.14/runc.amd64
sudo install -m 755 runc.amd64 /usr/local/sbin/runc
# ensure is enabled seccomp
runc --version
# runc version 1.2.0-rc.3
# commit: v1.2.0-rc.3-0-g45471bc9
# spec: 1.2.0
# go: go1.22.6
  libseccomp: 2.5.5

# install cni plugins
wget https://github.com/containernetworking/plugins/releases/download/v1.5.1/cni-plugins-linux-amd64-v1.5.1.tgz
sudo mkdir -p /opt/cni/bin
sudo tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.5.1.tgz

sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml

# update config
update SystemdCgroup = true 
sudo vi /etc/containerd/config.toml

# done

sudo apt-get update

```

# install kubeadm

```bash
# disables swap until next reboot (extending memory using disk)
swapoff -a

sudo apt-get update
# apt-transport-https may be a dummy package; if so, you can skip that package
sudo apt-get install -y apt-transport-https ca-certificates curl gpg
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

# enable kubelet before starting kubeadm
sudo systemctl enable --now kubelet

# kubeadm init $args

git remote add origin github.com:kuro337/kvmetal.git

git remote github.com:kuro337/kvmetal.git
echo -e "net.bridge.bridge-nf-call-iptables = 1\nnet.bridge.bridge-nf-call-ip6tables = 1\nnet.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-kubernetes-cri.conf
sudo sysctl --system
sudo apt update
sudo apt install -y socat

# Note: Very important to save the output
# CIDR commonly used with flannel
sudo kubeadm init --pod-network-cidr=10.244.0.0/16 | sudo tee /home/ubuntu/kubeadm-init.log

# Configure kubectl
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl get nodes

# optional flannel
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
kubectl cluster-info
# enable storage drivers
modprobe overlay
modprobe br_netfilter

# Add efficient storage drivers to persistent loading
sudo tee /etc/modules-load.d/kubernetes.conf > /dev/null <<EOF
overlay
br_netfilter
EOF
sudo systemctl restart systemd-modules-load.service


sudo kubeadm init 

kubeadm init --skip-phases=addon/kube-proxy | tee /home/ubuntu/kubeadm-init.log

```

## source build tbd

```bash
# below is from source, skip 
# install protobuf
wget https://github.com/protocolbuffers/protobuf/releases/download/v28.1/protoc-28.1-linux-x86_64.zip
sudo unzip protoc-28.1-linux-x86_64.zip -d /usr/local
which protoc && protoc --version

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz

wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
sudo mkdir -p /usr/local/lib/systemd/system/containerd.service
cp containerd.service /usr/local/lib/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable --now containerd


cd containerd


```

# Building containerd

```bash
git clone https://github.com/containerd/containerd
cd containerd
make
sudo make install

# creating config
sudo mkdir -p /etc/containerd
sudo containerd config default | sudo tee /etc/containerd/config.toml

sudo vi /etc/containerd/config.toml

plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options
```
