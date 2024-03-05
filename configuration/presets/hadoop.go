package presets

import (
	"log"

	"kvmgo/configuration"
	"kvmgo/constants"
)

func CreateHadoopUserData(username, pass, vmname, sshpub string) string {
	config, err := configuration.NewConfigBuilder(
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Zsh,
			constants.Hadoop,
		},
		[]constants.CloudInitPkg{
			constants.OpenJDK11,
			constants.ZSH,
			constants.Git,
			constants.Curl,
		},
		[]constants.InitSvc{
			constants.Restart,
		},
		username, pass, vmname, sshpub)
	if err != nil {
		log.Printf("Failed to create Configuration")
	}

	userdata := config.CreateCloudInitData()
	return userdata
}
