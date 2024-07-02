package vm

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"kvmgo/network"
	"kvmgo/utils"
)

/*
Launches a new VM running the Kubernetes Control Plane

Downloads the Ubuntu 22.04 cloud image (if not cached)

Defines the systemd services to configure the VM once kubeadm is setup
*/
func LaunchKubeControlNode(master_name string) (*VMConfig, error) {
	config := NewVMConfig(master_name).
		SetImageURL("https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img").
		SetImagesDir("data/images").
		SetBootFilesDir("data/scripts/master_kube").
		DefaultUserData().
		SetBootServices([]string{"kubemaster.service"}).
		SetCores(2).
		SetMemory(2048).
		SetArtifacts([]string{"/home/ubuntu/kubeadm-init.log"})

	config.PullImage()

	utils.LogSection("CREATING BASE IMAGE")

	if err := config.CreateBaseImage(); err != nil {
		slog.Error("Failed to Setup VM", "error", err)
		_ = Cleanup(config.VMName)
		return nil, err
	}
	utils.LogStepSuccess(fmt.Sprintf("Modified Base Image Creation Success at %s", config.ImagesDir))

	utils.LogSection("SETTING UP VM")

	if err := config.SetupVM(); err != nil {
		slog.Error("Failed to Setup VM", "error", err)
		_ = Cleanup(config.VMName)
		return nil, err

	}

	utils.LogSection("GENERATING CLOUDINIT USERDATA")

	if err := config.GenerateCustomUserDataImg(""); err != nil {
		slog.Error("Failed to Generate Cloud-Init Disk", "error", err)
		_ = Cleanup(config.VMName)
		return nil, err
	}

	utils.LogStepSuccess("Cloudinit user-data Generated Successfully")

	utils.LogSection("LAUNCHING VM")

	if err := config.CreateVM(); err != nil {
		slog.Error("Failed to Create VM", "error", err)
		return nil, err
	}

	utils.LogStepSuccess(fmt.Sprintf("VM %s created Successfully!", config.VMName))

	utils.LogSection("PULLING ARTIFACTS")

	_ = config.PullArtifacts()

	utils.LogStepSuccess("VM Healthy and Boot Artifacts Pulled Successfully.")

	return config, nil
}

// ClusterHealthCheck tests an Nginx deployment using the newly setup Cluster - validating Worker and Control Plane Health
func ClusterHealthCheck(master_client *network.VMClient, worker_name string) (bool, error) {
	deployment_success, err := TestNginxDeployment(master_client, worker_name)
	if err != nil {
		log.Printf("Failed Executing Nginx Deployment ERROR:%s", err)
		return false, err
	}

	if !deployment_success {
		log.Printf("Health Checks were Unsuccessful for Cluster")
		return false, nil
	}
	log.Printf("Health Checks Successful!")

	return true, nil
}

/*
Deploy nginx to the Cluster to validate that deployments are functional
Confirms valid setup of a Cluster with a Control Plane and Worker and Network Flow

	// Kubectl Equivalent of this Test:

	kubectl create deployment nginx --image=nginx
	kubectl expose deployment nginx --port=80 --type=NodePort
	kubectl get svc
	kubectl get pods -o wide

	POD_NAME=$(kubectl get pods -o custom-columns=":metadata.name" --no-headers=true)
	kubectl port-forward pod/$POD_NAME 8080:80 &
	curl localhost:8080
	sleep 5
	kubectl delete deployment nginx
	kubectl delete svc nginx
*/
func TestNginxDeployment(client *network.VMClient, worker_name string) (bool, error) {
	utils.LogStep(fmt.Sprintf("Waiting on Node Readiness %s", worker_name))

	/* this times out - doesnt work properly */
	err := client.WaitForNodeReadiness(worker_name)
	if err != nil {
		utils.LogError(fmt.Sprintf("Failed to Check Readiness for %s", worker_name))
		return false, err
	}
	// Deploy Nginx and wait for it to be Ready

	deployment_result, err := client.ManageDeployment("nginx", "nginx")
	if err != nil {
		return false, fmt.Errorf("failed to create nginx deployment: %v", err)
	}

	log.Printf("Deployment Result: %s", deployment_result)

	deployment_ready, err := client.WaitForDeploymentReadiness("nginx")
	if err != nil || !deployment_ready {
		log.Printf("Nginx deployment is not ready: %v", err)
		return false, fmt.Errorf("failed to check Nginx deployment Readiness: %v", err)
	}

	if deployment_ready {
		log.Printf("Nginx deployment is ready")
	}

	// Expose Nginx
	if _, _, err := client.RunCmd("kubectl expose deployment nginx --port=80 --type=NodePort"); err != nil {
		return false, fmt.Errorf("failed to expose nginx deployment: %v", err)
	}

	log.Printf("Nginx deployed and exposed")

	// Get the Pod name
	podName, _, err := client.RunCmd("kubectl get pods -o custom-columns=\":metadata.name\" --no-headers=true")
	if err != nil {
		return false, fmt.Errorf("failed to get pod name: %v", err)
	}
	podName = strings.TrimSpace(podName)

	// Forward Port - run in a goroutine as this is a blocking call
	var portForwardPID int
	go func() {
		output, _, err := client.RunCmd("echo $$; exec kubectl port-forward pod/" + podName + " 8080:80")
		if err != nil {
			log.Printf("Port forwarding failed: %v", err)
			return
		}

		// Extract PID from output
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			portForwardPID, _ = strconv.Atoi(lines[0])
		}
	}()

	// Small pause to ensure port forwarding set up
	time.Sleep(2 * time.Second)

	curlOutput, _, err := client.RunCmd("curl http://localhost:8080")
	if err != nil {
		return false, fmt.Errorf("failed to perform curl request: %v", err)
	}

	// Terminate port-forwarding once done
	if portForwardPID != 0 {
		_, _, err = client.RunCmd(fmt.Sprintf("kill %d", portForwardPID))
		if err != nil {
			log.Printf("Failed to kill port-forward process: %v", err)
		}
	}

	_, _, err = client.RunCmd("kubectl delete deployment nginx")
	if err != nil {
		return false, fmt.Errorf("failed to delete deployment: %v", err)
	}

	_, _, err = client.RunCmd("kubectl delete svc nginx")
	if err != nil {
		return false, fmt.Errorf("failed to delete deployment: %v", err)
	}

	if strings.Contains(curlOutput, "Welcome to nginx") {
		return true, nil
	}

	return false, fmt.Errorf("nginx deployment test failed")
}

// cloud-localds user-data.img user-data.txt
// cloud-localds ../sampf/user-data.img user-data.txt
