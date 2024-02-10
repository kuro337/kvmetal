package network

/*

These are marked Unsafe so SSH is preferable!


Meant to be run from guest VM

sudo apt-get install qemu-guest-agent

systemctl start qemu-guest-agent
systemctl enable qemu-guest-agent

sudo virsh -c qemu:///system qemu-agent-command kubecontrol \
  '{"execute": "guest-exec", "arguments": { "path": "/usr/bin/ls", "arg": [ "/" ], "capture-output": true }}'

{"return":{"pid":14925}}


virsh -c qemu:///system qemu-agent-command kubecontrol \
  '{"execute": "guest-exec-status", "arguments": { "pid": 14925 }}'

will return {"return":{"exitcode":0,"out-data":"YmluCmJvb3QKZGVhZC5sZXR0ZXIKZGV2CmV0Ywpob21lCmxpYgpsaWI2NApsb3N0K2ZvdW5kCm1lZGlhCm1udApvcHQKcHJvYwpyb290CnJ1bgpzYmluCnNlbGludXgKc3J2CnN5cwp0bXAKdXNyCnZhcgo=","exited":true}}

base64 decode the out-data

> command output decode


*/
