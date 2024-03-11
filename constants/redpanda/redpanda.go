package redpanda

var REDPANDA_RUNCMD_INITIAL_STEPS = []string{
	`curl -1sLf 'https://dl.redpanda.com/nzc4ZYQK3WRGd9sy/redpanda/cfg/setup/bash.deb.sh' | sudo -E bash`,
	`sudo apt-get install redpanda`,
	`sudo rpk redpanda mode production`,
	`sudo rpk redpanda tune`,
}

const REDPANDA_SETTINGS_RUNCMD_TEMPLATE = `sudo tee /etc/redpanda/redpanda.yaml  > /dev/null <<EOL
%s
EOL
`
const REDPANDA_START_CMD = `sudo systemctl start redpanda`

const REDPANDA_ADVERTISED_EXTERNAL_KAFKA_API = `
				- name: external
					address: $HOST_IP
					port: $HOST_PORT
`

const REDPANDA_ADVERTISED_INTERNAL_KAFKA_API = `
				- name: internal		
					address: $VM_IP
					port: 9092
`

const REDPANDA_EXTERNAL_KAFKA_API = `
				- name: external
					address: 0.0.0.0
					port: $VM_PORT
`

const REDPANDA_RUNCMD = `
curl -1sLf 'https://dl.redpanda.com/nzc4ZYQK3WRGd9sy/redpanda/cfg/setup/bash.deb.sh' | sudo -E bash
sudo apt-get install redpanda

sudo rpk redpanda mode production
sudo rpk redpanda tune

sudo systemctl start redpanda
rpk cluster info
`

const REDPANDA_SETTINGS_TEMPLATE = `
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
          port: $VM_PORT
    admin:
        - address: 0.0.0.0
          port: 9644
    advertised_rpc_api:
        address: 127.0.0.1
        port: 33145
    advertised_kafka_api:
        - name: internal		
          address: $VM_IP
          port: 9092
        - name: external
          address: $HOST_IP
          port: $HOST_PORT
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

const REDPANDA_SETTINGS = `

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
          port: 9095				# VM_PORT
    admin:
        - address: 0.0.0.0
          port: 9644
    advertised_rpc_api:
        address: 127.0.0.1
        port: 33145
    advertised_kafka_api:
        - name: internal		
          address: 192.168.122.50 # VM_IP
          port: 9092
        - name: external
          address: 192.168.1.10 # HOST_IP
          port: 8090						# HOST_PORT
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
