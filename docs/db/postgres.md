# Postgres

## linux

https://dev.to/rainbowhat/postgresql-16-installation-on-ubuntu-2204-51ia

```bash
# client on machines needing to connect
sudo apt install postgresql-client

sudo apt update
sudo apt install gnupg2 wget vim -y

sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'

curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg

sudo apt update
sudo apt install postgresql-16 postgresql-contrib-16 -y

sudo systemctl status postgresql

sudo systemctl start postgresql
sudo systemctl enable postgresql



sudo vi /etc/postgresql/16/main/postgresql.conf
listen_addresses = '*'


# Remote Connections - update pg_hba.conf
sudo sed -i '/^host/s/ident/md5/' /etc/postgresql/16/main/pg_hba.conf
sudo sed -i '/^local/s/peer/trust/' /etc/postgresql/16/main/pg_hba.conf
echo "host all all 0.0.0.0/0 md5" | sudo tee -a /etc/postgresql/16/main/pg_hba.conf

sudo systemctl restart postgresql

# validate listening on all addresses
ss -nlt | grep 5432
# 0.0.0.0:5432 - listening from all ipv4 to 5432
# [::]:5432    - listening from all ipv6 to 5432

sudo ufw allow 5432/tcp # Optional if using uncomplicated firewall

# login and start psql
sudo -u postgres psql
SELECT version();
ALTER USER postgres PASSWORD 'password';

# attempt remote connection
psql -h <your-server-ip> -U postgres -d postgres
psql -h 192.168.122.24 -U postgres -d postgres # go test -run TestGetHostIpRecommended to get IP

ALTER USER postgres PASSWORD 'VeryStronGPassWord@1137';

# 16 char password
openssl rand -base64 16

```

## commands

```bash
# Open psql terminal
psql postgres

# create role (run from outside of psql)
createuser kuro --interactive

createdb test # create db
psql -d test  # connect

# view tables
\dt

## using tls

mkdir -p ~/postgres_cert
cd ~/postgres_cert
# Generates Private Key (key.pem) and Self Signed Cert (cert.pem) valid for 365 days
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/C=US/ST=DC/L=Georgetown/O=Kuro/OU=Kuro/CN=Kurosaki/emailAddress=org@dev.com"

# find data dir
brew services list
ls /usr/local/var/postgresql@16

# move certs to pg
sudo mv ~/postgres_cert/cert.pem /usr/local/var/postgresql@16/server.crt
sudo mv ~/postgres_cert/key.pem /usr/local/var/postgresql@16/server.key
chmod 600 /usr/local/var/postgresql@16/server.key

vi /usr/local/var/postgresql@16/postgresql.conf

# edit postgresql.conf  (change requires restart)
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'

sudo systemctl restart postgresql

# Connect and verify SSL
psql "dbname=test user=kuro host=localhost sslmode=require"

SELECT * FROM pg_stat_ssl WHERE pid = pg_backend_pid();

```
