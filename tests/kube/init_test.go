package tests

import (
	"testing"

	kssh "kvmgo/network/ssh"
)

func TestKubeInit(t *testing.T) {
	worker := "worker"

	control := "control"

	wconn, err := kssh.EstablishSsh(control)
	if err != nil {
		t.Errorf("Failed to conn worker Error:%s", err)
	}

	mconn, err := kssh.EstablishSsh(worker)
	if err != nil {
		t.Errorf("Failed to conn control Error:%s", err)
	}

	defer wconn.Close()

	defer mconn.Close()

	out, err := kssh.RunCmd(mconn, "kubectl get nodes")
	if err != nil {
		t.Errorf("failed cmd Error:%s", err)
	}

	t.Log(out)

	t.Log("successfully connected to Kube Nodes")
}
