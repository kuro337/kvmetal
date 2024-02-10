# Use an official Spark base image
FROM bitnami/spark:3

# Copy your application JAR to the image
COPY MySparkApp-assembly-0.1.jar /opt/spark-apps/MySparkApp-assembly-0.1.jar

# Specify the default command for the container
CMD ["/opt/spark/bin/spark-submit", "--class", "MySparkApp", "--master", "k8s://https://kubeinfoip:port", "/opt/spark-apps/MySparkApp-assembly-0.1.jar"]

# kubectl cluster-info
# get Control Plane IP
# Kubernetes control plane is running at https://ip:port


# docker build -t my-spark-app .
# docker tag my-spark-app:latest ghcr.io/xxx/my-spark-app:latest
# docker push ghcr.io/xxx/my-spark-app:latest


# spark-submit \
#     --master k8s://https://ipfromkubeinfo:port \
#     --deploy-mode cluster \
#     --name my-spark-app \
#     --class MySparkApp \
#     --conf spark.executor.instances=2 \
#     --conf spark.kubernetes.container.image=ghcr.io/xxx/my-spark-app:latest \
#     --conf spark.kubernetes.namespace=default \
#     --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark
