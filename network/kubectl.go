package network

import (
	"fmt"
	"log"
	"strings"
	"time"

	"kvmgo/utils"
)

/*
Creates the Deployment if it does not exist - and explicitly checks if it already exists

Usage:

		  result, err := client.ManageDeployment("nginx", "nginx")

			if err != nil {
	        log.Printf("Error managing deployment: %v", err)
	        return
	    }

	    log.Printf(result)
*/
func (vm *VMClient) ManageDeployment(deploymentName, imageName string) (string, error) {
	// First, check if the deployment already exists
	checkCmd := fmt.Sprintf("kubectl get deployment %s", deploymentName)
	output, _, err := vm.RunCmd(checkCmd)
	if err == nil && strings.Contains(output, deploymentName) {
		// Deployment already exists
		return "Deployment already exists", nil
	}

	// If deployment does not exist, create it
	createCmd := fmt.Sprintf("kubectl create deployment %s --image=%s", deploymentName, imageName)

	log.Printf("Kicking off Deployment: %s", createCmd)

	_, stderr, err := vm.RunCmd(createCmd)
	if err != nil {
		if strings.Contains(stderr, "already exists") {
			// If the error is because the deployment already exists, it's not an actual error
			return "Deployment already exists", nil
		}
		// Actual error while creating the deployment
		return "", fmt.Errorf("failed to create deployment: %v, stderr: %s", err, stderr)
	}

	// Deployment created successfully
	return "Deployment created successfully", nil
}

/*
Checks if a Deployment is Ready - gets the Pod names first from the VM
Then it extracts the deployment name - and checks if the Pod is Ready

Usage:

	func main() {


	    ready, err := client.WaitForDeploymentReadiness("nginx")
	    if err != nil {
	        log.Printf("Error waiting for deployment readiness: %v", err)
	        return
	    }

	    if ready {
	        log.Println("Nginx deployment is ready")
	    }


	}
*/
func (vm *VMClient) WaitForDeploymentReadiness(deploymentName string) (bool, error) {
	deploymentRetryIntervals := []int{10, 15, 25, 30, 45} // Retry intervals in seconds

	for i, interval := range deploymentRetryIntervals {
		podNames, _, err := vm.RunCmd(fmt.Sprintf("kubectl get pods -l app=%s -o jsonpath='{.items[*].metadata.name}'", deploymentName))
		if err != nil {
			// Log error but don't return immediately. Continue retrying.
			utils.LogError(fmt.Sprintf("Error listing pods for deployment %s in retry %d: %v", deploymentName, i, err))

			// Sleep before retrying if not the last interval
			if i < len(deploymentRetryIntervals)-1 {
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			} else {
				return false, fmt.Errorf("failed to list pods for deployment %s after retries: %v", deploymentName, err)
			}
		}

		// Check readiness for each pod
		for _, podName := range strings.Fields(podNames) {
			if err := WaitForPodReadiness(vm, podName); err != nil {
				return false, fmt.Errorf("pod %s is not ready: %v", podName, err)
			}
		}

		return true, nil
	}

	return false, fmt.Errorf("deployment %s did not reach readiness within the specified retry intervals", deploymentName)
}
