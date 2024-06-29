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

	t.Log("successfully connected to Kube Nodes")

	// kubectl get nodes
	out, err = kssh.RunCmd(mclient, "tail -10 kubeadm-init.log")
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		t.Errorf("failed cmd Error:%s", err)
	}

	lines := strings.Split(out, "\n")

	var join strings.Builder

	f := false
	for _, line := range lines {

		l := strings.TrimSpace(line)
		if f == true {
			join.WriteString(l)
			t.Log("breaking")
			break
		}
		if strings.Contains(l, "kubeadm") {
			f = true
			join.WriteString(strings.TrimSuffix(l, "\\"))
			join.WriteRune(' ')
		}

		t.Logf("line: %s\n", line)
	}

	t.Log(out)

	joinCmd := join.String()
	t.Logf("Join Command: %s\n", joinCmd)

	out, err = kssh.RunCmd(wclient, joinCmd)
	//	out, err = kssh.RunCmd(mclient, "ls")
	if err != nil {
		t.Errorf("failed joining Error:%s", err)
	}

	t.Logf("Worker join output:\n%s\n", out)
}
