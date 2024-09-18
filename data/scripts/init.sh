#!/bin/bash
{
echo "Boot Script Start"
touch /home/ubuntu/boot_success.log
echo "Boot Script Ran Successfully"
echo "Boot Script End"
} >> /home/ubuntu/boot_script.log 2>&1
