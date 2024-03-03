package kube

const CALICO_LINUX_RUNCMD = `
  
  - kubectl --kubeconfig=/home/ubuntu/.kube/config apply -f https://docs.projectcalico.org/manifests/calico.yaml

`
