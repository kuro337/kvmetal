# simple metallb usage

Context

```bash
Running Virtual Machines on Ubuntu using libvirt https://libvirt.org/

uses KVM & qemu - and (i believe) the network for Virtual Machines uses a NAT to share hosts network

```

Good Docs here

Note: Official Docs is the only source of Truth - all online articles are obsolete

https://github.com/metallb/metallb/blob/v0.14.3/configsamples/ipaddresspool_simple.yaml

#### steps to run it

Install Metallb to the Control Node

Apply Config to give Metallb an IP Pool to use

Deploy nginx

`kubectl get svc` - to get the External IP MetalLB is using

and then curl that address - done

```bash

step 1: install metallb

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.14.3/config/manifests/metallb-native.yaml


Step 2: get ips to allocate to metallb

# specify available ip range by running ip addr show and getting ips unused

# ip addr show
# ping -c 1 191.143.1.100 , if returns failure - we will use it for MetalLB

kubectl apply -f metallb-config.yaml


apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: first-pool
  namespace: metallb-system
spec:
  addresses:
  - 191.148.1.100-191.148.1.104


step 3: deploy and test deployment

sudo vi nginx.yaml # :set paste  and paste , then :set nopaste to return to indent mode

apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1
        ports:
        - name: http
          containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer

```

#### validate

```bash
kubectl get svc

# get external ip col of nginx
curl external::ip

```
