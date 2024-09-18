package main

import (
	"log"

	"kvmgo/lib/api"
)

func main() {
	name := "testTemp"
	path := "/home/kuro/testtemp"
	vm, err := api.NewVM(name, path)
	if err != nil {
		log.Fatalf("Error new VM: %s\n", err)
	}

	url := "https://cloud-images.ubuntu.com/releases/noble/release/ubuntu-24.04-server-cloudimg-amd64.img"

	tmp, err := vm.AddImageHttp(url, "someVM")
	if err != nil {
		log.Printf("Error new VM: %s\n", err)
	}

	log.Printf("Generated : %s\n", tmp)
}
