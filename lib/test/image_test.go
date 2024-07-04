package test

import (
	"log"
	"testing"

	"kvmgo/lib"

	"libvirt.org/go/libvirt"
)

/*
func TestBaseImagePull(t *testing.T) {
	baseImgUrl := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"
	baseImgDir := "/var/lib/libvirt/images/base"

	if err := lib.PullImage(baseImgUrl, baseImgDir); err != nil {
		t.Logf("failed to pull image, %s\n", err)
	}

	t.Log("successfully pulled image")

}

	pool, err := im.client.conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return err
	}




*/

func TestFullKvmImageMgmt(t *testing.T) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Errorf("Error:%s", err)
	}

	_, err = conn.LookupStoragePoolByName("ubuntu")
	if err != nil {
		t.Fatalf("Error:%s", err)
	}

	imgManager, err := lib.NewImageMgr("ubuntu", "")
	if err != nil {
		t.Logf("failed to create imgMgr image, %s\n", err)
	}

	url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"
	name := "ubuntu-24.04-server-cloudimg-amd64.img"

	t.Log("returned imgManager - adding Image")

	if imgManager == nil {
		t.Fatal("imgManager is nil")
	}

	if err := imgManager.AddImage(url, name); err != nil {
		t.Logf("failed to pull image, %s\n", err)
	}

	t.Log("added image")

	t.Log("getting image")
	if _, err := imgManager.GetImage(name); err != nil {
		t.Logf("failed to Get image, %s\n", err)
	}

	t.Log("got image")

	if err := imgManager.CreateImageFromBase(name, "kvm", 10); err != nil {
		t.Logf("failed to Get image, %s\n", err)
	}
}
