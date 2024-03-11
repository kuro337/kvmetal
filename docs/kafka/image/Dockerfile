# Use an Ubuntu base image
FROM ubuntu:latest

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install necessary packages
RUN apt-get update && apt-get install -y \
  zsh \
  openjdk-11-jdk \
  default-jre \
  tar \
  wget \
  git \
  curl \
  software-properties-common \
  && rm -rf /var/lib/apt/lists/*

# Set environment variables
ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ENV PATH="${JAVA_HOME}/bin:${PATH}"
ENV KAFKA_HOME=/opt/kafka
ENV PATH="${KAFKA_HOME}/bin:${PATH}"
ENV HOME_DIR="/root"
ENV ENV_SHELL="$HOME_DIR/.zshrc"

# Install Scala
RUN wget https://downloads.lightbend.com/scala/2.13.8/scala-2.13.8.tgz -O /tmp/scala-2.13.8.tgz \
  && tar -xvzf /tmp/scala-2.13.8.tgz -C /usr/local \
  && mv /usr/local/scala-2.13.8 /usr/local/scala \
  && echo 'export PATH=$PATH:/usr/local/scala/bin' >> $ENV_SHELL

# Install sbt
RUN echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | tee /etc/apt/sources.list.d/sbt.list \
  && echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | tee /etc/apt/sources.list.d/sbt_old.list \
  && curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | apt-key add - \
  && apt-get update \
  && apt-get install -y sbt

# Setup Kafka
RUN wget https://downloads.apache.org/kafka/3.7.0/kafka_2.13-3.7.0.tgz \
  && tar -xzf kafka_2.13-3.7.0.tgz -C /opt \
  && mv /opt/kafka_2.13-3.7.0 $KAFKA_HOME \
  && mkdir -p $KAFKA_HOME/logs \
  && echo "export KAFKA_HOME=/opt/kafka" >> /etc/profile.d/kafka.sh \
  && echo "export PATH=\$PATH:\$KAFKA_HOME/bin" >> /etc/profile.d/kafka.sh \
  && chmod +x /etc/profile.d/kafka.sh

# Copy and run a script to configure Kafka (alternative to runcmd)
COPY setup_kafka.sh /usr/local/bin/setup_kafka.sh
RUN chmod +x /usr/local/bin/setup_kafka.sh && /usr/local/bin/setup_kafka.sh

# Change shell to Zsh (optional, might not be necessary in container)
RUN chsh -s $(which zsh)

# Expose Kafka ports
EXPOSE 9092 9093 9094 9095

# Kafka Settings
COPY setup_kafka.sh /usr/local/bin/setup_kafka.sh
RUN chmod +x /usr/local/bin/setup_kafka.sh && /usr/local/bin/setup_kafka.sh

# Entrypoint to start Kafka
COPY entrypoint_script.sh /usr/local/bin/entrypoint_script.sh
RUN chmod +x /usr/local/bin/entrypoint_script.sh

# Define entrypoint 
ENTRYPOINT ["/usr/local/bin/entrypoint_script.sh"]
