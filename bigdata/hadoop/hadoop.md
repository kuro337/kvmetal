# hadoop

setup hadoop from scratch

```bash
go run main.go --launch-vm=hadoop --mem=8192 --cpu=4 --userdata=data/userdata/shell/user-data.txt

hadoop version

df -h

# listing files in HDFS
hdfs dfs -ls /

# creating a dir in hdfs
hdfs dfs -mkdir /your-directory-name

# Copying files into hdfs
hdfs dfs -put local-file-path /hdfs-destination-path

# Getting files from hdfs to local fs
hdfs dfs -get /hdfs-file-path local-destination-path

# Running a mapreduce job

# Preparing input for MapReduce job
hdfs dfs -mkdir /input
hdfs dfs -put hadoop.sh /input

hadoop jar $HADOOP_HOME/share/hadoop/mapreduce/hadoop-mapreduce-examples-*.jar wordcount /input /output

# check output
hdfs dfs -ls /output

hdfs dfs -cat /output/part-r-00000
hdfs dfs -cat /output/part-r-00000 | sort

# k2,2 -> by 2nd value
# n    -> numeric sort instead of lexi
# r    -> specifies reverse - descending

hdfs dfs -cat /output/part-r-00000 | sort -k2,2nr

# hadoop doesnt overwrite so we need to delete the output if we run again
hdfs dfs -rm -r /output

# start yarn hdfs services
$HADOOP_HOME/sbin/start-yarn.sh
$HADOOP_HOME/sbin/start-dfs.sh
# stop yarn and hdfs services
$HADOOP_HOME/sbin/stop-yarn.sh
$HADOOP_HOME/sbin/stop-dfs.sh


# run jps to confirm ResourceManager is running
jps

# yarn UI
http://<your-vm-ip>:8088

0.0.0.0:8088

# now expose port 8088 on the VM to host port using port-forwarding

# update and change to port 8088
sudo vi /etc/ufw/before.rules 

# reload 
sudo bash /etc/libvirt/hooks/qemu
sudo ufw reload 

http://192.168.1.194:9999

sudo ufw allow 9999/tcp

sudo ufw reload

# view on laptop - should work
http://192.168.1.194:9999/cluster

```
