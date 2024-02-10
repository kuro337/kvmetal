package network

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
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

func SSHInsecure(ip, username, password, command string) (string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), config)
	if err != nil {
		return "", fmt.Errorf("failed to dial: %v", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %v", err)
	}

	return string(output), nil
}

func SSH(ip, username, privateKeyPath, command string) (string, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), config)
	if err != nil {
		return "", fmt.Errorf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Create a session
	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Run the command
	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %v", err)
	}

	return string(output), nil
}
