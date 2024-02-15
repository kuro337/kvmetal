#!/bin/bash

#
# This script will set kubectl usable by the user+password login for the VM

# normal user will be able to run kubectl
#

USER_HOME="/home/ubuntu"


# Create the .kube directory if it doesn't exist
sudo mkdir -p $USER_HOME/.kube

# Copy the kubeconfig file from the root user's home to the normal user's .kube directory
sudo cp /root/.kube/config $USER_HOME/.kube/config

# Change the ownership of the .kube directory and config file to the 'ubuntu' user
sudo chown -R ubuntu:ubuntu $USER_HOME/.kube

# Change the file permissions so that it's readable by the user
sudo chmod 644 $USER_HOME/.kube/config
