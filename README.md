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

# Expose the VM on Port 8081 to an external IP
kvmetal --expose-vm=hadoop --port=8081 --hostport=8003 --external-ip=192.168.1.224 --protocol=tcp

# Cleanup Resources
kvmetal --cleanup=hadoop

# To Change the OS of the VM launch with an os-img or link to a Cloud ISO Image
kvmetal --launch-vm=mymachine --mem=24576 --cpu=8 --os-img=ubuntu23.04

```

## Distributed Event Brokers

```bash
# Launch Kafka in Kraft Mode
kvmetal --launch-vm=kafka  --preset=kafka --mem=8192 --cpu=4

# Launch Redpanda
kvmetal --launch-vm=rpanda --preset=redpanda --mem=8192 --cpu=4

```

## Big Data

```bash
# Launch a Hadoop Node with HDFS configured
kvmetal --launch-vm=hadoop --preset=hadoop --mem=8192 --cpu=4

# Launch a Spark Node with Hadoop configured
kvmetal --launch-vm=spark --preset=spark --mem=8192 --cpu=4
```

Prerequisites

```bash
sudo apt install -y qemu qemu-kvm libvirt-daemon libvirt-clients bridge-utils virt-manager cloud-image-utils libguestfs-tools


```

/home/kuro/.vscode-server/cli/servers/Stable-863d2581ecda6849923a2118d93a088b0745d9d6/server/node packages.md

/home/kuro/.vscode-server/cli/servers/Stable-863d2581ecda6849923a2118d93a088b0745d9d6/server/bin/code packages.md

$ /home/<user>/.vscode-server/bin/054a9295330880ed74ceaedda236253b4f39a335/bin/code myfile.txt
