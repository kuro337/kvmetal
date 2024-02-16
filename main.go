package main

import (
	"log"

	"kvmgo/cli"
	"kvmgo/network"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cli.Evaluate()

	network.GetHostIP(true)
	network.PrivateIPAddrAllVMs(true)
	network.VMIpAddrInfoList(true)

	rule := network.CreateUfwBeforeRule("192.168.1.194", "8088", "9999")
	log.Printf("ufw rule: %s", rule)

	// hadoop := utils.GenerateDefaultCloudInitZshKernelUpgrade("hadoop")
	// log.Printf("%s", hadoop)
	// os.WriteFile("test.txt", []byte(hadoop), 0o644)

	// config, err := configuration.NewConfigBuilder(
	// 	constants.Ubuntu,
	// 	[]constants.Dependency{
	// 		constants.Zsh,
	// 		constants.Hadoop,
	// 	},
	// 	[]constants.CloudInitPkg{
	// 		constants.OpenJDK11,
	// 		constants.Git,
	// 		constants.NetTools,
	// 		constants.Curl,
	// 	},
	// 	"ubuntu", "password", "hadoop")
	// if err != nil {
	// 	log.Printf("Failed to create Configuration")
	// }

	// userdata := config.CreateCloudInitData()

	// log.Printf("Userdata %s", userdata)

	// os.WriteFile("testuserdata.yaml", []byte(userdata), 0o644)
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
