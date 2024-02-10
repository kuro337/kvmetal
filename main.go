package main

import (
	"log"

	"kvmgo/cli"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cli.Evaluate()

	// vm.LaunchCluster("kubecontrol", "kubeworker")

	//	vm.LaunchKubeControlNode("kubecontrol")
	//  vm.LaunchKubeWorkerNode("kubeworker", "kubecontrol")

	//	vm.FullCleanup("kubecontrol")

	// vm.FullCleanup("kubeworker")
}

/*

 */
