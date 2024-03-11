# Creating an Image for Kafka Kraft

setup_kafka.sh

```bash
#!/bin/bash
# Format Kafka storage and set up initial configuration
/opt/kafka/bin/kafka-storage.sh format -t f9afb3af-f962-409e-b243-3ec98001cef7 -c /opt/kafka/config/kraft/server.properties

sudo tee /opt/kafka/config/kraft/server.properties > /dev/null <<EOL
process.roles=broker,controller
node.id=1

```

entrypoint_script.sh

```bash
#!/bin/bash
# Start Kafka or other necessary services
# For example:
/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties

# Keep the container running (if necessary)
tail -f /dev/null


```

## Build and Run Container

```bash

docker build -t my-kafka-setup .

docker run -d --name my-kafka-container my-kafka-setup

```
