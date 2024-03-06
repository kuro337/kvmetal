package kafka

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
  #- sudo sed -i 's/#listeners=PLAINTEXT:\/\/[^,]*/listeners=PLAINTEXT:\/\/0.0.0.0:9092/' /opt/kafka/config/kraft/server.properties
  #- sudo sed -i 's/#advertised.listeners=PLAINTEXT:\/\/your.host.name:9092/advertised.listeners=PLAINTEXT:\/\/127.0.0.1:9092/' /opt/kafka/config/kraft/server.properties
  - sudo echo -e "listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093\n\n" >> /opt/kafka/config/kraft/server.properties
  - sudo echo -e "advertised.listeners=PLAINTEXT://kafka:9092\n\n" >> /opt/kafka/config/kraft/server.properties
  - sudo /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/kraft/server.properties

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
