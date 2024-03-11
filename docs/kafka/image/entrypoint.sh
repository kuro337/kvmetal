#!/bin/bash
# Start Kafka
/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties

# Keep the container running (adjust according to your needs)
tail -f /dev/null
