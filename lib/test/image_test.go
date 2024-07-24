package test

import (
	"log"
	"testing"

	"kvmgo/lib"

	"libvirt.org/go/libvirt"
)

func TestCreateVM(t *testing.T) {
	vm := lib.NewVMConfig("testvm")

	userdata := "/home/kuro/Documents/Code/Go/kvmgo/data/userdata/default/user-data.img"
	img := "/home/kuro/kvm/images/ubuntu/ubuntu-24.04-server-cloudimg-amd64.img"

	vm.SetMemory(2048).
		SetCores(2).
		SetBaseImage(img).
		SetUserDataPath(userdata).
		SetNetwork("default").
		SetOSVariant("ubuntu18.04")

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Fatalf("Error:%s", err)
	}

	if err := vm.CreateAndStartVM(conn); err != nil {
		t.Fatalf("Error Starting VM:%s", err)
	}
}

/*

Storage Pools


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

/*
Can u give me below command with sampel values and paths in a single line so I can run it from the CLI?

	cmdArgs := []string{
		"--name", s.VMName,
		"--virt-type", "kvm",
		"--memory", fmt.Sprint(s.Memory),
		"--vcpus", fmt.Sprint(s.CPUCores),
		"--disk", "path=" + generatedVmImg + ",device=disk",
		"--disk", "path=" + vm_userdata_img + ",format=raw",
		"--graphics", "none",
		"--boot", "hd,menu=on",
		"--network", "network=default",
		"--os-variant", "ubuntu18.04",
		"--noautoconsole",

Certainly! Here is the command with sample values and paths in a single line:

virt-install --name sampleVM --virt-type kvm --memory 2048 --vcpus 2 --disk path=/home/kuro/kvm/images/ubuntu/base/ubuntu-24.04-server-cloudimg-amd64.img,device=disk --disk path=tests/user-data.img,format=raw --graphics none --boot hd,menu=on --network network=default --os-variant ubuntu18.04 --noautoconsole

	   data/artifacts/worker/userdata/user-data.img

	   tests/user-data.img
	/home/kuro/kvm/images/ubuntu/base/ubuntu-24.04-server-cloudimg-amd64.img

imgManager, err := lib.NewImageMgr("ubuntu", "")

	if err != nil {
		t.Logf("failed to create imgMgr image, %s\n", err)
	}
*/

func TestAPI(t *testing.T) {
	base := "ubuntu"
	vm := "kube"
	// getImage(base)
	log.Println(base, vm)

	// getBaseImage()

	// createVM for kube at path

	// specify Kube path

	// check if kube img already exists

	// create the Pool with name of VM
}

func TestFullKvmImageMgmt(t *testing.T) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Errorf("Error:%s", err)
	}

	// It is saved globally , anytime created
	if _, err := conn.LookupStoragePoolByName("bubuntu"); err == nil {
		t.Fatalf("Nonexistent never created Storage Pool should error:%s", err)
	}

	if _, err = conn.LookupStoragePoolByName("ubuntu"); err != nil {
		t.Fatalf("Once created the storage pools should be persistent Error:%s", err)
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

	//
	if err := imgManager.CreateImageFromBase(name, "kvm", 10); err != nil {
		t.Logf("failed to Get image, %s\n", err)
	}

	if err := imgManager.DeleteImgVolume("kvm"); err != nil {
		t.Logf("failed to Delete image, %s\n", err)
	}
}
