package vm

import (
	"fmt"
	"log"
	"time"

	"kvmgo/utils"
)

func LaunchCluster(master_name, worker_name string) error {
	utils.LogMainAction("Launching & Configuring Master Node")

	master, err := LaunchKubeControlNode(master_name)
	if err != nil {
		utils.LogError("Failed during Master Node setup")
		return fmt.Errorf("Failed during Master Node setup")
	}

	utils.LogMainAction("Launching & Configuring Worker Node")

	worker, err := LaunchKubeWorkerNode(worker_name, master_name)
	if err != nil {
		utils.LogError("Failed during Master Node setup")
		return fmt.Errorf("Failed during Master Node setup")
	}

	log.Printf("Pausing Briefly before performing the Health Check")

	time.Sleep(10 * time.Second)

	utils.LogMainAction("Checking Cluster Health")

	master_client, err := master.GetSSHClient()
	if err != nil {
		utils.LogError(fmt.Sprintf("Master SSH Client Creation Failed ERROR:%s", err))
		return err
	}

	healthy, _ := ClusterHealthCheck(master_client, worker.VMName)

	if healthy {
		log.Printf("Kubernetes Cluster Successfully Launched and Healthy")
	} else {
		utils.LogError("Cluster Health Checks Failed")
		return fmt.Errorf("Cluster Launched but Health Checks Failed")
	}

	return nil
	//	FullCleanup("kubeworker")
	//	FullCleanup("kubecontrol")
}
