# pure go library for using k8 with kvm

kvm

```bash
# install kvm and required libs
sudo apt install -y qemu qemu-kvm libvirt-daemon libvirt-clients bridge-utils virt-manager cloud-image-utils libguestfs-tools

sudo reboot

# free open source cloud img from Ubuntu
wget https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img

# create a writeable clone of the boot drive
# creating a qcow : qemu copy on write
qemu-img create -b ubuntu-18.04-server-cloudimg-amd64.img -F qcow2 -f qcow2 ubuntu-vm-disk.qcow2 20G

## creating a user-data file to define username/pass
cat >user-data.txt <<EOF
#cloud-config
password: password
chpasswd: { expire: False }
ssh_pwauth: True
EOF
# create .img file from config
cloud-localds user-data.img user-data.txt

# virt install with options explanation
virt-install --name ubuntu-vm \
  --virt-type kvm \
  --os-type Linux --os-variant ubuntu18.04 \
  --memory 2048 \
  --vcpus 2 \
  --boot hd,menu=on \
  --disk path=ubuntu-vm-disk.qcow2,device=disk \
  --disk path=user-data.img,format=raw \
  --graphics none \
  --noautoconsole

```

ssh to node

```bash
# getting VM IP

sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml kubecontrol | awk -F"'" '/mac address/{print $2}') | awk '{print $1}'

#  Copy file to VM
scp /path/to/your/jar/MySparkApp-assembly-0.1.jar username@192.168.122.91:/destination/path

# can only sent to /tmp/
scp NOTES.md ubuntu@192.168.122.91:/tmp/

```

snapshotting the VM

```bash

virsh domblklist spark # view attached disks

# We can only take snapshots of qcow2 - but the userdata is in raw format
# workaround is to detach the secondary user-data disk , and take the backup

# while the vm is running detatch it then run next steps
virsh detach-disk spark --target vdb

virsh snapshot-create-as --domain spark spark_hadoop --description "Machine with Spark,Hadoop,Java,Scala configured"

virsh attach-disk spark /home/kuro/Documents/Code/Go/kvmgo/data/artifacts/spark/userdata/user-data.img vdb --cache none

# To restore the VM to the snapshot
virsh snapshot-revert --domain spark spark_hadoop

virsh snapshot-delete --domain spark --snapshotname <snapshot-name>


# confirm disk image format
qemu-img info /home/kuro/Documents/Code/Go/kvmgo/data/artifacts/spark/userdata/user-data.img

# we curr store the userdata.img disk in data/artifacts/spark/userdata/user-data.img
virsh shutdown spark

# Snapshotting a VM (can be done while its running or shut down - running preferred)
virsh snapshot-create-as --domain spark spark_hadoop --description "Machine with Spark,Hadoop,Java,Scala configured"

virsh snapshot-create-as --domain vmName snapshotName --description "spark VM"

# Restoring a VM (can be done while its NOT running or in a Paused State)
virsh snapshot-revert --domain vmName snapshotName

virsh snapshot-list spark
virsh domblklist spark


```

libvirtd

```bash
sudo systemctl restart libvirtd
sudo usermod -aG libvirt kuro
sudo reboot

lscpu | grep '^CPU(s):' # viewing cpus available on host


# detailed info such as memory and cpu
virsh dominfo <domain-name>
virsh list --all
virsh reboot hadoop # in case we lose access during console session 
virsh shutdown ubuntu-vm
virsh suspend ubuntu-vm
virsh resume ubuntu-vm
virsh undefine  # delete .qcow2 file after undefining to completely clear it
virsh undefine hadoop --remove-all-storage # completely clears the VM including Snapshots

# in case we ever lose access to a VM session and need to kill the console
ps aux | grep 'virsh console hadoop'
kill -9 pid

sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml hadoop | awk -F"'" '/mac address/{print $2}') | awk '{print $1}'


sudo apt-get install arp-scan

# getting mac address
virsh dumpxml worker | grep 'mac address'
# getting IP from mac
sudo arp-scan --interface=virbr0 --localnet | grep "52:54:00:25:40:cb"


sudo virt-ls -d vmname /home/ubuntu/
sudo virt-copy-out -d vmname /root/init.log /tmp/extract/
sudo virt-copy-in -a pathto/vm-disk.qcow2 file-to-copy /path/in/vm

```

```bash
# Check if a mnt point is in use

sudo lsof /mnt/vmName
sudo fuser -vm /mnt/vmName

sudo guestunmount /mnt/controlplanevm

sudo guestmount -a %s -i --rw %s", absImagePath, mountPath
```

virsh networking common issues

```bash
# if default network bridge is not seen on normal user

virsh net-list --all
sudo virsh net-list --all


sudo systemctl restart libvirtd
virsh net-start default
virsh net-autostart default

# run this and log back in

sudo usermod -aG libvirt kuro
cp /etc/libvirt/libvirt.conf ~/.config/libvirt/

# uncomment last line at vi /etc/libvirt/libvirt.conf

sudo vi /etc/libvirt/libvirt.conf

#uri_default="qemu:///system"

# now it should work and show up
virsh net-list --all

# make sure user is set to <user> kuro and group is libvirt
# by default it will be root and root

sudo vi /etc/libvirt/qemu.conf

user = "kuro"
group = "libvirt"


```

#### testing connectivity between Host and VM

```bash

# run this to get the Mac Address of the VM
virsh dumpxml worker | grep 'mac address'

# get the IP of VM host from Mac Address
sudo arp-scan --interface=virbr0 --localnet | grep "52:54:00:25:40:cb"

192.168.122.103

sudo apt-get install arp-scan
sudo arp-scan --localnet | grep "<MAC_ADDRESS>"

# data should flow 2 ways

ping 192.168.122.103

```

#### kubeadm

```bash
# Kubeadm Join Command:

kubeadm join 192.168.122.3:6443 --token z6tqnt.fcuk6iw86n4hccg8 \
	--discovery-token-ca-cert-hash sha256:d66dd8820b4966a30628b0d63d96173fced313942834f130878f71cf2af435c8


```

library usage

```go
// launching a full Cluster

func main() {

utils.Help()

	vm.LaunchKubeControlNode()
	vm.LaunchKubeWorkerNode()

  healthy, _ := vm.ClusterHealthCheck()

  if healthy == true {
      log.Printf("Success!")
	} else {
    	log.Printf("Health Checks Failed")
	}

	vm.FullCleanup("kubeworker")

	vm.FullCleanup("kubecontrol")
}




```

```go
func main() {
  /* Creating a VM as required with any Config */

  // provide basic overview of required params

	config := NewVMConfig("kubecontrol").
		SetImageURL("https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img").
		SetImagesDir("data/images").
		SetBootFilesDir("data/scripts/master_kube").
		DefaultUserData().
		SetBootServices([]string{"kubemaster.service"}).
		SetCores(2).
		SetMemory(2048).
		SetArtifacts([]string{"/home/ubuntu/kubeadm-init.log"})
}

```
