package bigdata

const SPARK_UBUNTU_RUNCMD = `# Setup Spark 
- echo -e "Downloading Spark" 
- wget https://dlcdn.apache.org/spark/spark-3.5.0/spark-3.5.0-bin-hadoop3-scala2.13.tgz
- echo -e "Decompressing Release Download"
- tar xvf spark-3.5.0-bin-hadoop3-scala2.13.tgz
`
