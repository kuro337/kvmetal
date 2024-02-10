package network

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"kvmgo/utils"

	"golang.org/x/crypto/ssh"
)

type VMClient struct {
	VMName         string
	IP             string
	Username       string
	Password       string
	PrivateKeyPath string
	SSHClient      *ssh.Client
	printFlag      bool
}

/*
NewInsecureSSHClient creates an SSH client for a VM using password authentication.

Usage:

	ip , err := GetVMIPAddr("kubecontrol")
	log.Printf("IP of Control Node is %s",ip)

	client , _ := NewInsecureSSHClientVM("kubecontrol",ip,"ubuntu","password")

	// use GetVMIPAddr to get the IP of a running node
*/
func NewInsecureSSHClientVM(vmName, ip, username, password string) (*VMClient, error) {
	client := &VMClient{
		VMName:   vmName,
		IP:       ip,
		Username: username,
		Password: password,
	}

	config := &ssh.ClientConfig{
		User: client.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(client.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(client.IP, "22"), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	client.SSHClient = sshClient
	return client, nil
}

// RunCommand executes a command on the VM and returns its output.
func (vm *VMClient) RunCommand(command string) (string, string, error) {
	session, err := vm.SSHClient.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	if vm.printFlag == true {
		log.Printf("Running Command:%s", command)
	}
	err = session.Run(command)
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if err != nil {
		return stdout, stderr, fmt.Errorf("failed to run command: %v, stderr: %s", err, stderr)
	}

	return stdout, stderr, nil
}

func (vm *VMClient) Uptime(print bool) (string, error) {
	output, _, err := vm.RunCommand("uptime")
	if err != nil {
		log.Printf("Error running command:%s", err)
		return "", err
	}
	if print == true {
		log.Printf("%s", output)
	}

	return output, nil
}

func (vm *VMClient) CheckConnection() bool {
	session, err := vm.SSHClient.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	return true
}

func (vm *VMClient) CheckNodeReadiness(vmName string) (bool, error) {
	output, _, err := vm.RunCommand(fmt.Sprintf("kubectl get nodes %s -o jsonpath='{.status.conditions[?(@.type==\"Ready\")].status}'", vmName))
	if vm.printFlag == true {
		log.Printf("kubectl get nodes output: %s", output)
	}

	if err != nil {
		// Check if the error is due to the node not being found, which is a transient error
		if strings.Contains(err.Error(), "NotFound") {
			// Node not found; treat as not ready rather than an error
			return false, nil
		}
		// For other errors, return them
		return false, fmt.Errorf("error checking node readiness: %v", err)
	}

	return strings.TrimSpace(output) == "True", nil
}

var backoffIntervals = []int{10, 15, 15, 25, 25, 30, 30, 30, 30} // Retry intervals in seconds

func (vm *VMClient) WaitForNodeReadiness(vmName string) error {
	for i, interval := range backoffIntervals {
		ready, err := vm.CheckNodeReadiness(vmName)
		if err != nil {
			// Log the error and continue the loop instead of returning immediately
			utils.LogError(fmt.Sprintf("Error during checkNodeReadiness in Backoff number %d ERROR:%s", i, err))
		} else if ready {
			log.Printf("%s Node %s is Ready", utils.TICK_GREEN, vmName)
			return nil
		}

		if i < len(backoffIntervals)-1 {
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
	return fmt.Errorf("node %s did not reach Ready state within the specified retry intervals", vmName)
}

// Close terminates the SSH connection.
func (vm *VMClient) Close() {
	if vm.SSHClient != nil {
		vm.SSHClient.Close()
	}
}

var retryIntervals = []int{5, 10, 15, 25, 30, 20, 20, 20, 45}

func WaitForPodReadiness(client *VMClient, podName string) error {
	for i, interval := range retryIntervals {
		output, _, err := client.RunCommand("kubectl get pod " + podName + " -o jsonpath='{.status.phase}'")

		// Log the error but don't return immediately. Continue retrying.
		if err != nil {
			// Check if the error is transient (like pod not found)
			if strings.Contains(err.Error(), "NotFound") {
				utils.LogError(fmt.Sprintf("Pod %s not found in Backoff number %d. Retrying...", podName, i))
			} else {
				utils.LogError(fmt.Sprintf("Error during checkPodReadiness in Backoff number %d for pod %s: %s", i, podName, err))
			}
		} else if strings.TrimSpace(output) == "Running" {
			// Pod is running, return nil
			return nil
		}

		// Sleep before retrying if not the last interval
		if i < len(retryIntervals)-1 {
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
	return fmt.Errorf("pod %s did not reach Running state within the specified retry intervals", podName)
}

func (vm *VMClient) SetDebugFlag() {
	vm.printFlag = true
}
