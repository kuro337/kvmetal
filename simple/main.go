package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func exists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

// Downloads Base Linux Cloud Image to data/images - only done once and shared among VM's in data/images/ubuntu.img
// PullImage(url,"/home/kuro/Documents/Code/Go/kvmgo/simple") will download to
//
//	/home/kuro/Documents/Code/Go/kvmgo/simple/ubuntu-22.04-server-cloudimg-amd64.img
func PullImage(url, dir string) error {
	imageName := filepath.Base(url)
	imagePath := filepath.Join(dir, imageName)
	pullImgsStr := fmt.Sprintf("Pulling Base Image: URL:%s, Dir:%s, ImgPath: %s\n", url, dir, imagePath)
	log.Println(pullImgsStr)

	if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
		log.Printf("Image %s already exists", imageName)
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ExecCmd("ls -a",true) will run a command and return the full string result and print/not print
func ExecCmd(command string, print bool) (string, error) {
	// Splitting command into command and arguments
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	// Log the output, whether successful or not
	if print {
		log.Printf("%s", out.String())
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

var (
	imgFile = "ubuntu-24.04-server-cloudimg-amd64.img"
	url     = "https://cloud-images.ubuntu.com/releases/server/server/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img"

	qcowImg     = "ubuntu-vm-disk.qcow2"
	vmName      = "ubuntu-base-vm"
	userDataImg = "user-data.img"

	workerName        = "kube-worker"
	qcowWorker        = "kube-worker.qcow2"
	workerUserdataImg = "user-data-worker.img"
)

func stageUbuntuImg() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
	}
	log.Printf("Current working directory: %s", cwd)

	if err := PullImage(url, cwd); err != nil {
		return "", fmt.Errorf("error pulling:%s", err)
	}
	return filepath.Join(cwd, filepath.Base(url)), nil
}

// use this + user-data.img in virt-install

func writeNew(path, content string) error {
	if e := exists(path); e {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to clean existing file %s : %s", path, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", path, err)
	}
	return nil
}

/*
working config

	virt-install \
				  --name=ubuntu-base-vm \
				  --virt-type=kvm \
				  --memory=8192 \
				  --vcpus=4 \
		      --disk path=ubuntu-vm-disk.qcow2,device=disk \
				  --disk path=user-data.img,format=raw \
				  --graphics none \
				  --boot hd,menu=on \
				  --network network=default \
				  --os-variant ubuntu22.04 \
				  --cloud-init user-data=user-data.txt,network-config=network-config.yaml \
				  --noautoconsole \
				  --check all=off \
				  -d
*/

func workingConfig(qcowDiskFile string) {
	userDataContent := `#cloud-config
password: password
chpasswd: { expire: False }
ssh_pwauth: True

packages:
  - zsh
  - git
  - curl

package_update: true
package_upgrade: true`

	createUserDataImgCmd := "cloud-localds user-data.img user-data.txt"

	networkCfg := `network:
  version: 2
  ethernets:
    enp1s0:
      dhcp4: true`

	virtInstallCmd := fmt.Sprintf(`virt-install --name %s \
--virt-type kvm \
--os-type Linux --os-variant ubuntu18.04 \
--memory 2048 \
--vcpus 2 \
--boot hd,menu=on \
--disk path=%s,device=disk \
--disk path=%s,format=raw \
--graphics none \
--cloud-init user-data=user-data.txt,network-config=network-config.yaml \
--noautoconsole`, vmName, qcowDiskFile, userDataImg)

	log.Print(virtInstallCmd, networkCfg, userDataContent, createUserDataImgCmd)
}

func createVMAndRun(vmName, qcowDisk, vmUserdataImg, runCmd string) error {
	// Write user-data file

	runcmd := fmt.Sprintf(`runcmd:
%s`, runCmd)
	if runCmd == "" {
		runcmd = ""
	}

	userDataContent := fmt.Sprintf(`#cloud-config
password: password
chpasswd: { expire: False }
ssh_pwauth: true
package_update: true
package_upgrade: true
ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDj4wnyp6PWYdwtJIEj8Kc6eomUCzD6/pesIbIFABeeOPLRUL12s60PNObkevBBOZglHfLJScEOqkxMo1sC99PeeZ+vQbQDFXex8b7TzUHpwq/KpupC7cAjQ1wWnzcZSkgq5/zGaIT+WE8SS+bKk+Ar7Z89nJ6wgU+kbtB5GvOaU0DZdZL13ztRLRseOU57oaJ8TP4nMSIW2wP0Q27e63vPSBo/FjB3V6w4PF61guC2dlcfJJSFcxdag4rERP+3lZDMffO9KlHvGlnkk0j/nK5b558VSycMI3jiKldPBpH/qTBy3IH5kCaWsQ2GpW6HZWtYYEEZqJ1FuFz0QcobKRw5

%s
`, runcmd)

	if err := writeNew("user-data.txt", userDataContent); err != nil {
		return fmt.Errorf("failed to write user-data.txt: %v", err)
	}

	// Write network-config file
	networkCfg := `network:
  version: 2
  ethernets:
    enp1s0:
      dhcp4: true`

	if err := writeNew("network-config.yaml", networkCfg); err != nil {
		return fmt.Errorf("failed to write network-config.yaml: %v", err)
	}

	if err := os.Remove(qcowDisk); err != nil {
		return fmt.Errorf("failed to clean existing qCow img: %s", err)
	}

	qemuCmd := fmt.Sprintf("qemu-img create -b %s -F qcow2 -f qcow2 %s 20G", imgFile, qcowImg)
	if _, err := ExecCmd(qemuCmd, true); err != nil {
		return fmt.Errorf("failed to create qcow2 disk from img %s", err)
	}

	// user-data.img
	if err := os.Remove(vmUserdataImg); err != nil {
		return fmt.Errorf("failed to clean existing user-data img: %s", err)
	}
	// Create user-data.img
	// createUserDataImgCmd := exec.Command("cloud-localds", "user-data.img", "user-data.txt")
	// if output, err := createUserDataImgCmd.CombinedOutput(); err != nil {
	// 	return fmt.Errorf("failed to create user-data.img: %v\nOutput: %s", err, output)
	// }

	// user-data.img
	if _, err := ExecCmd(fmt.Sprintf("cloud-localds %s user-data.txt", vmUserdataImg), true); err != nil {
		return fmt.Errorf("failed to create user-data disk from img %s", err)
	}
	vcpu := 4
	mem := 8192

	// Run virt-install
	virtInstallCmd := fmt.Sprintf(`virt-install --name %s \
--virt-type kvm \
--os-variant ubuntu24.04 \
--memory %d \
--vcpus %d \
--boot hd,menu=on \
--disk path=%s,device=disk \
--disk path=user-data.img,format=raw \
--cloud-init network-config=network-config.yaml \
--graphics none \
--noautoconsole`, vmName, mem, vcpu, qcowDisk)

	// --cloud-init user-data=user-data.txt,network-config=network-config.yaml \
	// --cloud-init network-config=network-config.yaml \
	// --cloud-init user-data=user-data.txt,network-config=network-config.yaml \

	cmd := exec.Command("bash", "-c", virtInstallCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run virt-install: %v\nOutput: %s", err, output)
	}

	fmt.Printf("VM '%s' created and started successfully.\n", vmName)
	return nil
}

func virtInstall() error {
	cmdArgs := []string{
		"virt-install",
		"--name=ubuntu-base-vm",
		"--virt-type=kvm",
		"--memory=8192",
		"--vcpus=4",
		"--disk", "path=" + qcowImg + ",device=disk",
		"--disk", "path=user-data.img,format=raw",
		"--graphics", "none",
		"--boot", "hd,menu=on",
		"--network", "network=default",
		"--os-variant=ubuntu24.04", // 22.04
		// "--network", "bridge=virbr0,model=virtio",
		// "--cloud-init user-data=user-data.yaml,network-config=network-config.yaml",
		// "--cloud-init user-data=user-data.yaml,network-config=network-config.yaml",
		"--noautoconsole",
		"--check",
		"-d",
	}
	cmdStr := strings.Join(cmdArgs, " ")
	log.Print(cmdStr)

	if _, err := ExecCmd(cmdStr, true); err != nil {
		return fmt.Errorf("failed to run virt-install %s", err)
	}

	// virsh destroy ubuntu-base-vm && virsh undefine ubuntu-base-vm
	return nil
}

// sudo -i

const KUBE_RUNCMD = `  - swapoff -a
  - apt-get update
  - apt-get install make
  - apt-get install unzip
  - modprobe overlay
  - modprobe br_netfilter

  - |
    wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile.d/go.sh
    chmod +x /etc/profile.d/go.sh
 

  - wget https://github.com/containerd/containerd/releases/download/v1.7.22/containerd-1.7.22-linux-amd64.tar.gz
  - tar Cxzvf /usr/local containerd-1.7.22-linux-amd64.tar.gz
  - wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
  - mkdir -p /usr/local/lib/systemd/system/
  - mv containerd.service /usr/local/lib/systemd/system/
  - systemctl daemon-reload
  - systemctl enable --now containerd
  
  - wget https://github.com/opencontainers/runc/releases/download/v1.1.14/runc.amd64
  - install -m 755 runc.amd64 /usr/local/sbin/runc
  - runc --version
  - wget https://github.com/containernetworking/plugins/releases/download/v1.5.1/cni-plugins-linux-amd64-v1.5.1.tgz
  - mkdir -p /opt/cni/bin
  - tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.5.1.tgz

  - mkdir -p /etc/containerd
  - containerd config default | sudo tee /etc/containerd/config.toml
  - echo 'set SystemdCgroup = true in /etc/containerd/config.toml'
  - sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
  - apt-get update

  - apt-get install -y apt-transport-https ca-certificates curl gpg
  - curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  - echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
  - apt-get update
  - apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl
  - systemctl enable --now kubelet

  - echo -e "net.bridge.bridge-nf-call-iptables = 1\nnet.bridge.bridge-nf-call-ip6tables = 1\nnet.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-kubernetes-cri.conf
  - sysctl --system
  - apt update
  - apt install -y socat

  - kubeadm init --pod-network-cidr=10.244.0.0/16 | tee /home/ubuntu/kubeadm-init.log
  - mkdir -p /home/ubuntu/.kube
  - cp -i /etc/kubernetes/admin.conf /home/ubuntu/.kube/config
  - chown -R ubuntu:ubuntu /home/ubuntu/.kube
  - export KUBECONFIG=/etc/kubernetes/admin.conf
  - kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml

  - |
    if [ ! -f /home/ubuntu/.bashrc ]; then
      touch /home/ubuntu/.bashrc
      chown ubuntu:ubuntu /home/ubuntu/.bashrc
    fi
    echo "source <(kubectl completion bash)" >> /home/ubuntu/.bashrc
    echo "alias k=kubectl" >> /home/ubuntu/.bashrc
    echo "complete -o default -F __start_kubectl k" >> /home/ubuntu/.bashrc

  - systemctl start kubelet
  `

const KUBE_WORKER_CMD = `
  - swapoff -a
  - apt-get update
  - apt-get install -y make unzip
  - modprobe overlay
  - modprobe br_netfilter
  - wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
  - sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
  - export PATH=$PATH:/usr/local/go/bin
  - echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile.d/go.sh
  - chmod +x /etc/profile.d/go.sh
  - wget https://github.com/containerd/containerd/releases/download/v1.7.22/containerd-1.7.22-linux-amd64.tar.gz
  - tar Cxzvf /usr/local containerd-1.7.22-linux-amd64.tar.gz
  - wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
  - mkdir -p /usr/local/lib/systemd/system/
  - mv containerd.service /usr/local/lib/systemd/system/
  - systemctl daemon-reload
  - systemctl enable --now containerd
  - wget https://github.com/opencontainers/runc/releases/download/v1.1.14/runc.amd64
  - install -m 755 runc.amd64 /usr/local/sbin/runc
  - wget https://github.com/containernetworking/plugins/releases/download/v1.5.1/cni-plugins-linux-amd64-v1.5.1.tgz
  - mkdir -p /opt/cni/bin
  - tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.5.1.tgz
  - mkdir -p /etc/containerd
  - containerd config default | sudo tee /etc/containerd/config.toml
  - sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
  - apt-get update
  - apt-get install -y apt-transport-https ca-certificates curl gpg
  - curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  - echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
  - apt-get update
  - apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl
  - systemctl enable --now kubelet
  - echo -e "net.bridge.bridge-nf-call-iptables = 1\nnet.bridge.bridge-nf-call-ip6tables = 1\nnet.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-kubernetes-cri.conf
  - sysctl --system
  - apt update
  - apt install -y socat
  `

func main() {
	if err := createVMAndRun(vmName, qcowImg, "user-data.img", KUBE_RUNCMD); err != nil {
		log.Fatalf("Failed to create VM: %s", err)
	}

	if err := createVMAndRun(workerName, qcowWorker, workerUserdataImg, KUBE_WORKER_CMD); err != nil {
		log.Fatalf("Failed to create VM: %s", err)
	}

	// imagePath, err := stageUbuntuImg()
	// if err != nil {
	// 	log.Printf("Error pulling image: %v", err)
	// }
	// log.Printf("pulled to %s", imagePath)

	// if err := qcowDisk(); err != nil {
	// 	log.Print("QCOW CMD RUNNING - ERROR HERE")
	// 	log.Fatal(err)
	// 	return
	// }

	// if err := virtInstall(); err != nil {
	// 		log.Print("virt-install failed")

	// 		log.Fatal(err)
	// 	}
	// cat /etc/netplan/01-netcfg.yaml

	// sudo journalctl -b

	// sudo cloud-init query userdata

	// ens1:
	// dhcp4: true

	// sudo dhclient enp1s0
	// ip addr show enp1s0

	// sudo cloud-init schema --system
	// sudo cloud-init query userdata

	// viewing user-data used
	// sudo cat /var/lib/cloud/instance/user-data.txt

	// sudo cat /var/log/cloud-init-output.log
	// sudo cat /var/log/cloud-init.log
	// sudo grep -i kubectl /var/log/cloud-init.log
	// sudo grep -i error /var/log/cloud-init-output.log
	// error logs sudo grep -i error /var/log/cloud-init.log

	// sudo /var/log/apt
}

/*
sudo apt-get update && sudo apt-get install containerd

sudo mkdir -p /etc/containerd
sudo containerd config default | sudo tee /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
systemctl restart containerd
systemctl enable containerd



*/

const NEW_KUBE_CONTROL = `  - swapoff -a
  - apt-get update
  - apt-get install -y containerd
  - |
    if ! command -v containerd &> /dev/null; then
        echo "ERROR: containerd installation failed"
        exit 1
    fi
  - mkdir -p /etc/containerd
  - containerd config default | tee /etc/containerd/config.toml
  - sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
  - systemctl restart containerd
  - systemctl enable containerd
  - |
    if ! systemctl is-active --quiet containerd; then
        echo "ERROR: containerd is not running"
        exit 1
    fi

  - modprobe overlay
  - modprobe br_netfilter

  - |
    echo "net.bridge.bridge-nf-call-iptables  = 1
    net.ipv4.ip_forward                 = 1
    net.bridge.bridge-nf-call-ip6tables = 1" | tee /etc/sysctl.d/99-kubernetes-cri.conf
  - sysctl --system

  - apt-get install -y apt-transport-https ca-certificates curl gpg
  - mkdir -p -m 755 /etc/apt/keyrings
  - curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  - echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list

  - apt-get update && apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl
  - systemctl enable --now kubelet

  - kubeadm init --skip-phases=addon/kube-proxy | tee /home/ubuntu/kubeadm-init.log

  - |
    if [ $? -ne 0 ]; then
     echo "kubeadm init failed. Check /home/ubuntu/kubeadm-init.log for details."
     exit 1
    fi

  - mkdir -p /home/ubuntu/.kube
  - cp /etc/kubernetes/admin.conf /home/ubuntu/.kube/config
  - chown $(id -u ubuntu):$(id -g ubuntu) /home/ubuntu/.kube/config
  - export KUBECONFIG=/home/ubuntu/.kube/config

  - curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  - |
    until kubectl get nodes; do
      echo "Waiting for Kubernetes API Server to become ready..."
      sleep 5
    done

  - kubectl taint nodes --all node-role.kubernetes.io/control-plane-
  - kubectl label nodes --all node.kubernetes.io/exclude-from-external-load-balancers-
  
  - |
    CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
    CLI_ARCH=amd64
    if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi
    curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
    sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
    tar xzvf cilium-linux-${CLI_ARCH}.tar.gz -C /usr/local/bin
    rm cilium-linux-${CLI_ARCH}.tar.gz cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
    cilium version --client

  - |
    HUBBLE_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/hubble/master/stable.txt)
    HUBBLE_ARCH=amd64
    if [ "$(uname -m)" = "aarch64" ]; then HUBBLE_ARCH=arm64; fi
    curl -L --fail --remote-name-all https://github.com/cilium/hubble/releases/download/$HUBBLE_VERSION/hubble-linux-${HUBBLE_ARCH}.tar.gz{,.sha256sum}
    sha256sum --check hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum
    tar xzvf hubble-linux-${HUBBLE_ARCH}.tar.gz -C /usr/local/bin
    rm hubble-linux-${HUBBLE_ARCH}.tar.gz hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum

  - cilium install --set kubeProxyReplacement=strict
  - cilium status --wait`
