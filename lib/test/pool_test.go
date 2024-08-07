package test

import (
	"log"
	"testing"

	"kvmgo/lib"

	"libvirt.org/go/libvirt"
)

func TestDeletePool(t *testing.T) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Errorf("Error:%s", err)
	}
	defer conn.Close()
	// Create pool "ubuntu" at "/home/kuro/kvm/test"
	poolName := "ubuntu"

	if ex, _ := lib.PoolExists(conn, poolName); ex {
		t.Logf("Pool exists")

		if err := lib.DeletePool(conn, poolName); err != nil {
			t.Fatalf("failed to delete: %s\n", err)
		}

		t.Log("Deleted the active Pool")
	}

	log.Printf("Pool %s successfully deleted\n", poolName)
}

func TestUbuntuPool(t *testing.T) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Printf("Error Connecting %s", err)
		t.Errorf("Error:%s", err)
	}

	defer conn.Close()

	// Create pool "ubuntu" at "/home/kuro/kvm/test"
	poolName := "ubuntu"
	poolPath := "/home/kuro/kvm/test"

	if ex, _ := lib.PoolExists(conn, poolName); ex {
		t.Logf("Pool exists")

		if err := lib.DeletePool(conn, poolName); err != nil {
			t.Fatalf("failed to delete: %s\n", err)
		}

		t.Log("Deleted the active Pool")
	}

	t.Log("Deleted")

	pool, err := lib.NewPool(conn, poolName, poolPath)
	if err != nil {
		t.Fatalf("Failed to create storage pool: %v", err)
	}

	t.Log("created new pool")

	// Add an image with the URL and name "latest"
	url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"
	volumeNameLatest := "latest"

	if err := pool.DeleteImage(volumeNameLatest); err != nil {
		t.Fatalf(err.Error())
	}

	t.Log("Creating Image URL")
	if err := pool.CreateImageURL(volumeNameLatest, url, 10); err != nil {
		t.Fatalf("Failed to add image 'latest' to pool: %v", err)
	}

	t.Log("CREATED Image XML from URL")

	t.Log("getting volume")
	// Get and print the path of "latest"
	latestPath, err := pool.GetVolume(volumeNameLatest)
	if err != nil {
		t.Fatalf("Failed to get the path of 'latest': %v", err)
	}
	t.Logf("Path of 'latest': %s", latestPath)

	// Add another image called "copy"
	volumeNameCopy := "copy"

	t.Log("creating image path")
	if err := pool.CreateImagePath(volumeNameCopy, latestPath, 10); err != nil {
		t.Fatalf("Failed to add image 'copy' to pool: %v", err)
	}

	// Print the path of "copy"
	copyPath, err := pool.GetVolume(volumeNameCopy)
	if err != nil {
		t.Fatalf("Failed to get the path of 'copy': %v", err)
	}
	t.Logf("Path of 'copy': %s", copyPath)

	// create pool "ubuntu" at "/home/kuro/kvm/test"

	// add an Image with the URL and name "latest"
	// 	url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"

	// get and print the path of "latest"

	// then add another Image called "copy" -

	// then - print its path -
}
