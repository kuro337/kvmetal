package presets

import (
	"log"

	"kvmgo/configuration"
	"kvmgo/constants"
)

func CreateKubeControlPlaneUserData(username, pass, vmname string) string {
	config, err := configuration.NewConfigBuilder(
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Zsh,
			constants.KubernetesControl,
		},
		[]constants.CloudInitPkg{
			constants.Containerd,
			constants.TransportHttps,
			constants.ZSH,
			constants.Curl,
		},
		[]constants.InitSvc{
			constants.Restart,
		},
		username, pass, vmname)
	if err != nil {
		log.Printf("Failed to create Configuration")
	}

	userdata := config.CreateCloudInitData()
	return userdata
}

func CreateKubeWorkerUserData(username, pass, vmname string) string {
	config, err := configuration.NewConfigBuilder(
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Zsh,
			constants.KubeWorker,
		},
		[]constants.CloudInitPkg{
			constants.Containerd,
			constants.TransportHttps,
			constants.ZSH,
			constants.Curl,
		},
		[]constants.InitSvc{
			constants.Restart,
		},
		username, pass, vmname)
	if err != nil {
		log.Printf("Failed to create Configuration")
	}

	userdata := config.CreateCloudInitData()
	return userdata
}
