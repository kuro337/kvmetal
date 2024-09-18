package presets

import (
	"fmt"
	"log"
	"strings"

	"kvmgo/configuration"
	"kvmgo/constants"
	"kvmgo/constants/redpanda"
	"kvmgo/utils"
)

type Redpanda struct {
	domain string
}

func (k Redpanda) Substitutions(userdata string) string {
	return userdata
	// return SubstitueAdvertisedListenersKafka(userdata, k.domain)
}

/* Test using Preset now - this should generate full metadata */
func CreateRedpandaUserdata(
	username, pass, vmname, sshpub,
	vmIP, vmPort, hostIP, hostPort string,
) string {
	config, err := configuration.NewConfigBuilder(
		Redpanda{domain: vmname},
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Zsh,
		},
		[]constants.CloudInitPkg{
			constants.ZSH,
			constants.Git,
			constants.Curl,
			constants.Wget,
			constants.Tar,
		},
		[]constants.InitSvc{
			constants.Restart,
		},
		username, pass, vmname, sshpub)
	if err != nil {
		log.Printf("Failed to create Configuration")
	}

	userdata := config.CreateCloudInitData()

	return userdata + GenerateRedpandaUserdata(vmIP, vmPort, hostIP, hostPort)
}

func GenerateRedpandaUserdata(vmIP, vmPort, hostIP, hostPort string) string {
	initCmds := utils.IndentArrayRunCmd(redpanda.REDPANDA_RUNCMD_INITIAL_STEPS)

	config := GenerateRedpandaConfig(vmIP, vmPort, hostIP, hostPort)

	substitued := GetInitSettingsReplacementString(config)

	runCmdReplaceIndented := IndentRpSettingsReplaceCmds(substitued)

	return initCmds + "\n" +
		runCmdReplaceIndented +
		fmt.Sprintf("  - %s\n\n", redpanda.REDPANDA_START_CMD) +
		`final_message: "Redpanda has been successfully installed and started."` + "\n"
}

func IndentRpSettingsReplaceCmds(replaceRpSettingsCmd string) string {
	var runCmdSettingsBuilder strings.Builder
	runCmdSettingsBuilder.WriteString("  - |-\n")

	settingsLines := strings.Split(replaceRpSettingsCmd, "\n")

	for _, line := range settingsLines {
		runCmdSettingsBuilder.WriteString(fmt.Sprintf("    %s\n", line))
	}

	return runCmdSettingsBuilder.String()
}

func RedpandaAdvertisedExternalKafkaAPI(hostIP, hostPort string) string {
	return fmt.Sprintf(`
- name: external
  address: %s
  port: %s
`, hostIP, hostPort)
}

func RedpandaAdvertisedInternalKafkaAPI(vmIP string) string {
	return fmt.Sprintf(`
- name: internal
  address: %s
  port: 9092
`, vmIP)
}

func RedpandaExternalKafkaAPI(vmPort string) string {
	return fmt.Sprintf(`
- name: external
  address: 0.0.0.0
  port: %s
`, vmPort)
}

func GetInitSettingsReplacementString(evaluatedSettings string) string {
	return fmt.Sprintf(redpanda.REDPANDA_SETTINGS_RUNCMD_TEMPLATE, evaluatedSettings)
}

/*
Create Redpanda Config to Launch and Expose

	config := GenerateRedpandaConfig(vmIP, vmPort, hostIP, hostPort)
	fmt.Println(config)
*/
func GenerateRedpandaConfig(vmIP, vmPort, hostIP, hostPort string) string {
	config := `
redpanda:	
    data_directory: /var/lib/redpanda/data
    seed_servers: []
    rpc_server:
        address: 0.0.0.0
        port: 33145
    kafka_api:
        - name: internal
          address: 0.0.0.0
          port: 9092
        - name: external
          address: 0.0.0.0
          port: %s
    admin:
        - address: 0.0.0.0
          port: 9644
    advertised_rpc_api:
        address: 127.0.0.1
        port: 33145
    advertised_kafka_api:
        - name: internal
          address: %s
          port: 9092
        - name: external
          address: %s
          port: %s
rpk:
    tune_network: true
    tune_disk_scheduler: true
    tune_disk_nomerges: true
    tune_disk_write_cache: true
    tune_disk_irq: true
    tune_cpu: true
    tune_aio_events: true
    tune_clocksource: true
    tune_swappiness: true
    coredump_dir: /var/lib/redpanda/coredump
    tune_ballast_file: true

pandaproxy: {}
schema_registry: {}
`

	config = fmt.Sprintf(config, vmPort, vmIP, hostIP, hostPort)
	return config
}
