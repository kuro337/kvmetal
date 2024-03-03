# cilium

Prerequisites

```bash
# On machine
Linux Kernel : >= 5.4

# these are handled within k8
etcd         : >= 3.1.0
Clang/llvm   : >= 10.0

# check kernel
uname -r     : 5.15.0-92-generic
```

```bash

helm install cilium cilium/cilium --version 1.15.1 --namespace kube-system
helm uninstall cilium cilium/cilium --namespace kube-system

CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
CLI_ARCH=amd64
if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi
curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
sudo tar xzvfC cilium-linux-${CLI_ARCH}.tar.gz /usr/local/bin
rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
# Validate version v1.15.1
cilium version --client

# Install Cilium
git clone https://github.com/cilium/cilium.git
cd cilium
cilium install --chart-directory ./install/kubernetes/cilium
cilium status --wait

# Enable Hubble
cilium hubble enable

# Install Hubble CLI
HUBBLE_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/hubble/master/stable.txt)
HUBBLE_ARCH=amd64
if [ "$(uname -m)" = "aarch64" ]; then HUBBLE_ARCH=arm64; fi
curl -L --fail --remote-name-all https://github.com/cilium/hubble/releases/download/$HUBBLE_VERSION/hubble-linux-${HUBBLE_ARCH}.tar.gz{,.sha256sum}
sha256sum --check hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum
sudo tar xzvfC hubble-linux-${HUBBLE_ARCH}.tar.gz /usr/local/bin
rm hubble-linux-${HUBBLE_ARCH}.tar.gz{,.sha256sum}


cilium hubble port-forward&
hubble status

```

## Setting up Cilium to Handle Ingress and Networking

```bash
hostname -I | awk '{print $1}'

# Specifying an IP and Port for Cilium Internal Networking
# 192.168.1.10 Host IP
# 9000 : Port

cilium install --chart-directory ./install/kubernetes/cilium \
--set=kubeProxyReplacement=true \
--set=k8sServiceHost=${API_SERVER_IP} \
--set=k8sServicePort=${API_SERVER_PORT}
cilium status --wait

# Ensure each Worker has an InternalIP which is assigned with the same Name on Each Node
kubectl get nodes -o wide

kubectl -n kube-system delete ds kube-proxy # these not required unless we have existing nodes
kubectl -n kube-system delete cm kube-proxy
iptables-save | grep -v KUBE | iptables-restore # Run on each node with root permissions:


cilium install --chart-directory ./install/kubernetes/cilium \
--set=kubeProxyReplacement=true \
--set=k8sServiceHost=192.168.1.10 \
--set=k8sServicePort=9000
cilium status --wait


# Validate
kubectl -n kube-system exec ds/cilium -- cilium-dbg status --verbose
```

## Validations for Cilium KubeProxy Replacement Deploy

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx
spec:
  selector:
    matchLabels:
      run: my-nginx
  replicas: 2
  template:
    metadata:
      labels:
        run: my-nginx
    spec:
      containers:
        - name: my-nginx
          image: nginx
          ports:
            - containerPort: 80
```

```bash

kubectl exec -it -n kube-system cilium-fmh8d -- cilium-dbg service list

kubectl apply -f nginx.yaml
kubectl get pods -l run=my-nginx -o wide # should see 2 Running

# expose
kubectl expose deployment my-nginx --type=NodePort --port=80
kubectl get svc my-nginx

# We should see IP's and Ports here - for the Deployment
kubectl -n kube-system exec ds/cilium -- cilium-dbg service list

# capture the Node port
node_port=$(kubectl get svc my-nginx -o=jsonpath='{@.spec.ports[0].nodePort}')

# confirm no IPTables Rule is present
sudo iptables-save | grep KUBE-SVC

# confirm services are reachable
curl 127.0.0.1:$node_port
curl 192.168.122.163:$node_port
curl 0.0.0.0:$node_port

```

## Deploying by Exposing the App on the Node Port Directly

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-nginx-ext
spec:
  selector:
    matchLabels:
      run: my-nginx-ext
  replicas: 1
  template:
    metadata:
      labels:
        run: my-nginx-ext
    spec:
      containers:
        - name: my-nginx-ext
          image: nginx
          ports:
            - containerPort: 80
              hostPort: 8088
```

```bash
# check if hostport is enabled
kubectl -n kube-system exec ds/cilium -- cilium-dbg status --verbose | grep HostPort
# if enabled we can deploy
kubectl apply -f nginx_ext.yaml
get nodes -o wide
# Curl the Node IP with the Port we specified
curl 192.168.122.135:8088
```

## Troubleshooting Failures

```bash
# Troubleshooting Components that fail
kubectl -n kube-system get deployment hubble-relay
kubectl -n kube-system describe deployment hubble-relay
kubectl -n kube-system get pods -l k8s-app=hubble-relay
kubectl -n kube-system describe pod hubble-relay-5cff74fdb4-xpx7q


```
