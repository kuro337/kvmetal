# Redpanda

## Sizing Considerations

https://docs.redpanda.com/current/deploy/deployment-option/self-hosted/manual/sizing/

- 2 GB per Core

- 4mb memory per topic partition replica (topic_memory_per_partition)

```bash

https://docs.redpanda.com/current/deploy/deployment-option/self-hosted/manual/production/dev-deployment/


curl -1sLf \
  'https://dl.redpanda.com/nzc4ZYQK3WRGd9sy/redpanda/cfg/setup/bash.deb.sh' \
  | sudo -E bash
sudo apt-get install redpanda

# Install Web Console (optional)
curl -1sLf 'https://dl.redpanda.com/nzc4ZYQK3WRGd9sy/redpanda/cfg/setup/bash.deb.sh' | \
sudo -E bash && sudo apt-get install redpanda-console -y


# <listener-address>  - 0.0.0.0
# <your-machine-ip>   - IP of Server Running | go run main.go --getip redpanda

# 192.168.122.50


## Default Config
sudo rpk redpanda mode production
sudo rpk redpanda tune
sudo systemctl start redpanda
rpk cluster info

rpk topic create test
curl -s "localhost:8082/topics"

## Validations from Host
curl -s "http://192.168.122.50:8082/topics"
curl -s "http://redpanda.kuro.com:8082/topics"
kafka-topics.sh --bootstrap-server redpanda.kuro.com:9092 --list
kafka-topics.sh --bootstrap-server redpanda.kuro.com:9092 --create --topic my_new_topic --partitions 3 --replication-factor 1

kafka-topics.sh --bootstrap-server rpanda.kuro.com:9092 --list
kafka-topics.sh --bootstrap-server 192.168.122.52:9092 --list



echo "Hello from external client" | kafka-console-producer.sh --broker-list redpanda.kuro.com:9092 --topic test

# Expose
go run main.go --expose-vm=redpanda \
--port=9095 \
--hostport=8090 \
--external-ip=192.168.1.225 \
--protocol=tcp

## Configuring Settings

sudo rpk redpanda config bootstrap --self <listener-address> --ips <seed-server1-ip>,<seed-server2-ip>,<seed-server3-ip> && \
sudo rpk redpanda config set redpanda.empty_seed_starts_cluster false


sudo rpk redpanda config bootstrap --self 0.0.0.0 --ips 127.0.0.1 && \
sudo rpk redpanda config set redpanda.empty_seed_starts_cluster false

# setting multiple listeners
redpanda:
  kafka_api:
    - name: external
      address: 0.0.0.0
      port: 9092
  advertised_kafka_api:
    - name: external
      address: 192.168.1.10
      port: 9092

# Check /etc/redpanda/redpanda.yaml
sudo cat /etc/redpanda/redpanda.yaml
sudo vi /etc/redpanda/redpanda.yaml

# Start
sudo systemctl start redpanda-tuner redpanda

# redpanda has an HTTP Client

curl -s "localhost:8082/topics"


go run main.go --launch-vm=redpanda --mem=8192 --cpu=4

```

Example Config - to make Kafka API Accessible over NAT

```bash
# set this to VM IP Addr

    advertised_kafka_api:
        - address: 192.168.122.50
          port: 9092
# sudo cat /etc/redpanda/redpanda.yaml

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
          port: 9095
    admin:
        - address: 0.0.0.0
          port: 9644
    advertised_rpc_api:
        address: 127.0.0.1
        port: 33145
    advertised_kafka_api:
        - name: internal
          address: 192.168.122.50
          port: 9092
        - name: external
          address: 192.168.1.10
          port: 8090
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

```

## Validations

```bash
rpk cluster info
sudo cat /etc/redpanda/redpanda.yaml


```

## reset

```bash
# Remove the data directory
sudo rm -rf /var/lib/redpanda/data

# (Optional) Create the data directory again, if Redpanda doesn't automatically
sudo mkdir -p /var/lib/redpanda/data
sudo chown redpanda:redpanda /var/lib/redpanda/data


sudo chown -R redpanda:redpanda /var/lib/redpanda/data

```
