package kube

const KUBE_CONTROL_UBUNTU_RUNCMD = `
  # Disable Swap
  - swapoff -a

  # Install and Configure containerd
  - mkdir -p /etc/containerd
  - containerd config default | tee /etc/containerd/config.toml
  - sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
  - systemctl restart containerd
  - systemctl enable containerd

  # Load Modules
  - modprobe overlay
  - modprobe br_netfilter

  # Set Sysctl
  - |
    echo "net.bridge.bridge-nf-call-iptables  = 1
    net.ipv4.ip_forward                 = 1
    net.bridge.bridge-nf-call-ip6tables = 1" | tee /etc/sysctl.d/99-kubernetes-cri.conf
  - sysctl --system

  # Add Kubernetes Repo
  - curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/kubernetes-archive-keyring.gpg
  - echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | tee /etc/apt/sources.list.d/kubernetes.list

  # Install Kubernetes Components
  - apt-get update && apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl

  # Initialize Kubernetes
  - kubeadm init | tee /home/ubuntu/kubeadm-init.log

  # Setup Kubeconfig
  - mkdir -p /home/ubuntu/.kube
  - cp -i /etc/kubernetes/admin.conf /home/ubuntu/.kube/config
  - chown $(id -u ubuntu):$(id -g ubuntu) /home/ubuntu/.kube/config

  # Setup pod networking (Calico)
  - kubectl --kubeconfig=/home/ubuntu/.kube/config apply -f https://docs.projectcalico.org/manifests/calico.yaml

  # Install Helm
  - curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

`

const KUBE_WORKER_UBUNTU_RUNCMD = `
  # Disable Swap
  - swapoff -a
  # Load Modules
  - modprobe overlay
  - modprobe br_netfilter

  # Enable Containerd
  - mkdir -p /etc/containerd
  - containerd config default | tee /etc/containerd/config.toml
  - sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
  - systemctl restart containerd
  - systemctl enable containerd

  # Set Sysctl
  - |
    echo "net.bridge.bridge-nf-call-iptables  = 1" >> /etc/sysctl.d/k8s.conf
    echo "net.ipv4.ip_forward                 = 1" >> /etc/sysctl.d/k8s.conf
    echo "net.bridge.bridge-nf-call-ip6tables = 1" >> /etc/sysctl.d/k8s.conf
    sysctl --system
  # Add Kubernetes Repo
  - curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /etc/apt/trusted.gpg.d/kubernetes.gpg
  - echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
  # Update & Install Kubernetes Components
  - apt-get update
  - apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl



`
