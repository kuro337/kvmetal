package presets

import (
	"fmt"
	"log"
	"strings"

	"kvmgo/configuration"
	"kvmgo/constants"
	"kvmgo/constants/kafka"
	"kvmgo/utils"
)

type Kafka struct {
	domain                    string
	netBufferSizeBytes        int
	netMaxReqSizeBytes        int
	netThreads                int
	ioThreads                 int
	replicationFactor         int
	partitionsPerTopicDefault int
	recoveryThreads           int
	logRetentionHours         int
	maxLogsegmentSize         int
	balanceInterval           int

	logDirs []string
}

func NewKafkaConfig(domain string) Kafka {
	return Kafka{
		netBufferSizeBytes:        102400,
		netMaxReqSizeBytes:        104857600,
		netThreads:                3,
		ioThreads:                 8,
		replicationFactor:         1,
		partitionsPerTopicDefault: 1,
		recoveryThreads:           1,
		logRetentionHours:         168,
		maxLogsegmentSize:         1073741824,
		balanceInterval:           300000,
		logDirs:                   []string{"/tmp/kraft-combined-logs"},
	}
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

func CreateKafkaKraftCluster(username, pass, vmname, sshpub string,
	vmPort int, hostIP string, hostPort int, externalIP string,
	nodeId int, role kafka.KafkaRole,
) string {
	kafkaCfg := NewKafkaConfig(vmname)

	config, err := configuration.NewConfigBuilder(
		kafkaCfg,
		constants.Ubuntu,
		[]constants.Dependency{
			constants.Zsh,
			constants.JDK_SCALA,
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

	kraftUserdata := kafkaCfg.GenerateKraftUserdata(
		vmname,
		vmname+".kuro.com",
		vmPort,
		hostIP,
		hostPort,
		externalIP,
		nodeId,
		kafka.BrokerController)

	return userdata + kraftUserdata
}

/*
DOMAIN=kafkavm
VM_PORT=9095
HOST_PUBPORT=9094
HOST_PUBIP=192.168.1.10
VM_IP=192.168.122.20 # or kafka.kuro.com if we know host can resolve
EXT_IP=192.168.1.225
*/

func (k Kafka) GenerateKraftUserdata(domain,
	vmIPorDomain string, vmPort int,
	hostIP string, hostPort int, externalIP string,
	nodeId int, role kafka.KafkaRole,
) string {
	exposeCmd := ExposeBrokerCmd(domain, vmPort, hostPort, externalIP)
	log.Printf("Expose Command once Kafka Cluster is Running:\n%s\n", exposeCmd)

	kraftStorageFormatCmd := KafkaFormatKraftStorage()

	initCmds := utils.IndentArrayRunCmd(append(
		kafka.KAFKA_RUNCMD_INITIAL_STEPS,
		kraftStorageFormatCmd))

	settings := k.GenerateKafkaSettings(
		domain,
		vmIPorDomain,
		vmPort,
		hostIP,
		hostPort,
		externalIP, nodeId,
		kafka.BrokerController,
	)
	fmt.Println(settings)

	replaceCmd := ReplaceKafkaKraftSettings(settings)

	runCmdReplaceIndented := DefineKafkaSettingsInRunCmd(replaceCmd)

	return initCmds + "\n" +
		runCmdReplaceIndented +
		fmt.Sprintf("  - %s\n\n", kafka.KAFKA_KRAFT_START_CLUSTER) +
		`final_message: "Kafka has been successfully installed and started."` + "\n"
}

func ReplaceKafkaKraftSettings(clusterSettings string) string {
	return fmt.Sprintf(kafka.KAFKA_SETTINGS_RUNCMD_TEMPLATE, clusterSettings)
}

func DefineKafkaSettingsInRunCmd(replaceKraftSettingsCmd string) string {
	var runCmdSettingsBuilder strings.Builder
	runCmdSettingsBuilder.WriteString("  - |-\n")

	settingsLines := strings.Split(replaceKraftSettingsCmd, "\n")

	for _, line := range settingsLines {
		runCmdSettingsBuilder.WriteString(fmt.Sprintf("    %s\n", line))
	}

	return runCmdSettingsBuilder.String()

	// sudo tee /opt/kafka/config/kraft/server.properties > /dev/null <<EOL
}

func (k Kafka) GenerateKafkaSettings(
	domain,
	vmIPorDomain string, vmPort int,
	hostIP string, hostPort int,
	externalIP string,
	nodeId int,
	role kafka.KafkaRole,
) string {
	var kafkaUserdata strings.Builder

	kafkaUserdata.WriteString(
		KafkaRoleSetting(role) + "\n")

	kafkaUserdata.WriteString(
		KafkaNodeIdSetting(nodeId) + "\n\n")

	kafkaUserdata.WriteString(
		kafka.KAFKA_CONTROLLER_QUORUM + "\n\n")

	/* Listener Config */
	kafkaUserdata.WriteString(
		KafkaListenerSettings(fmt.Sprintf("%d", vmPort)) + "\n")

	kafkaUserdata.WriteString(
		KafkaAdvertisedListeners(
			vmIPorDomain, hostIP, fmt.Sprintf("%d", hostPort)) + "\n")

	kafkaUserdata.WriteString(kafka.KAFKA_LISTENER_SETTINGS + "\n")

	/* Compute Config */
	kafkaUserdata.WriteString(
		KafkaThreadsConfig(
			k.netThreads, k.ioThreads) + "\n\n")

	kafkaUserdata.WriteString(
		KafkaNetworkBufferSetting(k.netBufferSizeBytes,
			k.netBufferSizeBytes, k.netMaxReqSizeBytes) + "\n\n")

	kafkaUserdata.WriteString(
		KafkaLogDirsKraft(k.logDirs) + "\n")

	kafkaUserdata.WriteString(
		KafkaDefaultTopicPartitions(k.partitionsPerTopicDefault) + "\n")

	kafkaUserdata.WriteString(
		KafkaRecoveryThreads(k.recoveryThreads) + "\n")

	kafkaUserdata.WriteString(
		KafkaReplicationSettings(k.replicationFactor) + "\n\n")

	kafkaUserdata.WriteString(KafkaLogRetentionConfig(
		k.logRetentionHours, k.maxLogsegmentSize, k.balanceInterval) + "\n\n")

	kafkaUserdata.WriteString(KafkaLogFlushSettings(10000, 1000, false) + "\n")

	return kafkaUserdata.String()
}

func KafkaRoleSetting(role kafka.KafkaRole) string {
	switch role {
	case kafka.Broker:
		return "process.roles=broker"
	case kafka.BrokerController:
		return "process.roles=broker,controller"
	case kafka.Controller:
		return "process.roles=controller"
	case kafka.Zookeeper:
		return ""
	default:
		return ""
	}
}

func KafkaFormatKraftStorage() string {
	return fmt.Sprintf(kafka.KRAFT_FORMAT_CLUSTER,
		utils.GetUUID(),
	)
}

func KafkaNodeIdSetting(nodeId int) string {
	return fmt.Sprintf("node.id=%d", nodeId)
}

func KafkaListenerSettings(vmPort string) string {
	repl := "$VM_PORT"

	return strings.Replace(kafka.KAFKA_LISTENERS, repl, vmPort, 1)

	// default listeners=PLAINTEXT://:9092,CONTROLLER://:9093
	// "listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093,EXTERNAL://0.0.0.0:$VM_PORT"
}

func KafkaAdvertisedListeners(vmIPorDomain, hostIP, hostPort string) string {
	repl := "$HOST_PUBIP:$HOST_PUBPORT"
	dom := "$VM_DOMAIN_OR_IP"

	return strings.Replace(
		strings.Replace(kafka.KAFKA_ADVERTISED_LISTENERS, dom, vmIPorDomain, 1),
		repl,
		fmt.Sprintf("%s:%s", hostIP, hostPort), 1)

	/*
		If Host can resolve domain to Private IP of VM we can use:
		PLAINTEXT://kafka.kuro.com:9092
		Otherwise use the actual IP of the VM
		PLAINTEXT://192.168.1.10:9092

		If we cannot successfully replace - use default
		- Default:
		advertised.listeners=PLAINTEXT://kafka.kuro.com:9092,EXTERNAL://192.168.1.10:9094
		- Template:
		advertised.listeners=PLAINTEXT://$VM_DOMAIN_OR_IP:9092,EXTERNAL://$HOST_PUBIP:$HOST_PUBPORT
	*/
}

func ExposeBrokerCmd(
	domain string,
	vmPort int, hostPort int,
	externalIP string,
) string {
	return fmt.Sprintf(kafka.KAFKA_EXPOSE_BROKER, domain, vmPort, hostPort, externalIP)
}

/*
Network Buffer Settings for Kafka
Defaults are:
100KB for Net Send/Recv Buffers
100MB for Max Request Size (important for OOM Heap issues)
*/
func KafkaNetworkBufferSetting(netSendBuffer, netRecvBuffer, netMaxReqSize int) string {
	return fmt.Sprintf(kafka.KAFKA_NETWORK_BUFFER,
		netSendBuffer, netRecvBuffer, netMaxReqSize)
}

/*
Threads Settings for Compute
- The number of threads that the server uses for receiving requests from the network and sending responses to the network
num.network.threads=3

- The number of threads that the server uses for processing requests, which may include disk I/O
num.io.threads=8
*/
func KafkaThreadsConfig(netThreads, ioThreads int) string {
	return fmt.Sprintf(kafka.KAFKA_THREADS, netThreads, ioThreads)
}

/*
Internal Topic Settings
Replication factor for group metadata internal topics
- __consumer_offsets
- __transaction_state

For Dev  ->  1
For Prod -> 3  (recommended)

	offsets.topic.replication.factor=1
	transaction.state.log.replication.factor=1
	transaction.state.log.min.isr=1
*/
func KafkaReplicationSettings(replicationFactor int) string {
	return fmt.Sprintf(kafka.KAFKA_REPLICATION,
		replicationFactor, replicationFactor, replicationFactor)
}

/*
A comma separated list of directories under which to store log files
log.dirs=/tmp/kraft-combined-logs
*/
func KafkaLogDirsKraft(dirs []string) string {
	return fmt.Sprintf("log.dirs=%s\n", strings.Join(dirs, ","))
}

/*
The default number of log partitions per topic. More partitions allow greater
parallelism for consumption, but this will also result in more files across
the brokers.

	num.partitions=1
*/
func KafkaDefaultTopicPartitions(defaultPartitions int) string {
	return fmt.Sprintf("num.partitions=%d\n", defaultPartitions)
}

/*
The number of threads per data directory to be used for log recovery at startup and flushing at shutdown.
This value is recommended to be increased for installations with data dirs located in RAID array.

	num.recovery.threads.per.data.dir=1
*/
func KafkaRecoveryThreads(numThreads int) string {
	return fmt.Sprintf("num.recovery.threads.per.data.dir=%d\n", numThreads)
}

/*
Log Retention Settings for the Kafka Server:

- The minimum age of a log file to be eligible for deletion due to age

	log.retention.hours=168 # 1 Week

- The maximum size of a log segment file. When this size is reached a new log segment will be created.

	log.segment.bytes=1073741824 # 1024 MB

- The interval at which log segments are checked to see if they can be deleted according to the retention policies

	log.retention.check.interval.ms=300000 # 5 minutes

- A size-based retention policy for logs. Segments are pruned from the log unless the remaining segments drop below log.retention.bytes. Functions independently of log.retention.hours.

	# log.retention.bytes=1073741824 (OPTIONAL)


	log.retention.hours=168
	log.segment.bytes=1073741824
	log.retention.check.interval.ms=300000
*/
func KafkaLogRetentionConfig(retentionHours, maxLogsegmentSize, balanceInterval int) string {
	return fmt.Sprintf(kafka.KAFKA_LOG_RETENTION,
		retentionHours, maxLogsegmentSize, balanceInterval)
}

/*
Kafka Event Flush Settings:

- Kafka immediately writes to the Filesystem Cache

- fsync() (Event Flushing) from Cache to Physical Disk is evaluated Lazily

	-- 1. Durability -> Unflushed data may be lost if you are not using replication.

	== 2. Latency -> Very large flush intervals may lead to latency spikes when the flush does occur as there will be a lot of data to flush.

	-- 3. Throughput -> The flush is generally the most expensive operation, and a small flush interval may lead to excessive seeks.

Flushing can be configured according to:
  - Time
  - Number of Messsages
  - Both

Settings can be applied at the:

  - Global Level
  - Topic level

- The number of messages to accept before forcing a flush of data to disk

	log.flush.interval.messages=10000

- The maximum amount of time a message can sit in a log before we force a flush

	log.flush.interval.ms=1000
*/
func KafkaLogFlushSettings(numEventsThreshold, flushInterval int, active bool) string {
	if active {
		return fmt.Sprintf("log.flush.interval.messages=%d\nlog.flush.interval.ms=%d", numEventsThreshold, flushInterval)
	} else {
		return fmt.Sprintf("#log.flush.interval.messages=%d\n#log.flush.interval.ms=%d", numEventsThreshold, flushInterval)
	}
}

/*

# Network Expose Cmd
# go run main.go --expose-vm=$DOMAIN \
# --port=$VM_PORT \
# --hostport=$HOST_PUBPORT \
# --external-ip=$EXT_IP \
# --protocol=tcp
*/
