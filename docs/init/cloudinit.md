# cloud init

Cloud init defines the Metadata such as username/pass and Boot Scripts to be defined for the Machine

This is typically attached as a raw disk instead of a qCow2 disk - which has to be detached during Snapshots

Basic user-data.txt file that specifies username/password for the Machine

```bash
#cloud-config
password: password
chpasswd: { expire: False }
ssh_pwauth: True


```

Config that defines multiple scripts to run during launch time. Note that any scripts are ran as the Root user - so we need to define commands such as `sudo -u <user> dothis` if we need to run as the default login user

```bash
#cloud-config

write_files:
  - path: /root/first_script.sh
    permissions: '0755'
    content: |
      #!/bin/bash
      echo "This is the first script." > /tmp/first_script_output.txt
  - path: /root/second_script.sh
    permissions: '0755'
    content: |
      #!/bin/bash
      echo "This is the second script." > /tmp/second_script_output.txt

runcmd:
  - /root/first_script.sh
  - /root/second_script.sh


```
