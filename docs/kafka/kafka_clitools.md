# Kafka CLI Tools

```bash


wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
sudo tar -xzf kafka_2.13-3.7.0.tgz
sudo rm -f kafka_2.13-3.7.0.tgz
mv kafka_2.13-3.7.0 kafka
sudo mv kafka /usr/local/kafka
echo 'export PATH=/usr/local/kafka/bin:$PATH' >> ~/.zshrc
source ~/.zshrc

kafka-topics.sh

kafka-topics.sh --bootstrap-server <kafka-broker-ip>:9092 --list



kafka-topics.sh --bootstrap-server kafka.kvm:9092 --list

kafka-topics.sh --bootstrap-server 192.168.122.243:9092 --list

kafka-topics.sh --bootstrap-server kafka:9092 --create --topic my_new_topic --partitions 3 --replication-factor 1


192.168.122.211

# current
# export PATH=/usr/local/kafka/bin:$PATH
/usr/local/kafka_2.13-3.7.0/bin/kafka-topics.sh

```
