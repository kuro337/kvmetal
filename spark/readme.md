# spark kube

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install my-spark bitnami/spark
# Save output of helm install

helm uninstall my-spark bitnami/spark

# Run Sample App

 export EXAMPLE_JAR=$(kubectl exec -ti --namespace default my-spark-worker-0 -- find examples/jars/ -name 'spark-example*\.jar' | tr -d '\r')

 kubectl exec -ti --namespace default my-spark-worker-0 -- spark-submit --master spark://my-spark-master-svc:7077 \
    --class org.apache.spark.examples.SparkPi \
    $EXAMPLE_JAR 5

# Pi is roughly 3.141550283100566
```

#### Kick off Custom Jobs

```bash
# 1. Write App and create .jar file

# Create Folder on Worker
kubectl exec my-spark-worker-0 -- mkdir -p /opt/spark-apps/

# Stage .jar
kubectl cp MySparkApp-assembly-0.1.jar default/my-spark-worker-0:/opt/spark-apps/MySparkApp-assembly-0.1.jar

# Run Job - should work

 kubectl exec -ti --namespace default my-spark-worker-0 -- spark-submit --master spark://my-spark-master-svc:7077 \
    --class MySparkApp  \
    /opt/spark-apps/MySparkApp-assembly-0.1.jar

```
