package network

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

/*
go get golang.org/x/crypto/ssh

WORKER_IP=$(sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml worker | awk -F"'" '/mac address/{print $2}') | awk '{print $1}')

Make sure SSH is active on Node
systemctl status ssh

sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml kubecontrol | awk -F"'" '/mac address/{print $2}') | awk '{print $1}'

sudo apt-get install openssh-server
*/

/*
Using arp-scan to get the IP of a VM

Usage:

	ip , err := GetVMIPAddr("kubecontrol")
	log.Printf("IP of Control Node is %s",ip)
*/
func GetVMIPAddr(vmName string) (string, error) {
	cmdString := fmt.Sprintf("sudo arp-scan --interface=virbr0 --localnet | grep -f <(virsh dumpxml %s | awk -F\"'\" '/mac address/{print $2}') | awk '{print $1}'", vmName)

	cmd := exec.Command("bash", "-c", cmdString)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(&out)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}

	return "", fmt.Errorf("no IP address found for VM: %s", vmName)
}
