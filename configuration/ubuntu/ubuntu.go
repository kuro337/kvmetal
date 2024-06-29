package ubuntu

import (
	"log"

	"kvmgo/constants"
	"kvmgo/constants/bigdata"
	"kvmgo/constants/db"
	"kvmgo/constants/jvm"
	"kvmgo/constants/kafka"
	"kvmgo/constants/kube"
	"kvmgo/constants/shell"
)

type UbuntuConfig struct{}

func (u *UbuntuConfig) DefaultCloudInit() string {
	return constants.DefaultUserdata
}

func (u *UbuntuConfig) GetImageUrl() string {
	return "https://cloud-images.ubuntu.com/releases/jammy/release/ubuntu-22.04-server-cloudimg-amd64.img"
}

func (u *UbuntuConfig) GetVersion() string {
	return "22.04_Jammy_amd64"
}

func (u *UbuntuConfig) GetPackage(dep constants.CloudInitPkg) string {
	return string(dep)
}

func (u *UbuntuConfig) GetRunCmd(dep constants.Dependency) string {
	switch dep {
	case constants.Zsh:
		return shell.ZSH_UBUNTU_RUNCMD
	case constants.JDK_SCALA:
		return jvm.JDK_SCALA_RUNCMD
	case constants.Kafka:
		return kafka.KAFKA_KRAFT_RUNCMD
	case constants.Hadoop:
		return bigdata.HADOOP_UBUNTU_RUNCMD
	case constants.Spark:
		return bigdata.SPARK_UBUNTU_RUNCMD
	case constants.KubernetesControlCalico:
		return kube.KUBE_CONTROL_CALICO_UBUNTU_RUNCMD
	case constants.KubernetesControlCilium:
		return kube.KUBE_CONTROL_CILIUM_UBUNTU_RUNCMD
	case constants.KubeWorker:
		return kube.KUBE_WORKER_UBUNTU_RUNCMD
	case constants.Calico:
		return kube.CALICO_LINUX_RUNCMD
	case constants.Cilium:
		return kube.CILIUM_LINUX_RUNCMD
	case constants.Clickhouse:
		return db.CLICKHOUSE_RUNCMD
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
