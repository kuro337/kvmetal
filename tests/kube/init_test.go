package tests

import (
	"strings"
	"testing"

	kssh "kvmgo/network/ssh"
)

func TestKubeInit(t *testing.T) {
	worker := "worker"

	control := "control"

	wclient, err := kssh.EstablishSsh(worker)
	if err != nil {
		t.Errorf("Failed to conn worker Error:%s", err)
	}

	defer wclient.Close()

	// Run commands on the worker
	out, err := kssh.RunCmd(wclient, "ls")
	if err != nil {
		t.Errorf("failed cmd Error:%s", err)
	}
	t.Log(out)

	mclient, err := kssh.EstablishSsh(control)
	if err != nil {
		t.Errorf("Failed to conn control Error:%s", err)
	}
	defer mclient.Close()

	// kubectl get nodes
	out, err = kssh.RunCmd(mclient, "tail -10 kubeadm-init.log")
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		t.Errorf("failed cmd Error:%s", err)
	}

	lines := strings.Split(out, "\n")

	for _, line := range lines {
		t.Logf("line: %s\n", line)
	}

	t.Log(out)

	t.Log("successfully connected to Kube Nodes")
}
