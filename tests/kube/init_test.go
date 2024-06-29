package tests

import (
	"testing"

	kssh "kvmgo/network/ssh"
)

func TestKubeInit(t *testing.T) {
	//	worker := "worker"

	control := "control"

	wclient, wsess, err := kssh.EstablishSsh(control)
	if err != nil {
		t.Errorf("Failed to conn worker Error:%s", err)
	}

	defer wclient.Close()
	defer wsess.Close()

	// Run commands on the worker
	out, err := kssh.RunCmd(wsess, "ls")
	if err != nil {
		t.Errorf("failed cmd Error:%s", err)
	}
	t.Log(out)

	/*
		mclient, msess, err := kssh.EstablishSsh(worker)
		if err != nil {
			t.Errorf("Failed to conn control Error:%s", err)
		}
		defer mclient.Close()

		defer msess.Close()

		out, err = kssh.RunCmd(msess, "ls")
		if err != nil {
			t.Errorf("failed cmd Error:%s", err)
		}

		t.Log(out)

		t.Log("successfully connected to Kube Nodes")
	*/
}
