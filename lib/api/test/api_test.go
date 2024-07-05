package test

import (
	"log"
	"strings"
	"testing"

	"kvmgo/lib/api"

	"libvirt.org/go/libvirt"
)

func TestImageApi(t *testing.T) {
	base := "ubuntu"

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

	imgs, err := api.ListPoolVolumes(conn, path)
	if err != nil {
		t.Fatalf("Failed to get Path Error:%s", err)
	}

	if imgs == nil || len(imgs) == 0 {
		t.Error("No Images Present")
	}

	t.Logf("Volumes: %s\n", strings.Join(imgs, "\n"))

	// Pool exists

	// get path

	// get volumes
}
