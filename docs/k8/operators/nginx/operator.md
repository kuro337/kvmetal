# Nginx Operator

0. Make sure the Operator Framework is set up `sdk/sdk.go`


```bash

mkdir nginx-operator
cd nginx-operator

# example.com just the namespace
operator-sdk init --domain=example.com --repo=github.com/example/nginx-operator

# Create API and Controller
operator-sdk create api --group webapp --version v1 --kind Nginx --resource --controller


```


