package presets

import (
	"log"

	"kvmgo/configuration"
	"kvmgo/constants"
)

/* Launch Kafka */
func CreateKafkaUserData(username, pass, vmname, sshpub string) string {
	config, err := configuration.NewConfigBuilder(
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Zsh,
			constants.JDK_SCALA,
			constants.Kafka,
		},
		[]constants.CloudInitPkg{
			constants.ZSH,
			constants.OpenJDK11,
			constants.DefaultJre,
			constants.Tar,
			constants.Wget,
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
