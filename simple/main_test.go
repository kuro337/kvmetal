package main_test

import (
	"path/filepath"
	"testing"
)

func TestSomething(t *testing.T) {
	url := "https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img"

	dir := "/home/kuro/Documents/Code/Go/kvmgo/simple"

	imageName := filepath.Base(url)
	imagePath := filepath.Join(dir, imageName)
	// pullImgsStr := fmt.Sprintf("Pulling Base Image: URL:%s, Dir:%s, ImgPath: %s\n", url, dir, imagePath)

	// /home/kuro/Documents/Code/Go/kvmgo/simple/ubuntu-22.04-server-cloudimg-amd64.img
	t.Log(imagePath)
}
