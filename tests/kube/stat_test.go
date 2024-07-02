package tests

import (
	"testing"

	"kvmgo/kube"
)

func TestNodesCtrl(t *testing.T) {
	controlDomain := "control"
	control, err := kube.NewControl(controlDomain)
	if err != nil {
		t.Errorf("Error:%s", err)
	}

	ctlnodes, serr, err := control.CheckNodes()

	t.Logf("kubectl resp: %s , serr: %s\n", ctlnodes, serr)
}
