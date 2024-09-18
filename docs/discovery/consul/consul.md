# Consul

```bash
# install on a server

wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list

sudo apt update && sudo apt install consul


sudo chown -R <user>:<group> /var/consul
sudo chown <user>:<group> /etc/consul.d/consul-config.json

# bind_addr - set to IP of the VM consul is installed on

/etc/consul.d/consul-config.json

{
  "server": true,
  "node_name": "consul-server-1",
  "datacenter": "dc1",
  "data_dir": "/var/consul",
  "bind_addr": "192.168.122.173",
  "client_addr": "0.0.0.0",
  "bootstrap_expect": 1,
  "ui": true,
  "log_level": "INFO",
  "enable_syslog": true
}

# start consul
sudo consul agent -config-dir=/etc/consul.d/

# verify

consul members

```

## systemd service for consul

```bash
[Unit]
Description="Consul Service"
After=network.target

[Service]
Type=simple
User=consul
Group=consul
ExecStart=/usr/bin/consul agent -config-dir=/etc/consul.d/
Restart=on-failure

[Install]
WantedBy=multi-user.target

# sudo systemctl enable consul
# sudo systemctl start consul

```

## launching services

Now if we launch Postgres with its own IP

```json
// consul config - /etc/consul.d/consul-config.json
192.168.122.70
{
  "datacenter": "dc1",
  "data_dir": "/var/consul",
  "bind_addr": "192.168.122.70",
  "client_addr": "0.0.0.0",
  "retry_join": ["192.168.122.173"]
}

// Dont need to set own IP - can bind to 0.0.0.0
{
  "datacenter": "dc1",
  "data_dir": "/var/consul",
  "bind_addr": "0.0.0.0",
  "client_addr": "0.0.0.0",
  "retry_join": ["192.168.122.173"]
}


Bind Addr   -> IP of Server running Service
Client Addr -> Accept Connections from Anywhere
Retry Join  -> IPs to Join at Startup (Main Consul Server)

TestVM   -> 192.168.122.70
Postgres -> 192.168.122.24
Consul   -> 192.168.122.173

// Service Definition /etc/consul.d/postgres-service.json
{
  "service": {
    "name": "postgresql",
    "tags": ["db"],
    "port": 5432,
    "check": {
      "tcp": "localhost:5432",
      "interval": "10s",
      "timeout": "1s"
    }
  }
}


```

```bash
screen -S consul
sudo consul agent -config-dir=/etc/consul.d/
ctrl ad

consul catalog services
- consul
- postgresql

# Now we can discover our service - consul runs on port 8680
dig @127.0.0.1 -p 8600 postgresql.service.consul

# or use HTTP client for Consul
curl http://localhost:8500/v1/catalog/service/postgresql


consul reload # for reload
```

Service Systemd file

```bash
[Unit]
Description=Consul
Documentation=https://www.consul.io/
Requires=network-online.target
After=network-online.target

[Service]
User=consul
Group=consul
ExecStart=/usr/bin/consul agent -config-dir=/etc/consul.d/
ExecReload=/bin/kill -HUP $MAINPID
KillSignal=SIGINT
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target


```
