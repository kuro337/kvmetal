package ubuntu

import (
	"log"

	"kvmgo/constants"
	"kvmgo/constants/bigdata"
	"kvmgo/constants/shell"
)

type UbuntuConfig struct{}

func (u *UbuntuConfig) DefaultCloudInit() string {
	return constants.CloudInitUbuntu
}

func (u *UbuntuConfig) GetImageUrl() string {
	return "https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img"
}

func (u *UbuntuConfig) GetVersion() string {
	return "22.04_Jammy_amd64"
}

func (u *UbuntuConfig) GetPackage(dep constants.CloudInitPkg) string {
	switch dep {
	case constants.OpenJDK11:
		return string(constants.OpenJDK11)
	case constants.ZSH:
		return string(constants.ZSH)
	case constants.Git:
		return string(constants.Git)
	case constants.Curl:
		return string(constants.Curl)
	default:
		log.Printf("No Default Package Found")
		return ""
	}
}

func (u *UbuntuConfig) GetRunCmd(dep constants.Dependency) string {
	switch dep {
	case constants.Zsh:
		return shell.ZSH_UBUNTU_RUNCMD
	case constants.Hadoop:
		return bigdata.HADOOP_UBUNTU_RUNCMD
	case constants.Spark:
		return bigdata.SPARK_UBUNTU_RUNCMD
	default:
		log.Printf("No Run Command found for Dependency")
		return ""
	}
}

func (u *UbuntuConfig) GetInitSvc(dep constants.InitSvc) string {
	switch dep {
	case constants.Restart:
		return constants.RebootCloudInit
	default:
		log.Printf("No Init Svc found for Cloud init Svc")
		return ""
	}
}
