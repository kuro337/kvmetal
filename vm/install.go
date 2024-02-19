package vm

type VirtInstall struct {
	name          string
	memory        int
	vcpu          int
	image         string
	userdata      string
	noautoconsole bool
}

func (v *VirtInstall) install() {
	/*
				cmd := exec.Command("virt-install", "--name", s.VMName, "--virt-type", "kvm", "--memory", fmt.Sprint(s.Memory), "--vcpus", fmt.Sprint(s.CPUCores), "--boot", "hd,menu=on", "--disk", "path="+modifiedImagePath+",device=disk", "--disk", "path="+userDataImgPath+",format=raw", "--graphics", "none", "--os-type", "Linux", "--os-variant", "ubuntu18.04", "--noautoconsole")

				/usr/bin/virt-install \
		--name kubecontrol \
		--virt-type kvm \
		--memory 2048 \
		--vcpus 2 \
		--boot hd,menu=on \
		--disk path=data/images/kubecontrol-vm-disk.qcow2,device=disk \
		--disk path=data/userdata/default/user-data.img,format=raw \
		--graphics none \
		--os-type Linux --os-variant ubuntu18.04 --noautoconsole


	*/
}

// func InstallVM(conn *libvirt.Connect) {
// 	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
// 	if err != nil {
// 		log.Printf("Error Listing %s", err)
// 	}

// }

/*
func (s *VMConfig) CreateVM() error {
	s.navigateToRoot()

	modifiedImagePath := filepath.Join(s.ImagesDir, s.VMName+"-vm-disk.qcow2")
	vm_userdata_img := filepath.Join("data", "artifacts", s.VMName, "userdata", "user-data.img")

	cmd := exec.Command("virt-install", "--name", s.VMName,
		"--virt-type", "kvm",
		"--memory", fmt.Sprint(s.Memory),
		"--vcpus", fmt.Sprint(s.CPUCores),
		"--disk", "path="+modifiedImagePath+",device=disk",
		"--disk", "path="+vm_userdata_img+",format=raw",
		"--graphics", "none",
		"--boot", "hd,menu=on",
		"--network", "network=default",
		"--os-variant", "ubuntu18.04", "--noautoconsole")

	log.Printf("%sCreating Virtual Machine%s %s%s%s%s: %s\n", utils.BOLD, utils.NC, utils.BOLD, utils.COOLBLUE, s.VMName, utils.NC, cmd.String())

	var stderr bytes.Buffer // Capture stderr
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("ERROR Failed to Create VM error=%q", stderr.String())
		return err
	}

	return nil
}


*/
