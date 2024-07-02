package kube

import (
	"fmt"
	"log"

	"kvmgo/lib"
	"kvmgo/utils"
)

type KubeCluster struct {
	nodes   map[string]*KubeClient
	control *KubeClient

	mainVirtClient *lib.VirtClient
}

func (k *KubeCluster) CloseClient() {
	k.mainVirtClient.Close()
}

func (k *KubeCluster) ControlNode() *KubeClient {
	return k.control
}

func (k *KubeCluster) Workers() map[string]*KubeClient {
	return k.nodes
}

// New Kube Cluster
func NewCluster(controlDomain string, nodes []string) (*KubeCluster, error) {
	cluster := KubeCluster{
		nodes: make(map[string]*KubeClient),
	}

	lvirtConn, doms, err := lib.AwaitDomains(nodes)
	if err != nil {
		return nil, err
	}

	cluster.mainVirtClient = lvirtConn

	log.Printf(utils.TurnSuccess("Cluster Nodes are initalized"))

	ctrld, ok := doms[controlDomain]

	if !ok {
		return nil, fmt.Errorf("Control node domain not found")
	}

	cnode, err := NewKubeNodeFromDomain(ctrld, true)
	cluster.control = cnode

	if err != nil {
		return nil, err
	}

	for _, node := range nodes {

		worker, ok := doms[node]
		if !ok {
			return nil, fmt.Errorf("Control node domain not found")
		}

		wnode, err := NewKubeNodeFromDomain(worker, false)
		if err != nil {
			return nil, err
		}

		cluster.nodes[node] = wnode
	}

	return &cluster, nil
}
