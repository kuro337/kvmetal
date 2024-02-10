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
