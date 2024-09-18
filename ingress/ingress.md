# ingress using traefik

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik

helm install traefik traefik/traefik -n kube-system --debug

helm install traefik traefik/traefik --set service.type=NodePort --namespace kube-system

helm uninstall traefik -n kube-system

```

#### quick http ingress route using IP of host

```bash
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: my-app-route
  namespace: default
spec:
  entryPoints:
    - web
  routes:
  - match: HostRegexp(`{host:.+}`)
    kind: Rule
    services:
    - name: my-app-service
      port: 80


```

```bash

helm install traefik traefik/traefik --set service.type=NodePort --namespace kube-system

kubectl apply -f ingress.yaml

kubectl get svc -n kube-system | grep traefik

kubectl create deployment nginx --image=nginx
kubectl expose deployment nginx --port=80 --name=my-app-service



# get external ip of traefic service
kubectl get svc -n kube-system


curl http://<External-IP>

kubectl delete svc my-app-service
kubectl delete deployment nginx


```
