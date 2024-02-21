# kvmetal

## launch and manage virtual machines and isolated clusters

```bash
# build
git clone https://github.com/kuro337/kvmetal
go build -o kvmetal

# Launch a VM with 24gb memory and 8 vcpus
kvmetal --launch-vm=mymachine --mem=24576 --cpu=8

# Launch a Kubernetes cluster with 1 Control Node and 2 Workers
kvmetal --cluster --control=kubecontrol --workers=kubeworker1,kubeworker2

# Launch a Hadoop Node with HDFS configured
kvmetal --launch-vm=hadoop --preset=hadoop --mem=8192 --cpu=4

# Launch a Spark Node with Hadoop configured
kvmetal --launch-vm=hadoop --preset=hadoop --mem=8192 --cpu=4

# Expose the VM on Port 8081 to an external IP
kvmetal --expose-vm=hadoop --port=8081 --hostport=8003 --external-ip=192.168.1.224 --protocol=tcp

# Cleanup Resources
kvmetal --cleanup=hadoop

# To Change the OS of the VM launch with an os-img or link to a Cloud ISO Image
kvmetal --launch-vm=mymachine --mem=24576 --cpu=8 --os-img=ubuntu23.04

```

Prerequisites

```bash
sudo apt install -y qemu qemu-kvm libvirt-daemon libvirt-clients bridge-utils virt-manager cloud-image-utils libguestfs-tools

```
