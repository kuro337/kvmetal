package presets

import (
	"log"
	"strings"

	"kvmgo/configuration"
	"kvmgo/constants"
)

type Kafka struct {
	domain string
}

func SubstitueAdvertisedListenersKafka(yamlTemplate, domain string) string {
	fqdn := domain + ".kuro.com"
	r := "$FQDN"
	ans := strings.Replace(yamlTemplate, r, fqdn, 1)
	return strings.Replace(ans, "##-", "  -", 1)
}

func (k Kafka) Substitutions(userdata string) string {
	return userdata
	// return SubstitueAdvertisedListenersKafka(userdata, k.domain)

}

/* Launch Kafka */
func CreateKafkaUserData(username, pass, vmname, sshpub string) string {
	config, err := configuration.NewConfigBuilder(
		Kafka{domain: vmname},
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
