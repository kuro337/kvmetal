# Kafka CLI Tools

```bash


wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz
sudo tar -xzf kafka_2.13-3.7.0.tgz
sudo rm -f kafka_2.13-3.7.0.tgz
mv kafka_2.13-3.7.0 kafka
sudo mv kafka /usr/local/kafka
echo 'export PATH=/usr/local/kafka/bin:$PATH' >> ~/.zshrc
source ~/.zshrc


https://forum.confluent.io/t/kafka-does-not-works-through-nat-as-expected/4351


# Host to VM is still on Port 9092

kafka-topics.sh --bootstrap-server 192.168.122.113:9092 --list


kafka-topics.sh --bootstrap-server kraft.kuro.com:9092 --list

kafka-topics.sh --bootstrap-server kraft.kuro.com:9092 --list

kafka-topics.sh --bootstrap-server kafka.kuro.com:9092 --create --topic my_new_topic --partitions 3 --replication-factor 1

echo "Hello from external client" | kafka-console-producer.sh --broker-list kafka.kuro.com:9092 --topic my_new_topic

kafka-console-consumer.sh --bootstrap-server kafka.kuro.com:9092 --topic my_new_topic --from-beginning



# current

# export PATH=/usr/local/kafka/bin:$PATH
/usr/local/kafka_2.13-3.7.0/bin/kafka-topics.sh



```
