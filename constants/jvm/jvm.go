package jvm

const JDK_SCALA_RUNCMD = `

  - HOME_DIR="/home/ubuntu"
  - ENV_SHELL="$HOME_DIR/.zshrc"
  - JAVA_HOME_DIR=$(dirname $(dirname $(readlink -f /usr/bin/java)))
  - echo "export JAVA_HOME=$JAVA_HOME_DIR" >> $ENV_SHELL
  
  # Setup Scala
  - |
    # Step 2. Scala
    echo -e "Setting up Scala 2.13"
    wget https://downloads.lightbend.com/scala/2.13.8/scala-2.13.8.tgz -O /tmp/scala-2.13.8.tgz
    tar -xvzf /tmp/scala-2.13.8.tgz -C /tmp
    mv /tmp/scala-2.13.8 /usr/local/scala
    echo 'export PATH=$PATH:/usr/local/scala/bin' >> $ENV_SHELL
  
  # Setup sbt
  - |
    echo -e "Installing and Setting up sbt"
    echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | tee /etc/apt/sources.list.d/sbt.list
    echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | tee /etc/apt/sources.list.d/sbt_old.list
    curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | apt-key add -
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y sbt
  

`
