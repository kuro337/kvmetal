# kafka

```bash

https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz

bin/zookeeper-server-start.sh config/zookeeper.properties # launch (screen)
bin/kafka-server-start.sh config/server.properties # launch kafka (screen)

packages:
  - openjdk-11-jdk
  - wget
  - tar
  - default-jre

runcmd:
  - wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
  - tar -xzf kafka-*.tgz && rm kafka-*.tgz
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

```

## Exposing

- If we want to communicate with the Cluster Node externally - we need to bind to a network interface.

- Add this to `/opt/kafka/config/server.properties`

```bash
listeners=PLAINTEXT://0.0.0.0:9092

# After Adding it - restart the server

sudo /opt/kafka/bin/kafka-server-stop.sh
sudo /opt/kafka/bin/kafka-server-start.sh -daemon /opt/kafka/config/kraft/server.properties

# Use a Kafka Client/Kafka CLI tool to interact with the Cluster (from other machine)

# publish
/path/to/kafka-console-producer.sh --bootstrap-server <VM_IP>:9092 --topic test-topic

# consume
/path/to/kafka-console-consumer.sh --bootstrap-server <VM_IP>:9092 --topic test-topic --from-beginning


```

Validation

```bash
ps aux | grep zookeeper
ps aux | grep kafka

Kafka Connect for Reading from DBs/etc.
Streams: https://kafka.apache.org/documentation/streams/quickstart

# Create a Topic
/opt/kafka/bin/kafka-topics.sh --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic test-topic
# describe topic
/opt/kafka/bin/kafka-topics.sh --describe --topic test-topic --bootstrap-server localhost:9092
# List Topics
/opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092
# Produce some Messages
/opt/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic test-topic
# Consume Messages
/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic test-topic --from-beginning

# stopping Kafka & zookeeper before launching in KRAFT
jps | grep Kafka
sudo kill -SIGINT 16531
sudo jps | grep QuorumPeerMain
sudo kill -SIGINT 16111

sudo lsof -i :9092

```

## using Kraft

```bash
# Each Server needs a UUID
KAFKA_CLUSTER_ID="$(/opt/kafka/bin/kafka-storage.sh random-uuid)"

# Format Log Directories
/opt/kafka/bin/kafka-storage.sh format -t $KAFKA_CLUSTER_ID -c /opt/kafka/config/kraft/server.properties

echo -e "listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093\n\n" >> sample
echo -e "advertised.listeners=PLAINTEXT://kafka.kuro.com:9092\n\n" >> sample
# kraft/server.properties line:
listeners=PLAINTEXT://:9092,CONTROLLER://:9093  # change to below
listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093

listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093
advertised.listeners=PLAINTEXT://kafka.kuro.com:9092


# Other Settings try
listeners=INTERNAL://0.0.0.0:9092,EXTERNAL://0.0.0.0:19092
listener.security.protocol.map=INTERNAL:PLAINTEXT,EXTERNAL:PLAINTEXT
advertised.listeners=EXTERNAL://[External_IP]:19092,INTERNAL://[Internal_IP]:9092
inter.broker.listener.name=INTERNAL

# CONTROLLER is for KRAFT, PLAINTEXT is for Clients (ssl,etc.)

# if we want to listen for reqs externally
# Change listeners
sudo sed -i 's/#listeners=PLAINTEXT:\/\/[^,]*/listeners=PLAINTEXT:\/\/0.0.0.0:9092/' /opt/kafka/config/kraft/server.properties

# set to hostname
listeners=PLAINTEXT://0.0.0.0:9092
# set to public ip of VM
advertised.listeners=PLAINTEXT://192.168.122.243:9092

# just add it at this point - to end of /opt/kafka/config/kraft/server.properties
listeners=PLAINTEXT://0.0.0.0:9092
# set to public ip of VM
advertised.listeners=PLAINTEXT://192.168.122.243:9092


# Change advertised.listeners
sudo sed -i 's/#advertised.listeners=PLAINTEXT:\/\/0.0.0.0:9092/advertised.listeners=PLAINTEXT:\/\/127.0.0.1:9092/' /opt/kafka/config/kraft/server.properties

# Change listeners
sudo sed -i 's/#listeners=PLAINTEXT:\/\/[^,]*/listeners=PLAINTEXT:\/\/0.0.0.0:9092/' /opt/kafka/config/kraft/server.properties

# Change advertised.listeners
sudo sed -i 's/#advertised.listeners=PLAINTEXT:\/\/your.host.name:9092/advertised.listeners=PLAINTEXT:\/\/127.0.0.1:9092/' /opt/kafka/config/kraft/server.properties

# Launch in KRAFT mode
sudo /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/kraft/server.properties

```

- Adjusting server.properties

```bash
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
```

## Example 3 Node Setup - 1 Controller 2 Brokers

```bash

Controller Nodes do not need to define Listeners

Unless they also serve client requests or other specific network configurations.


# Controller -> Node 1 Configuration
process.roles=controller
node.id=1
controller.quorum.voters=1@node1:9093,2@node2:9093,3@node3:9093

# Broker -> Node 2 Configuration:
process.roles=broker,controller
node.id=2
controller.quorum.voters=1@node1:9093,2@node2:9093,3@node3:9093
listeners=PLAINTEXT://node2:9092
inter.broker.listener.name=PLAINTEXT
Node 3 Configuration:

# Broker -> Node 3 Configuration:
process.roles=broker,controller
node.id=3
controller.quorum.voters=1@node1:9093,2@node2:9093,3@node3:9093
listeners=PLAINTEXT://node3:9092
inter.broker.listener.name=PLAINTEXT


sudo vi /opt/kafka/config/kraft/server.properties
sudo cat /opt/kafka/config/kraft/server.properties

sudo /opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/kraft/server.properties
```
