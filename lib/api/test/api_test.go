package test

import (
	"log"
	"strings"
	"testing"

	"kvmgo/lib/api"

	"libvirt.org/go/libvirt"
)

func TestWrite(t *testing.T) {
	url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"

	s, err := api.FetchImageUrl(url, "/home/kuro/kvm/test/")
	if err != nil {
		log.Printf("Failed operation Error:%s", err)
	}

	t.Log(s)
}

func TestVM(t *testing.T) {
	name := "testTemp"
	path := "/home/kuro/testtemp"
	vm, err := api.NewVM(name, path)
	if err != nil {
		t.Errorf("Error new VM: %s\n", err)
	}

	t.Logf("VM:\n%+v\n", vm)

	// url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"

	//tmp, err := vm.AddImageHttp(url, "someVM")
	//if err != nil {
	//	t.Errorf("Error new VM: %s\n", err)
	//}
	//t.Logf("Generated : %s\n", tmp)
}

// List all Storage Pools : go test -v --run TestListAll | fzf
func TestListAll(t *testing.T) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Errorf("Error:%s", err)
	}

	all, err := api.ListAllStoragePools(conn)
	if err != nil {
		t.Errorf("Failed to list all Error:%s", err)
	}

	for _, p := range all {
		t.Logf("Name:%s Path:%s\n", p.Name, p.Path)
	}
}

func TestImageApi(t *testing.T) {
	base := "images"

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Errorf("Error:%s", err)
	}

	if !api.CheckPoolExists(conn, base) {
		t.Fatal("Pool does not exist")
	}

	path, err := api.GetPoolPath(conn, base)
	if err != nil {
		t.Fatalf("Failed to get Path Error:%s", err)
	}

	t.Logf("Pool Path: %s\n", path)

	imgs, err := api.ListPoolVolumes(conn, base)
	if err != nil {
		t.Fatalf("Failed to get Path Error:%s", err)
	}

	if len(imgs) == 0 {
		t.Error("No Images Present")
	}

	t.Logf("Volumes: %s\n", strings.Join(imgs, "\n"))

	// Pool exists

	// get path

	// get volumes
}
