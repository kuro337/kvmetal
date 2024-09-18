package presets

import (
	"log"

	"kvmgo/configuration"
	"kvmgo/constants"
)

func CreateClickhouseUserData(username, pass, vmname, sshpub string) string {
	config, err := configuration.NewConfigBuilder(
		configuration.DefaultPreset{},
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Clickhouse,
		},
		[]constants.CloudInitPkg{
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
