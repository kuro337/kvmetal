# OpenEBS

for kvm

```bash
helm repo add openebs https://openebs.github.io/charts
helm repo update

helm install openebs openebs/openebs --namespace openebs --create-namespace \
--set legacy.enabled=false \
--set lvm-localpv.enabled=true

kubectl get pods -n openebs # 7
kubectl get storageclass # openebs-device, openebs-hostpath


lsblk # List Disks and Partitions


lsblk
  - vda   (Virtual Disk A)
  - vda14 (metadata disk)
  - vda15 (/boot/efi boot info)

mount | grep /dev/sdb # List Mounted Filesystems
```

`kubectl get storageclass`

```bash
helm repo add openebs https://openebs.github.io/charts
helm repo update

# (Default) Install Jiva, cStor and Local PV with out-of-tree provisioners
helm install openebs --namespace openebs openebs/openebs --create-namespace

# For LVM Local (high perf, recoverable)

https://github.com/openebs/lvm-localpv
# Install LVM Local PV
helm install openebs openebs/openebs --namespace openebs --create-namespace \
--set legacy.enabled=false \
--set lvm-localpv.enabled=true

# Uninstall
helm uninstall openebs -n openebs
kubectl get all -n openebs

# monitor
kubectl get pods -n openebs
helm ls -n openebs
kubectl get storageclass

```

## Create Storage Group for Nodes

We need to assign a Disk Partition for OpenEBS to use.

```bash
# On each Node
# Initialize Physical Volume (PV)
sudo pvcreate /dev/sdb

# Init Volume Group (VG)
sudo vgcreate lvmvg /dev/sdb

# Check available options
ls /dev/

sda/sdb -> SATA Drives
nvmeXn1 -> NVMe Drives
mmcblkX -> SD Cards/MMC Devices
```

## Create Storage Class

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvmpv
parameters:
  storage: "lvm"
  volgroup: "lvmvg"
provisioner: local.csi.openebs.io
```

## Engines https://openebs.io/docs/concepts/casengines#data-engine-capabilities

cStore :

- Multiple Disks on Nodes
- Building k8 native storage services similar to EBS

LocalPV :

- Offers Near Disk Perf
- ideal for statefulsets, apps share host disk
- LocalPV has types such as HostPath, LVM , ZFS.
