package main

import (
	"log"

	"kvmgo/cli"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// network.PrivateIPAddrAllVMs(true)
	// network.GetHostIP(true)
	// network.VMIpAddrInfoList(true)

	cli.Evaluate()

	// vm.LaunchCluster("kubecontrol", "kubeworker")

	//	vm.LaunchKubeControlNode("kubecontrol")
	//  vm.LaunchKubeWorkerNode("kubeworker", "kubecontrol")

	//	vm.FullCleanup("kubecontrol")

	// vm.FullCleanup("kubeworker")
}

/*

Taking snapshots of a VM

1. while running - detach raw disk
virsh detach-disk spark --target vdb

2. take snapshot
virsh snapshot-create-as --domain spark spark_hadoop --description "Machine with Spark,Hadoop,Java,Scala configured"

3. Reattach user-data.img raw disk
virsh attach-disk spark /home/kuro/Documents/Code/Go/kvmgo/data/artifacts/spark/userdata/user-data.img vdb --cache none

4. To restore the VM to the snapshot
virsh snapshot-revert --domain spark spark_hadoop

5. Permanently Deleting
virsh snapshot-delete --domain spark --snapshotname <snapshot-name>

virsh shutdown spark
*/
