package test

import (
	"testing"

	"kvmgo/lib"
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
*/

func TestFullKvmImageMgmt(t *testing.T) {
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

	/*

	   	// Pull image - to /var/lib/libvirt/images/base - we store base images there
	   	if err := lib.PullImage(baseImgUrl, poolPath); err != nil {
	   		t.Logf("failed to pull image, %s\n", err)
	   	}

	   	t.Log("successfully pulled image") // or already exists

	   	poolName := "default"
	   	mgrName := "default"

	   	baseImgName := "ubuntu-24.04-server-cloudimg-amd64.img"

	   	// Manager - "default"
	   	client, err := lib.NewImageMgr(mgrName, poolPath)

	   	// "default" storage pool
	   	if err := client.CreateStoragePool(poolName, poolPath); err != nil {
	   		t.Fatalf("Failed to create storage pool: %s", err)
	   	}

	   	// imgPath := poolPath + baseImgName // create our image

	       vmName := "kvm"

	   	if err := client.CreateImgVolume(poolName, vmName, 10); err != nil {
	   		t.Fatalf("Error checking if image exists: %s", err)
	   	}

	   	// NEED an imageExists method
	   	// imageExists , err := client.ImageExists(baseImg)

	   	if err != nil {
	   		t.Errorf("Error:%s", err)
	   	}

	*/

	// kvmImg := "worker"

	// client.CreateNewImage(kvmImg, 20)

	// imgExists := client.ImageExists(kvmImg) ???

	// make sure image for
}
