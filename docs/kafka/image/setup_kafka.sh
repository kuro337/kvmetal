#!/bin/bash
# Format Kafka storage and set up initial configuration
/opt/kafka/bin/kafka-storage.sh format -t f9afb3af-f962-409e-b243-3ec98001cef7 -c /opt/kafka/config/kraft/server.properties

sudo tee /opt/kafka/config/kraft/server.properties > /dev/null <<EOL
process.roles=broker,controller
node.id=1

controller.quorum.voters=1@localhost:9093

listeners=PLAINTEXT://0.0.0.0:9092,CONTROLLER://:9093,EXTERNAL://0.0.0.0:9095
advertised.listeners=PLAINTEXT://kraft.kuro.com:9092,EXTERNAL://192.168.1.10:9094

inter.broker.listener.name=PLAINTEXT
controller.listener.names=CONTROLLER

listener.security.protocol.map=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,SSL:SSL,SASL_PLAINTEXT:SASL_PLAINTEXT,SASL_SSL:SASL_SSL,EXTERNAL:PLAINTEXT

num.network.threads=3
num.io.threads=8

socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600

log.dirs=/tmp/kraft-combined-logs

num.partitions=1

num.recovery.threads.per.data.dir=1

offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1

log.retention.hours=168
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000

#log.flush.interval.messages=10000
#log.flush.interval.ms=1000