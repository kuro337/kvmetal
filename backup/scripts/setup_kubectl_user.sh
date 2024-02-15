#!/bin/bash

# echo "New Script Running - comment in below and use _original.sh config to restore to 1st try"

#
# This script will set kubectl usable by the user+password login for the VM

# normal user will be able to run kubectl
#

USER_HOME="/home/ubuntu"
KUBEADM_INIT_LOG="/tmp/kubeadm-init.log"

sudo cp $KUBEADM_INIT_LOG $USER_HOME/kubeadm-init.log
sudo chown ubuntu:ubuntu $USER_HOME/kubeadm-init.log

# Create the .kube directory if it doesn't exist
sudo mkdir -p $USER_HOME/.kube

# Copy the kubeconfig file from the root user's home to the normal user's .kube directory
sudo cp /root/.kube/config $USER_HOME/.kube/config

# Change the ownership of the .kube directory and config file to the 'ubuntu' user
sudo chown -R ubuntu:ubuntu $USER_HOME/.kube

# Change the file permissions so that it's readable by the user
sudo chmod 644 $USER_HOME/.kube/config

# these are new steps added

# Setup Helm for ubuntu user
echo "-----> Setting up Helm for user ubuntu..."

sudo mkdir -p $USER_HOME/.config/helm
sudo chown -R ubuntu:ubuntu $USER_HOME/.config

