#!/bin/bash


RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' 

# Set exit on error
set -e

# Update & Upgrade Packages
echo -e "${BOLD}*** Updating and Upgrading Packages... ***${NC}"
sudo apt-get update && sudo apt-get upgrade -y


# Install Transport HTTPS
echo -e "${GREEN}*** Installing Transport HTTPS... ***${NC}"
sudo apt-get install -y apt-transport-https -y

# Disable Swap
echo -e "${BLUE}*** Disabling Swap... ***${NC}"
sudo swapoff -a



# Install and Configure containerd
echo -e "${BOLD}*** Installing containerd... ***${NC}"
sudo apt-get install -y containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml


echo -e "${GREEN}*** Configuring systemd cgroup driver... ***${NC}"
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd


# Load Modules
echo -e "${BLUE}*** Loading Kernel Modules... ***${NC}"
sudo modprobe overlay
sudo modprobe br_netfilter

# Set Sysctl
echo -e "${BOLD}*** Setting Sysctl... ***${NC}"
cat <<EOF | sudo tee /etc/sysctl.d/99-kubernetes-cri.conf
net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
sudo sysctl --system


# Add Kubernetes Repo
echo -e "${GREEN}*** Adding Kubernetes Repository... ***${NC}"
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/kubernetes.gpg
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list


# Install Kubernetes Components
echo -e "${BLUE}*** Installing Kubernetes Components... ***${NC}"
sudo apt-get update -y
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

echo -e "${GREEN}Worker Node Setup Complete! Node is ready to be joined to the cluster.${NC}"
echo -e "${RED}To join the cluster, use the join command provided by the control-plane node.${NC}"


echo -e "${BLUE}*** Joining worker node to Kubernetes Cluster initiated by kubecontrol ***${NC}"
