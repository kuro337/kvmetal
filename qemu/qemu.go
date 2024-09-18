package qemu

import "fmt"

type Qemu struct{}

// Qemu Flags
// qemu-img create`: This is the command to create a new disk image.
// -b backing_img` Specifies the backing file.
// The new image will be created as a copy-on-write (COW) image based on this backing file.
// -F qcow2`: This option specifies the format of the backing file. In this case, the backing file format is `qcow2`, which stands for QEMU Copy-On-Write version 2.
// -f qcow2`: This option specifies the format of the new image file. Here, the new image will also be in the `qcow2` format.
// qemu-img create -b <backing_file> -F <backing_format> -f output_format <output_name>

// CowImgFromBackingImg creates a new Image from a Base Backing Image OS File
func CowImgFromBackingImg(baseOsImg, outputImgName string) string {
	return fmt.Sprintf("qemu-img create -b %s -F qcow2 -f qcow2 %s 20G",
		baseOsImg, outputImgName)
}

func (q *Qemu) CreateImage() {
}
