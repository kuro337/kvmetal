package discovery

const CONSUL_UBUNTU_RUNCMD = `

  - wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
  - echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
  - sudo apt update && sudo apt install consul

`

/*

wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list

sudo apt update && sudo apt install consul


wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install consul

192.168.122.173

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

*/
