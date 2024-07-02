package kube

import (
	"fmt"
	"strings"

	"kvmgo/network"
)

type KubeNode int

const (
	Control KubeNode = 1 << iota
	Worker
)

type KubeClient struct {
	domain string
	ip     string
	client *network.VMClient

	role KubeNode
}

func NewControl(domain string) (*KubeClient, error) {
	client, err := network.GetSSHClient(domain)
	if err != nil {
		return nil, fmt.Errorf("Failed to conn control Error:%s", err)
	}

	return &KubeClient{domain: domain, client: client, ip: client.IP, role: Control}, nil
}

func NewWorker(domain string) (*KubeClient, error) {
	client, err := network.GetSSHClient(domain)
	if err != nil {
		return nil, fmt.Errorf("Failed to conn control Error:%s", err)
	}

	return &KubeClient{domain: domain, client: client, ip: client.IP, role: Worker}, nil
}

// CheckNodes returns the current nodes active on the Cluster
func (c *KubeClient) CheckNodes() (string, string, error) {
	// kubectl get nodes
	out, serr, err := c.client.RunCmd("kubectl get nodes")
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		return out, serr, fmt.Errorf("failed cmd Error:%s", err)
	}

	return out, serr, nil
}

// GetJoinCmd gets the kubeadm Cluster join command for workers
func (c *KubeClient) GetJoinCmd() (string, string, error) {
	// kubectl get nodes
	out, serr, err := c.client.RunCmd("tail -10 kubeadm-init.log")
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		return "", serr, fmt.Errorf("failed cmd Error:%s", err)
	}

	lines := strings.Split(out, "\n")

	var join strings.Builder
	join.WriteString("sudo ")
	f := false
	for _, line := range lines {

		l := strings.TrimSpace(line)
		if f == true {
			join.WriteString(l)
			break
		}
		if strings.Contains(l, "kubeadm") {
			f = true
			join.WriteString(strings.TrimSuffix(l, "\\"))
			join.WriteRune(' ')
		}

	}

	return join.String(), serr, nil
}

// RunJoinCmd runs the kubeadm join on the worker and returns sout & err if any
func (c *KubeClient) RunJoinCmd(joinCmd string) (string, error) {
	out, _, err := c.client.RunCmd(joinCmd)
	if err != nil {
		return "", fmt.Errorf("failed cmd Error:%s", err)
	}

	return out, nil
}
