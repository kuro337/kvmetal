package kube

const KUBE_CONTROL_CALICO_UBUNTU_RUNCMD = `
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
  - apt-get install -y apt-transport-https ca-certificates curl gpg
  - mkdir -p /etc/apt/keyrings/ && curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  - echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list

  # - curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /etc/apt/trusted.gpg.d/kubernetes.gpg
  # - echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
  # Update & Install Kubernetes Components
  - apt-get update
  - apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl



`

const KUBE_CONTROL_CILIUM_UBUNTU_RUNCMD = `
  # Disable Swap - if connection fails to Kube this might be getting reactivated - call it again later
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
  - apt-get install -y apt-transport-https ca-certificates curl gpg
  - mkdir -p /etc/apt/keyrings/ && curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  - echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list

  # - curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/kubernetes-archive-keyring.gpg
  # - echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | tee /etc/apt/sources.list.d/kubernetes.list

  

  # Install Kubernetes Components
  - apt-get update && apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl
  - systemctl enable --now kubelet

  # Initialize Kubernetes (For Cilium we need to skip kube-proxy)
  - kubeadm init --skip-phases=addon/kube-proxy | tee /home/ubuntu/kubeadm-init.log
  # - kubeadm init | tee /home/ubuntu/kubeadm-init.log

  # Setup Kubeconfig
  - mkdir -p /home/ubuntu/.kube
  - cp /etc/kubernetes/admin.conf /home/ubuntu/.kube/config
  - chown $(id -u ubuntu):$(id -g ubuntu) /home/ubuntu/.kube/config
  - export KUBECONFIG=/home/ubuntu/.kube/config

  # Install Helm
  - curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  - |
    until kubectl get nodes; do
      echo "Waiting for Kubernetes API Server to become ready..."
      sleep 5
    done

  # Allows Scheduling of Pods onto Control Plane
  - kubectl taint nodes --all node-role.kubernetes.io/control-plane-
  - kubectl label nodes --all node.kubernetes.io/exclude-from-external-load-balancers-
  
  # Setup Cilium CLI
  - |
    CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
    CLI_ARCH=amd64
    if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi
    curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
    sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
    tar xzvf cilium-linux-${CLI_ARCH}.tar.gz -C /usr/local/bin
    rm cilium-linux-${CLI_ARCH}.tar.gz cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
    cilium version --client

  # Setup Hubble CLI
  - |
    HUBBLE_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/hubble/master/stable.txt)
    HUBBLE_ARCH=amd64
    if [ "$(uname -m)" = "aarch64" ]; then HUBBLE_ARCH=arm64; fi
    curl -L --fail --remote-name-all https://github.com/cilium/hubble/releases/download/$HUBBLE_VERSION/hubble-linux-${HUBBLE_ARCH}.tar.gz{,.sha256sum}
    sha256sum --check hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum
    tar xzvf hubble-linux-${HUBBLE_ARCH}.tar.gz -C /usr/local/bin
    rm hubble-linux-${HUBBLE_ARCH}.tar.gz hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum

  - cilium install --set kubeProxyReplacement=strict
  - cilium status --wait



`

/*
    # Clone Cilium repo and install Cilium
  - |
    sudo -u ubuntu -i bash -c '
    git clone https://github.com/cilium/cilium.git
    cd cilium
    cilium install --chart-directory ./install/kubernetes/cilium --set=kubeProxyReplacement=true
    cilium hubble enable
    '
*/
