package deployments

/* Standard Nginx Deployment - must be accompanied by a Service */
const NGINX = `
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
`

/* Exposes the Deployment on the Node's Port - so we can access it by the Node's IP it is deployed on without creating a Service. */
const NGINX_EXTERNAL_EXPOSED = `
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
`
