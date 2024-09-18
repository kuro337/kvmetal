package kafka

const KAFKA_MISC = `
# checking successful connection
nc -vz kuro.com 9094
nc -vx 192.168.1.10:9094

# KafkaCat Validations
sudo apt-get update && sudo apt-get install kafkacat
brew install kafkacat
kcat -b kuro.com:9095 -L

kafka-topics.sh --bootstrap-server <broker_ip>:port --list
kafka-topics.sh --bootstrap-server<broker_ip>:port --create --topic my_new_topic --partitions 3 --replication-factor 1
echo "Hello Kafka" | kafka-console-producer.sh --broker-list <broker_ip>:port --topic my_new_topic
kafka-console-consumer.sh --bootstrap-server <broker_ip>:port --topic my_new_topic --from-beginning


`

const KAFKA_CHECK = `
ps aux | grep kafka | grep -v grep
sudo kill -9 16892


`

const KAFKA_INSTALL = `
# linux
wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
sudo tar -xzf kafka_2.13-3.7.0.tgz
sudo rm -f kafka_2.13-3.7.0.tgz
mv kafka_2.13-3.7.0 kafka
sudo mv kafka /usr/local/kafka
echo 'export PATH=/usr/local/kafka/bin:$PATH' >> ~/.zshrc
source ~/.zshrc

# mac 

curl -O https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
sudo tar -xzf kafka_2.13-3.7.0.tgz
sudo rm -f kafka_2.13-3.7.0.tgz
sudo mv kafka_2.13-3.7.0 kafka
sudo mv kafka /usr/local/kafka
echo 'export PATH=/usr/local/kafka/bin:$PATH' >> ~/.zshrc
source ~/.zshrc

`
