package test

import (
	"log"
	"testing"

	"kvmgo/lib"
	"kvmgo/lib/api"

	"libvirt.org/go/libvirt"
)

func TestWrite(t *testing.T) {
	//	url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"
	url := "https://cloud-images.ubuntu.com/releases/mantic/release/ubuntu-23.10-server-cloudimg-amd64.img"

	imagesDir := "/home/kuro/kvm/images/ubuntu"

	s, err := api.DownloadImageProgress(url, imagesDir)
	// s, err := api.FetchImageUrl(url, "/home/kuro/kvm/test/")
	if err != nil {
		log.Printf("Failed operation Error:%s", err)
	}

	t.Log(s)
}

// Create a VM with a Storage Pool - which stores libvirt managed files such as .img , etc.
func TestVM(t *testing.T) {
	name := "testTemp"
	path := "/home/kuro/testtemp"
	img := "/home/kuro/kvm/images/ubuntu/ubuntu-24.04-server-cloudimg-amd64.img"

	// 1. Create the VM (retrieve/create the Storage Pool)
	vm, err := api.NewVM(name, path)
	if err != nil {
		t.Fatalf("Error creating new VM: %s\n", err)
	}

	t.Logf("VM created:%s path:%s\n", vm.Name, vm.StoragePath)

	// 2. Create the Base Image we need to create a VM
	if err := vm.CreateBaseImage(img, 20); err != nil {
		t.Fatalf("Error creating image :%s", err)
	}

	if err := vm.ListImages(); err != nil {
		t.Errorf("Error listing images :%s", err)
	}

	t.Log("VM Base Image Created")

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
		t.Fatalf("Error:%s", err)
	}

	all, err := api.ListAllStoragePools(conn)
	if err != nil {
		t.Errorf("Failed to list all Error:%s", err)
	}

	for _, p := range all {
		t.Logf("Name:%s Path:%s\n", p.Name, p.Path)
	}
}

// List all Images/Volumes associated with a Pool : go test -v --run TestListImages | fzf
func TestListImages(t *testing.T) {
	name := "testTemp"
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		t.Fatalf("Error Connecting:%s", err)
	}
	vols, err := api.ListAllVolumes(conn, name)
	if err != nil {
		t.Fatalf("Failed to get Volumes Error:%s", err)
	}
	for _, vol := range vols {
		t.Logf("Volume: %s\n", vol.String())
	}
}

// Delete a pool - and clear its volumes
func TestDeletePool(t *testing.T) {
	name := "testTemp"
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		t.Fatalf("Error Connecting:%s", err)
	}

	pool, err := lib.GetPool(conn, name)
	if err != nil {
		t.Fatalf("Failed to get Pool Error:%s", err)
	}

	vols, err := pool.GetVolumes(true)
	if err != nil {
		t.Fatalf("Failed to get Volumes Error:%s", err)
	}

	for _, vol := range vols {
		t.Logf("Volume: %s\n", vol)
	}

	t.Log("Deleting Pool")
	if err := pool.Delete(); err != nil {
		t.Fatalf("Failed to delete Pool Error:%s", err)
	}

	if api.CheckPoolExists(conn, name) {
		t.Error("Pool still exists")
	}
	t.Log("Successfully deleted pool")
}

// base has the Base OS Images
func TestImageApi(t *testing.T) {
	base := "base"

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

	for _, vol := range imgs {
		t.Logf("Volume: %s\n", vol.String())
	}

	// Pool exists

	// get path

	// get volumes
}
