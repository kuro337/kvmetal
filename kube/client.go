package kube

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
	"time"

	// "kvmgo/lib"
	dom "kvmgo/lib/domain"
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

	Nodes        map[string]KubectlNodeResp
	RunningNodes []KubectlNodeResp
	Children     []string
	role         KubeNode

	dom *dom.Domain
}

func NewKubeNodeFromDomain(domain *dom.Domain, control bool) (*KubeClient, error) {
	sshClient, err := domain.NewSSHClient()
	if err != nil {
		return nil, err
	}

	role := Control

	if !control {
		role = Worker
	}

	return &KubeClient{
		domain: domain.Name, client: sshClient, ip: sshClient.IP, role: role,
		Children: []string{},
		Nodes:    map[string]KubectlNodeResp{},
	}, nil
}

func NewControl(domain string) (*KubeClient, error) {
	client, err := network.GetSSHClient(domain)
	if err != nil {
		return nil, fmt.Errorf("Failed to conn control Error:%s", err)
	}

	return &KubeClient{
		domain: domain, client: client, ip: client.IP, role: Control,
		Children: []string{},
		Nodes:    map[string]KubectlNodeResp{},
	}, nil
}

func NewWorker(domain string) (*KubeClient, error) {
	client, err := network.GetSSHClient(domain)
	if err != nil {
		return nil, fmt.Errorf("Failed to conn control Error:%s", err)
	}

	return &KubeClient{domain: domain, client: client, ip: client.IP, role: Worker}, nil
}

func (c *KubeClient) IP() (string, error) {
	log.Printf("IP of CLient: %s\n", c.ip)
	return c.ip, nil
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

// CheckNodes returns the current nodes active on the Cluster
func (c *KubeClient) KubeInitalized() (bool, string, string, error) {
	// 2 4 8 16 32 64 128 256 512

	retries := 8
	delay := 5
	i := 0

	var ferr error
	errc := ""
	find := "kubeadm-init.log"

	for i < retries {

		out, serr, err := c.client.RunCmd("ls")

		if err == nil && strings.Contains(out, find) {
			return true, out, serr, nil
		}

		wait := delay + (1 << i)

		log.Printf("Failed seeking file attempt #%d , waiting %d seconds. %s", i, wait, err)

		time.Sleep(time.Duration(wait) * time.Second)
		i++
		ferr = err
		errc = serr
	}

	return false, "", errc, ferr
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

type KubectlNodeResp struct {
	Name    string
	Status  string
	Roles   string
	Age     string
	Version string
}

func (c *KubeClient) CheckNodesN() ([]KubectlNodeResp, error) {
	ctlnodes, serr, err := c.CheckNodes()
	if err != nil {
		return nil, fmt.Errorf("Error:%s", err)
	}

	var kgn []KubectlNodeResp

	// Name Status Roles Age Version
	log.Printf("kubectl resp: %s , serr: %s\n", ctlnodes, serr)

	re := regexp.MustCompile(`\s+`)

	lines := strings.Split(ctlnodes, "\n")

	for _, line := range lines[1:] {

		if line == "" {
			continue
		}

		result := re.ReplaceAllString(line, " ")

		cols := strings.Split(result, " ")

		if len(cols) == 0 {
			continue
		}

		if len(cols) >= 1 {
			if _, ok := c.Nodes[cols[0]]; ok {
				continue
			}
		}

		resp := KubectlNodeResp{}

		for i, col := range cols {
			switch i {
			case 0:
				if col != c.domain {
					c.Children = append(c.Children, col)
				}
				resp.Name = col
			case 1:
				resp.Status = col
			case 2:
				resp.Roles = col
			case 3:
				resp.Age = col
			case 4:
				resp.Version = col
			}
		}

		if len(cols) >= 2 && cols[0] != c.domain {
			log.Printf("Node:%s,Status:%s\n", cols[0], cols[1])
			c.RunningNodes = append(c.RunningNodes, resp)
			c.Nodes[resp.Name] = resp
		}

		kgn = append(kgn, resp)
	}

	slices.SortStableFunc(kgn, func(a KubectlNodeResp, b KubectlNodeResp) int {
		if a.Name == b.Name {
			return 0
		} else if a.Name > b.Name {
			return 1
		}
		return -1
	})

	return kgn, nil
}
