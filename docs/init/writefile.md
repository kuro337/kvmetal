# writefile for async runcmd

1. Use Cloud-Init for Late Binding of IP Addresses

```bash


#cloud-config
write_files:
  - path: /etc/myapp/configure.sh
    permissions: '0755'
    content: |
      #!/bin/bash
      IP_ADDR=$(hostname -I | cut -d' ' -f1)
      echo "Configuring application with IP $IP_ADDR"
      # Your configuration commands here, e.g., sed to insert $IP_ADDR into config files

runcmd:
  - /etc/myapp/configure.sh

```
