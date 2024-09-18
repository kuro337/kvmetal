package kafka

const KAFKA_REQUIRED_PARAMS = `
DOMAIN=kafkavm
VM_PORT=9095
HOST_PUBPORT=9094
HOST_PUBIP=192.168.1.10
VM_IP=192.168.122.20 # or kafka.kuro.com if we know host can resolve
EXT_IP=192.168.1.225
`

const (
	KAFKA_LISTENERS            = "listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093,EXTERNAL://0.0.0.0:$VM_PORT"
	KAFKA_ADVERTISED_LISTENERS = "advertised.listeners=PLAINTEXT://$VM_DOMAIN_OR_IP:9092,EXTERNAL://$HOST_PUBIP:$HOST_PUBPORT"
	KAFKA_CONTROLLER_QUORUM    = "controller.quorum.voters=1@localhost:9093"
	KRAFT_FORMAT_CLUSTER       = "/opt/kafka/bin/kafka-storage.sh format -t %s -c /opt/kafka/config/kraft/server.properties"

	KAFKA_KRAFT_START_CLUSTER = "sudo /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/kraft/server.properties"
	KAFKA_ZOO_START_CLUSTER   = "sudo sh -c 'nohup /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties > /opt/kafka/logs/kafka.log 2>&1 &'"
	KAFKA_NETWORK_BUFFER      = `socket.send.buffer.bytes=%d
socket.receive.buffer.bytes=%d
socket.request.max.bytes=%d`

	KAFKA_EXPOSE_BROKER = `go run main.go --expose-vm=%s \
  --port=%d \
  --hostport=%d \
  --external-ip=%s \
  --protocol=tcp`

	KAFKA_THREADS = `num.network.threads=%d
num.io.threads=%d`

	KAFKA_REPLICATION = `offsets.topic.replication.factor=%d
transaction.state.log.replication.factor=%d
transaction.state.log.min.isr=%d`

	KAFKA_LOG_RETENTION = `log.retention.hours=%d
log.segment.bytes=%d
log.retention.check.interval.ms=%d`

	KAFKA_SETTINGS_RUNCMD_TEMPLATE = `sudo tee /opt/kafka/config/kraft/server.properties > /dev/null <<EOL
%s
EOL
`
)

var KAFKA_RUNCMD_INITIAL_STEPS = []string{
	`wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz`,
	`tar -xzf kafka*.tgz && rm kafka*.tgz`,
	`mv kafka_*3.7.0 /opt/kafka`,
	`sudo sh -c 'echo "export KAFKA_HOME=/opt/kafka" >> /etc/profile.d/kafka.sh'`,
	`sudo sh -c 'echo "export PATH=\$PATH:\$KAFKA_HOME/bin" >> /etc/profile.d/kafka.sh'`,
	`sudo chmod +x /etc/profile.d/kafka.sh`,
	`source /etc/profile.d/kafka.sh`,
	`sudo mkdir -p /opt/kafka/logs`,
}

const KAFKA_LISTENER_SETTINGS = `
inter.broker.listener.name=PLAINTEXT
controller.listener.names=CONTROLLER

listener.security.protocol.map=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,SSL:SSL,SASL_PLAINTEXT:SASL_PLAINTEXT,SASL_SSL:SASL_SSL,EXTERNAL:PLAINTEXT
`

const KAFKA_NET_BUFFER_SETTINGS = `
# The send buffer (SO_SNDBUF) used by the socket server 1 (100KB)
socket.send.buffer.bytes=102400

# The receive buffer (SO_RCVBUF) used by the socket server (100KB)
socket.receive.buffer.bytes=102400

# The maximum size of a request that the socket server will accept (protection against OOM)
# 100MB
socket.request.max.bytes=104857600
`

/* For Dev 1 is recommended, Prod 3 is recommended */
const KAFKA_TOPIC_SETTINGS = `
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
`

const KAFKA_PORTFWD = `
go run main.go --expose-vm=$DOMAIN \
--port=$VM_PORT \
--hostport=$HOST_PUBPORT \
--external-ip=$EXT_IP \
--protocol=tcp
`

/* Runs Uuid.randomUuid | https://github.com/apache/kafka/blob/trunk/core/src/main/scala/kafka/tools/StorageTool.scala */
const KRAFT_FORMAT_DISK_CMD = `  - KAFKA_CLUSTER_ID="$(/opt/kafka/bin/kafka-storage.sh random-uuid)"`

/* Created from Running /opt/kafka/bin/kafka-storage.sh random-uuid */
const KAFKA_NODE_ID = "node.id=$ID"

const KAFKA_ROLES = `
process.roles=broker,controller
process.roles=broker
process.roles=controller
# process.roles=
# If process.roles is not set at all, it is assumed to be in ZooKeeper mode.
`

type KafkaRole int

const (
	BrokerController KafkaRole = iota
	Broker
	Controller
	Zookeeper
)

const KAFKA_ZOOKEEPER_RUNCMD = `

  - wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
  - tar -xzf kafka*.tgz && rm kafka*.tgz
  - mv kafka_*3.7.0 /opt/kafka
  - sudo sh -c 'echo "export KAFKA_HOME=/opt/kafka" >> /etc/profile.d/kafka.sh'
  - sudo sh -c 'echo "export PATH=\$PATH:\$KAFKA_HOME/bin" >> /etc/profile.d/kafka.sh'
  - sudo chmod +x /etc/profile.d/kafka.sh
  - source /etc/profile.d/kafka.sh
  - sudo mkdir -p /opt/kafka/logs
  - sudo sh -c 'echo "auto.create.topics.enable=true" >> /opt/kafka/config/server.properties'
  - sudo sh -c 'nohup /opt/kafka/bin/zookeeper-server-start.sh /opt/kafka/config/zookeeper.properties > /opt/kafka/logs/zookeeper.log 2>&1 &'
  - sleep 10
  - sudo sh -c 'nohup /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties > /opt/kafka/logs/kafka.log 2>&1 &'
  

final_message: "Kafka has been successfully installed and started."

`

const KAFKA_KRAFT_RUNCMD = `
  - wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
  - tar -xzf kafka*.tgz && rm kafka*.tgz
  - mv kafka_*3.7.0 /opt/kafka
  - sudo sh -c 'echo "export KAFKA_HOME=/opt/kafka" >> /etc/profile.d/kafka.sh'
  - sudo sh -c 'echo "export PATH=\$PATH:\$KAFKA_HOME/bin" >> /etc/profile.d/kafka.sh'
  - sudo chmod +x /etc/profile.d/kafka.sh
  - source /etc/profile.d/kafka.sh
  - sudo mkdir -p /opt/kafka/logs
  - KAFKA_CLUSTER_ID="$(/opt/kafka/bin/kafka-storage.sh random-uuid)"
  - /opt/kafka/bin/kafka-storage.sh format -t $KAFKA_CLUSTER_ID -c /opt/kafka/config/kraft/server.properties

  ###- sudo echo -e "listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093\n\n" >> /opt/kafka/config/kraft/server.properties
  ###- sudo echo -e "advertised.listeners=PLAINTEXT://192.168.1.10:9092:9092\n\n" >> /opt/kafka/config/kraft/server.properties
###- sudo echo -e "advertised.listeners=PLAINTEXT://$FQDN:9092\n\n" >> /opt/kafka/config/kraft/server.properties

  #- echo "listeners=INTERNAL://0.0.0.0:9092,EXTERNAL://0.0.0.0:9093" | sudo tee -a /opt/kafka/config/kraft/server.properties
  #- echo "advertised.listeners=INTERNAL://kafka.internal:9092,EXTERNAL://kuro.com:9093" | sudo tee -a /opt/kafka/config/kraft/server.properties
  #- echo "listener.security.protocol.map=INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT" | sudo tee -a /opt/kafka/config/kraft/server.properties
  #- echo "inter.broker.listener.name=INTERNAL" | sudo tee -a /opt/kafka/config/kraft/server.properties

  #- sudo /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/kraft/server.properties

final_message: "Kafka has been successfully installed and started."

`

// Roles -> Broker/Controller.

const KRAFT_CONFIG = `
# Set the process roles to both broker and controller for a standalone node
process.roles=broker,controller

# Assign a unique node ID
node.id=1

# Setup controller quorum voters for KRaft
controller.quorum.voters=1@localhost:9093

# Listener configuration for clients to connect
listeners=PLAINTEXT://localhost:9092

# Define which listener to use for inter-broker communication
inter.broker.listener.name=PLAINTEXT

# Directory where Kafka stores log files
log.dirs=/var/lib/kafka/data

# Enable deletion of topic
delete.topic.enable=true
`

const KAFKA_PKG = `
  - openjdk-11-jdk
  - wget
  - tar
  - default-jre
`
