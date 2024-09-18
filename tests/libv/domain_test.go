package libv_test

import (
	"fmt"
	"testing"

	"kvmgo/kube"
	"kvmgo/lib"
	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

//  go run main.go --cluster --control=kubecontrol --workers=kubeworker1,kubeworker2

func TestLibvirtAwait(t *testing.T) {
	doms := []string{"control", "worker"}

	cluster, err := kube.NewCluster("control", doms)
	if err != nil {
		t.Errorf("cluster failure %s\n", err)
	}

	// here THe ip is not ready even tho lvirt domain is ready
	control := cluster.ControlNode()

	ip, _ := control.IP()

	t.Log("IP:" + ip)

	f, out, _, err := control.KubeInitalized()
	if err != nil {
		t.Errorf("failure kubeinit %s\n", err)
	}

	t.Logf("found? :%v out: %s\n", f, out)

	//	if _, err := lib.AwaitDomains(doms); err != nil {
	//		t.Errorf("Failed to await doms : error %s\n", err)
	//	}

	t.Log("Successfully awaited domains")
}

func TestLibvirtApi(t *testing.T) {
	lib.ConnectLibvirt()

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		t.Errorf("Error Connecting %s", err)
	}
	defer conn.Close()

	utils.ListAllDomains(conn)

	if err := utils.GetDomainInfo(conn, "control"); err != nil {
		t.Logf("failed get dom info %s\n", err)
	}

	// doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_SHUTOFF | libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_RUNNING | libvirt.CONNECT_LIST_DOMAINS_OTHER)

	e := uint(0xffff)
	enum := libvirt.ConnectListAllDomainsFlags(e)

	doms, err := conn.ListAllDomains(enum)
	if err != nil {
		t.Errorf("Error Listing %s", err)
	}

	fmt.Printf("%d running domains:\n", len(doms))
	for _, dom := range doms {
		name, err := dom.GetName()
		if err == nil {
			fmt.Printf("  %s\n", name)
		}
		dom.Free()
	}
}
