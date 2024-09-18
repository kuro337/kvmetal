# Setting up Spark, Hadoop, Java, Scala, and SBT on Bare Metal

Super helpful : https://dev.to/awwsmm/installing-and-running-hadoop-and-spark-on-ubuntu-18-393h

```bash
#!/bin/bash

###### Step 0. Setup and Upgrade System Packages
echo -e "Updating Packages"
sudo apt update && sudo apt upgrade -y

###### Step 1. JDK
echo -e "Setting up @OpenJDK11"
sudo apt update
sudo apt install openjdk-11-jdk
java -version

echo -e "Finding JAVA_HOME path"
JAVA_HOME_DIR=$(dirname $(dirname $(readlink -f /usr/bin/java)))
echo -e "Setting JAVA_HOME to $JAVA_HOME_DIR"
if ! grep -q "export JAVA_HOME=$JAVA_HOME_DIR" ~/.bashrc; then
    echo "export JAVA_HOME=$JAVA_HOME_DIR" >> ~/.bashrc
fi
source ~/.bashrc

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
echo 'export PATH=$PATH:/usr/local/scala/bin' >> ~/.bashrc
source ~/.bashrc

###### Step 3. sbt
echo -e "Installing and Setting up sbt"

echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | sudo tee /etc/apt/sources.list.d/sbt.list
echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | sudo tee /etc/apt/sources.list.d/sbt_old.list
curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | sudo apt-key add
sudo apt-get update
sudo apt-get install sbt
sbt sbtVersion

###### Step 4. Spark
echo -e "Downloading Spark"
curl -OL https://dlcdn.apache.org/spark/spark-3.5.0/spark-3.5.0-bin-hadoop3-scala2.13.tgz
echo -e "Decompressing Release Download"
tar xvf spark-3.5.0-bin-hadoop3-scala2.13.tgz

source ~/.bashrc

echo 'import org.apache.spark.sql.SparkSession

object SparkScript {
    def main(args: Array[String]): Unit = {
        val master = "local[*]" // Use local mode with all available cores
        val spark = SparkSession.builder()
            .appName("Spark Script")
            .master(master)
            .getOrCreate()

        val count = spark.range(1000 * 1000 * 1000).count()
        println(s"Count: $count")

        spark.stop()
    }
}' >> spark_script.scala

echo -e "Validating Spark and Scala installation"

scala -J-Xmx8g -classpath "spark-3.5.0-bin-hadoop3-scala2.13/jars/*" spark_script.scala
rm -rf spark_script.scala

echo -e "Spark Configured Successfully and Health Check Passed"

echo -e "We can run Spark in shell-mode using ./bin/spark-shell and then run spark.range(1000 * 1000 * 1000).count()"
echo -e "Setting up Hadoop"

###### Step 5. Hadoop
echo -e "Setting up Hadoop"

curl -OL https://dlcdn.apache.org/hadoop/common/hadoop-3.3.6/hadoop-3.3.6.tar.gz
tar xvf hadoop-3.3.6.tar.gz

echo 'export HADOOP_HOME=/home/ubuntu/hadoop-3.3.6
export PATH=$PATH:$HADOOP_HOME/bin:$HADOOP_HOME/sbin
export HADOOP_MAPRED_HOME=$HADOOP_HOME
export HADOOP_COMMON_HOME=$HADOOP_HOME
export HADOOP_HDFS_HOME=$HADOOP_HOME
export YARN_HOME=$HADOOP_HOME
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HADOOP_HOME/lib/native' >> ~/.bashrc

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
ssh-keygen
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
# spark.range(1000 * 1000 * 1000).count()
```
