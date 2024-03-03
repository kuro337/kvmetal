package bigdata

const HADOOP_UBUNTU_RUNCMD = `   # Set up environment variables
  - |
    export JAVA_HOME=$(dirname $(dirname $(readlink -f /usr/bin/java)))
    echo "export JAVA_HOME=${JAVA_HOME}" >> /etc/profile.d/java_home.sh
    java -version || exit 1

  # Scala Setup
  - wget https://downloads.lightbend.com/scala/2.13.8/scala-2.13.8.tgz -O /tmp/scala-2.13.8.tgz
  - tar -xvzf /tmp/scala-2.13.8.tgz -C /usr/local
  - echo 'export PATH=$PATH:/usr/local/scala-2.13.8/bin' | tee -a /etc/profile.d/scala_env.sh

  # sbt setup
  - echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | sudo tee /etc/apt/sources.list.d/sbt.list
  - echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | sudo tee /etc/apt/sources.list.d/sbt_old.list
  - curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | sudo apt-key add
  - sudo apt-get update
  - sudo DEBIAN_FRONTEND=noninteractive apt-get install -y sbt

  # Download and Extract Hadoop
  - mkdir -p /opt/hadoop
  - curl -o /opt/hadoop/hadoop-3.3.6.tar.gz -OL https://dlcdn.apache.org/hadoop/common/hadoop-3.3.6/hadoop-3.3.6.tar.gz
  - tar -xzf /opt/hadoop/hadoop-3.3.6.tar.gz -C /opt/hadoop
  - mv /opt/hadoop/hadoop-3.3.6 /opt/hadoop/hadoop

  # Set HADOOP_HOME and update environment variables
  - echo 'export HADOOP_HOME=/opt/hadoop/hadoop' | tee -a /etc/profile.d/hadoop_env.sh
  - echo 'export PATH=$PATH:$HADOOP_HOME/bin:$HADOOP_HOME/sbin' | tee -a /etc/profile.d/hadoop_env.sh
  - echo 'export HADOOP_MAPRED_HOME=$HADOOP_HOME' | tee -a /etc/profile.d/hadoop_env.sh
  - echo 'export HADOOP_COMMON_HOME=$HADOOP_HOME' | tee -a /etc/profile.d/hadoop_env.sh
  - echo 'export HADOOP_HDFS_HOME=$HADOOP_HOME' | tee -a /etc/profile.d/hadoop_env.sh
  - echo 'export YARN_HOME=$HADOOP_HOME' | tee -a /etc/profile.d/hadoop_env.sh
  - echo 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HADOOP_HOME/lib/native' | tee -a /etc/profile.d/hadoop_env.sh

  # Update JAVA_HOME in Hadoop's hadoop-env.sh
  - |
    JAVA_HOME_PATH=$(readlink -f /usr/bin/java | sed "s:bin/java::")
    HADOOP_ENV_SH="/opt/hadoop/hadoop/etc/hadoop/hadoop-env.sh"
    grep -q 'export JAVA_HOME=' $HADOOP_ENV_SH && sed -i "/export JAVA_HOME=/c\export JAVA_HOME=${JAVA_HOME_PATH}" $HADOOP_ENV_SH || echo "export JAVA_HOME=${JAVA_HOME_PATH}" >> $HADOOP_ENV_SH

  # Validate Hadoop installation
  - source /etc/profile && hadoop version

  # Backup original Hadoop configuration files
  - cp /opt/hadoop/hadoop/etc/hadoop/hadoop-env.sh /opt/hadoop/hadoop/etc/hadoop/hadoop-env.sh.backup
  - cp /opt/hadoop/hadoop/etc/hadoop/core-site.xml /opt/hadoop/hadoop/etc/hadoop/core-site.xml.backup
  - cp /opt/hadoop/hadoop/etc/hadoop/hdfs-site.xml /opt/hadoop/hadoop/etc/hadoop/hdfs-site.xml.backup
  - cp /opt/hadoop/hadoop/etc/hadoop/mapred-site.xml /opt/hadoop/hadoop/etc/hadoop/mapred-site.xml.backup
  - cp /opt/hadoop/hadoop/etc/hadoop/yarn-site.xml /opt/hadoop/hadoop/etc/hadoop/yarn-site.xml.backup


  # Configure core-site.xml
  - |
    echo '<configuration>
      <property>
        <name>fs.defaultFS</name>
        <value>hdfs://localhost:9000</value>
      </property>
    </configuration>' > /opt/hadoop/hadoop/etc/hadoop/core-site.xml

  # Configure hdfs-site.xml
  - |
    echo '<configuration>
      <property>
        <name>dfs.replication</name>
        <value>1</value>
      </property>
      <property>
        <name>dfs.namenode.name.dir</name>
        <value>file:///opt/hadoop_tmp/hdfs/namenode</value>
      </property>
      <property>
        <name>dfs.datanode.data.dir</name>
        <value>file:///opt/hadoop_tmp/hdfs/datanode</value>
      </property>
    </configuration>' > /opt/hadoop/hadoop/etc/hadoop/hdfs-site.xml

  # Configure mapred-site.xml
  - |
    echo '<configuration>
      <property>
        <name>mapreduce.framework.name</name>
        <value>yarn</value>
      </property>
    </configuration>' > /opt/hadoop/hadoop/etc/hadoop/mapred-site.xml

  # Configure yarn-site.xml
  - |
    echo '<configuration>
      <property>
        <name>yarn.nodemanager.aux-services</name>
        <value>mapreduce_shuffle</value>
      </property>
      <property>
        <name>yarn.nodemanager.auxservices.mapreduce.shuffle.class</name>
        <value>org.apache.hadoop.mapred.ShuffleHandler</value>
      </property>
    </configuration>' > /opt/hadoop/hadoop/etc/hadoop/yarn-site.xml

  # Create Hadoop data directories
  - mkdir -p /opt/hadoop_tmp/hdfs/namenode
  - mkdir -p /opt/hadoop_tmp/hdfs/datanode
  - chown -R ubuntu:ubuntu /opt/hadoop_tmp

    # SSH setup for Hadoop
  - sudo -u ubuntu ssh-keygen -t rsa -P '' -f ~ubuntu/.ssh/id_rsa
  - cat ~ubuntu/.ssh/id_rsa.pub >> ~ubuntu/.ssh/authorized_keys
  - chown ubuntu:ubuntu ~ubuntu/.ssh/id_rsa*
  - chown ubuntu:ubuntu ~ubuntu/.ssh/authorized_keys

  # Initialize HDFS and start Hadoop services
  - sudo -u ubuntu /opt/hadoop/hadoop/bin/hdfs namenode -format
  - sudo -u ubuntu /opt/hadoop/hadoop/sbin/start-dfs.sh
  - sudo -u ubuntu /opt/hadoop/hadoop/sbin/start-yarn.sh

  # Validate HDFS setup
  - sudo -u ubuntu /opt/hadoop/hadoop/bin/hdfs dfs -mkdir /test
  - sudo -u ubuntu /opt/hadoop/hadoop/bin/hdfs dfs -ls /

  - echo "source /etc/profile.d/hadoop_env.sh" >> ~ubuntu/.zshrc
  - echo "source /etc/profile.d/hadoop_env.sh" >> ~ubuntu/.bashrc


final_message: "Hadoop is up, after $UPTIME seconds"
`

const Hadoop_Literal_cloudinit_full = `#cloud-config

users:
  - name: ubuntu
    shell: /usr/bin/zsh
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    groups: sudo
    passwd: password
    lock_passwd: false

package_upgrade: true

packages:
  - zsh
  - openjdk-11-jdk
  - git
  - curl

runcmd:
  # Set up environment variables
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
    # Step 3. sbt
    echo -e "Installing and Setting up sbt"
    echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | tee /etc/apt/sources.list.d/sbt.list
    echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | tee /etc/apt/sources.list.d/sbt_old.list
    curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | apt-key add -
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y sbt
 
 # Setup Hadoop
  - |
    # Step 4. Hadoop Setup
    echo -e "Setting up Hadoop"
    curl -OL https://dlcdn.apache.org/hadoop/common/hadoop-3.3.6/hadoop-3.3.6.tar.gz
    tar xvf hadoop-3.3.6.tar.gz -C $HOME_DIR
    chown -R ubuntu:ubuntu $HOME_DIR/hadoop-3.3.6
    HADOOP_HOME="$HOME_DIR/hadoop-3.3.6"
    echo "export HADOOP_HOME=$HADOOP_HOME" >> $ENV_SHELL
    echo "export PATH=\$PATH:\$HADOOP_HOME/bin:\$HADOOP_HOME/sbin" >> $ENV_SHELL
    echo "export HADOOP_MAPRED_HOME=\$HADOOP_HOME" >> $ENV_SHELL
    echo "export HADOOP_COMMON_HOME=\$HADOOP_HOME" >> $ENV_SHELL
    echo "export HADOOP_HDFS_HOME=\$HADOOP_HOME" >> $ENV_SHELL
    echo "export YARN_HOME=\$HADOOP_HOME" >> $ENV_SHELL
    echo "export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:\$HADOOP_HOME/lib/native" >> $ENV_SHELL

  # Copying JAVA_HOME settings to Hadoop 
  - |
    echo -e "Setting JAVA_HOME for Hadoop by modifying $HADOOP_HOME/etc/hadoop/hadoop-env.sh Config File"
    cp $HADOOP_HOME/etc/hadoop/hadoop-env.sh $HADOOP_HOME/etc/hadoop/hadoop-env-backup.sh
    new_java_home="export JAVA_HOME=$JAVA_HOME"
    line_number=$(grep -n JAVA_HOME "$HADOOP_HOME/etc/hadoop/hadoop-env.sh" | head -n 1 | cut -d ":" -f 1)
    sed -i "${line_number}s|.*|${new_java_home}|" "$HADOOP_HOME/etc/hadoop/hadoop-env.sh"

  # Create Copies of Hadoop Conf Files before Modifying
  - cp $HADOOP_HOME/etc/hadoop/yarn-site.xml $HADOOP_HOME/etc/hadoop/yarn-site-backup.xml
  - cp $HADOOP_HOME/etc/hadoop/mapred-site.xml $HADOOP_HOME/etc/hadoop/mapred-site-backup.xml
  - cp $HADOOP_HOME/etc/hadoop/hdfs-site.xml $HADOOP_HOME/etc/hadoop/hdfs-site-backup.xml  
  - cp $HADOOP_HOME/etc/hadoop/core-site.xml  $HADOOP_HOME/etc/hadoop/core-site-backup.xml
  - cp $HADOOP_HOME/etc/hadoop/hadoop-env.sh $HADOOP_HOME/etc/hadoop/hadoop-env-backup.sh

  # Modify hadoop-env.sh to set JAVA_HOME
  - sed -i 's|export JAVA_HOME=.*|export JAVA_HOME='$(readlink -f /usr/bin/java | sed "s:bin/java::")'|' $HOME_DIR/hadoop-3.3.6/etc/hadoop/hadoop-env.sh
  
  # Update Hadoop Files
  - echo '<configuration>
    <property>
      <name>fs.defaultFS</name>
      <value>hdfs://localhost:9000</value>
    </property>
    </configuration>' > $HOME_DIR/hadoop-3.3.6/etc/hadoop/core-site.xml
  
  # Configure hdfs-site.xml
  - echo '<configuration>
    <property>
      <name>dfs.replication</name>
      <value>1</value>
    </property>
    <property>
      <name>dfs.namenode.name.dir</name>
      <value>file:///opt/hadoop_tmp/hdfs/namenode</value>
    </property>
    <property>
      <name>dfs.datanode.data.dir</name>
      <value>file:///opt/hadoop_tmp/hdfs/datanode</value>
    </property>
    </configuration>' > $HOME_DIR/hadoop-3.3.6/etc/hadoop/hdfs-site.xml
  
  # Create directories for Hadoop data storage
  - mkdir -p /opt/hadoop_tmp/hdfs/namenode
  - mkdir -p /opt/hadoop_tmp/hdfs/datanode
  - chown -R ubuntu:ubuntu /opt/hadoop_tmp
  
  # Configure mapred-site.xml
  - echo '<configuration>
    <property>
      <name>mapreduce.framework.name</name>
      <value>yarn</value>
    </property>
    </configuration>' > $HOME_DIR/hadoop-3.3.6/etc/hadoop/mapred-site.xml
  
  # Configure yarn-site.xml
  - echo '<configuration>
    <property>
      <name>yarn.nodemanager.aux-services</name>
      <value>mapreduce_shuffle</value>
    </property>
    <property>
      <name>yarn.nodemanager.auxservices.mapreduce.shuffle.class</name>
      <value>org.apache.hadoop.mapred.ShuffleHandler</value>
    </property>
    </configuration>' > $HOME_DIR/hadoop-3.3.6/etc/hadoop/yarn-site.xml

  # Validate ssh 
  - sudo -u ubuntu ssh-keygen -q -t rsa -N '' -f $HOME_DIR/.ssh/id_rsa <<<y >/dev/null 2>&1
  - cat $HOME_DIR/.ssh/id_rsa.pub >> $HOME_DIR/.ssh/authorized_keys
  - chown ubuntu:ubuntu $HOME_DIR/.ssh/id_rsa*
  - chown ubuntu:ubuntu $HOME_DIR/.ssh/authorized_keys
  - chown ubuntu:ubuntu -R /home/ubuntu/

  
  # Initialize HDFS and start Hadoop services
  - su - ubuntu -c '$HOME_DIR/hadoop-3.3.6/bin/hdfs namenode -format'
  - su - ubuntu -c '$HOME_DIR/hadoop-3.3.6/sbin/start-dfs.sh'
  - su - ubuntu -c '$HOME_DIR/hadoop-3.3.6/sbin/start-yarn.sh'

  # Validate HDFS setup
  - su - ubuntu -c '$HOME_DIR/hadoop-3.3.6/bin/hdfs dfs -mkdir /test'
  - su - ubuntu -c '$HOME_DIR/hadoop-3.3.6/bin/hdfs dfs -ls /'
  

final_message: "The system is finally up, after $UPTIME seconds"
`

const Hadoop_Literal_jdk11_scala_direct_script = `#!/bin/bash

if [[ "$SHELL" == */zsh ]]; then
    ENV_SHELL="$HOME/.zshrc"
    echo -e "Using zsh"
elif [[ "$SHELL" == */bash ]]; then
    echo -e "Using bash"
    ENV_SHELL="$HOME/.bashrc"
else
    echo "Unsupported shell. Defaulting to .bashrc"
    ENV_SHELL="$HOME/.bashrc"
fi


###### Step 1. JDK
echo -e "Setting up @OpenJDK11"
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y openjdk-11-jdk
java -version

echo -e "Finding JAVA_HOME path"
JAVA_HOME_DIR=$(dirname $(dirname $(readlink -f /usr/bin/java)))
echo -e "Setting JAVA_HOME to $JAVA_HOME_DIR"
if ! grep -q "export JAVA_HOME=$JAVA_HOME_DIR" $ENV_SHELL; then
    echo "export JAVA_HOME=$JAVA_HOME_DIR" >> $ENV_SHELL
fi
source $ENV_SHELL

echo 'public class Hello {
    public static void main(String[] args) {
        int shouldWork = 10 + 10;
        if (shouldWork == 20) {
          System.out.println("Java Installation Successful!");
        } else {
          System.out.println("This wont run anyway. Whoop Whoop");
        }
    }
}' >> Hello.java && javac Hello.java && java Hello && rm -rf Hello.class && rm -rf Hello.java

###### Step 2. Scala
echo -e "Setting up Scala 2.13"

wget https://downloads.lightbend.com/scala/2.13.8/scala-2.13.8.tgz
tar -xvzf scala-2.13.8.tgz
sudo mv scala-2.13.8 /usr/local/scala
echo 'export PATH=$PATH:/usr/local/scala/bin' >> $ENV_SHELL
source $ENV_SHELL

###### Step 3. sbt
echo -e "Installing and Setting up sbt"

echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | sudo tee /etc/apt/sources.list.d/sbt.list
echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | sudo tee /etc/apt/sources.list.d/sbt_old.list
curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | sudo apt-key add
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y sbt
sbt sbtVersion

###### Step 4. Hadoop
echo -e "Setting up Hadoop"

curl -OL https://dlcdn.apache.org/hadoop/common/hadoop-3.3.6/hadoop-3.3.6.tar.gz
tar xvf hadoop-3.3.6.tar.gz

HADOOP_HOME="$HOME/hadoop-3.3.6"

ENV_CONFIG="export HADOOP_HOME=$HADOOP_HOME
export PATH=\$PATH:\$HADOOP_HOME/bin:\$HADOOP_HOME/sbin
export HADOOP_MAPRED_HOME=\$HADOOP_HOME
export HADOOP_COMMON_HOME=\$HADOOP_HOME
export HADOOP_HDFS_HOME=\$HADOOP_HOME
export YARN_HOME=\$HADOOP_HOME
export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:\$HADOOP_HOME/lib/native"
echo "$ENV_CONFIG" >> $ENV_SHELL
source $ENV_SHELL

# source $HOME/.bashrc 

# echo 'export HADOOP_HOME=/home/ubuntu/hadoop-3.3.6
# export PATH=$PATH:$HADOOP_HOME/bin:$HADOOP_HOME/sbin
# export HADOOP_MAPRED_HOME=$HADOOP_HOME
# export HADOOP_COMMON_HOME=$HADOOP_HOME
# export HADOOP_HDFS_HOME=$HADOOP_HOME
# export YARN_HOME=$HADOOP_HOME
# export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HADOOP_HOME/lib/native' >> ~/.bashrc

hadoop version

echo -e "Setting JAVA_HOME for Hadoop by modifying $HADOOP_HOME/etc/hadoop/hadoop-env.sh Config File"

cp $HADOOP_HOME/etc/hadoop/hadoop-env.sh $HADOOP_HOME/etc/hadoop/hadoop-env-backup.sh
new_java_home="export JAVA_HOME=$JAVA_HOME"
line_number=$(grep -n JAVA_HOME "$HADOOP_HOME/etc/hadoop/hadoop-env.sh" | head -n 1 | cut -d ":" -f 1)
sed -i "${line_number}s|.*|${new_java_home}|" "$HADOOP_HOME/etc/hadoop/hadoop-env.sh"

cp $HADOOP_HOME/etc/hadoop/core-site.xml  $HADOOP_HOME/etc/hadoop/core-site-backup.xml

echo '<configuration>
  <property>
    <name>fs.defaultFS</name>
    <value>hdfs://localhost:9000</value>
  </property>
</configuration>' > $HADOOP_HOME/etc/hadoop/core-site.xml

cp $HADOOP_HOME/etc/hadoop/hdfs-site.xml $HADOOP_HOME/etc/hadoop/hdfs-site-backup.xml


echo '<configuration>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>file:///opt/hadoop_tmp/hdfs/datanode</value>
  </property>
  <property>
    <name>dfs.namenode.name.dir</name>
    <value>file:///opt/hadoop_tmp/hdfs/namenode</value>
  </property>
  <property>
    <name>dfs.replication</name>
    <value>1</value>
  </property>
</configuration>' > $HADOOP_HOME/etc/hadoop/hdfs-site.xml

sudo mkdir -p /opt/hadoop_tmp/hdfs/datanode
sudo mkdir -p /opt/hadoop_tmp/hdfs/namenode
sudo chown ubuntu:ubuntu -R /opt/hadoop_tmp

cp $HADOOP_HOME/etc/hadoop/mapred-site.xml $HADOOP_HOME/etc/hadoop/mapred-site-backup.xml

echo '<configuration>
  <property>
    <name>mapreduce.framework.name</name>
    <value>yarn</value>
  </property>
</configuration>' $HADOOP_HOME/etc/hadoop/mapred-site.xml

cp $HADOOP_HOME/etc/hadoop/yarn-site.xml $HADOOP_HOME/etc/hadoop/yarn-site-backup.xml

echo '<configuration>
  <property>
    <name>yarn.nodemanager.aux-services</name>
    <value>mapreduce_shuffle</value>
  </property>
  <property>
    <name>yarn.nodemanager.auxservices.mapreduce.shuffle.class</name>
    <value>org.apache.hadoop.mapred.ShuffleHandler</value>
  </property>
</configuration> ' $HADOOP_HOME/etc/hadoop/yarn-site.xml

echo -e "Confirm ssh is enabled on Machine. This can be tested by running ssh localhost and then exit."
which sshd
# ssh localhost
# ssh-keygen
# nonblocking keygen
ssh-keygen -q -t rsa -N '' -f ~/.ssh/id_rsa <<<y >/dev/null 2>&1 

cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys

echo -e "Hadoop Configurations Updated. Formatting and Booting HDFS"
hdfs namenode -format -force
start-dfs.sh && start-yarn.sh
jps

echo -e "Confirm output of jps contains - DataNode and NameNode"

# 15425 Jps
# 13843 ResourceManager
# 14500 NameNode
# 14900 SecondaryNameNode
# 14677 DataNode
# 15229 NodeManager

echo -e "Validate HDFS"

hdfs dfs -mkdir /test
hdfs dfs -ls /

echo -e "Expected output should be ... ubuntu supergroup 0 2024-02-13 07:06 /test"
echo -e "To stop hdfs and yarn run stop-dfs.sh and stop-yarn.sh"


echo -e "Setup Completed Successfully!"

echo -e "Confirm by Running spark-shell and the Hadoop NativeLoader warning should be Gone"

# ./bin/spark-shell
# spark.range(1000 * 1000 * 1000).count()`
