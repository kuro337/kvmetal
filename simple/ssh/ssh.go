package ssh

import (
	"bytes"
	"fmt"
	"kvmgo/constants"
	"kvmgo/lib"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func extractJoinCommand(input string) (string, error) {
	// Regular expression to match the join command
	re := regexp.MustCompile(`kubeadm join \S+:\d+ --token \S+ \\\s+--discovery-token-ca-cert-hash \S+`)

	// Find the match
	match := re.FindString(input)
	if match == "" {
		return "", fmt.Errorf("join command not found in the input string")
	}

	// Clean up the extracted command
	joinCmd := strings.ReplaceAll(match, "\\\n", " ")
	joinCmd = strings.ReplaceAll(joinCmd, "\t", "")

	return joinCmd, nil
}

func GetJoinCmd(domain string) string {
	privateKeyPath := constants.SshPriv

	if domain == "" {
		domain = "ubuntu-base-vm"
	}

	qconn, _ := lib.ConnectLibvirt()
	dom, _ := qconn.GetDomain(domain)
	vmIP, _ := dom.GetIP()

	// Read the private key file
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}

	// Parse the private key file
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}

	// Configure the SSH client
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", net.JoinHostPort(vmIP, "22"), config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Start a session
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Run a simple command
	var b bytes.Buffer
	session.Stdout = &b

	// cmd := "ls"
	cmd := "cat kubeadm-init.log"
	//    cmd := "echo 'SSH connection successful'"

	if err := session.Run(cmd); err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}
	e, err := extractJoinCommand(b.String())
	if err != nil {
		log.Fatalf("failure: %s", err)
	}
	return e
}

func TestSSHConnection(t *testing.T) {
	privateKeyPath := constants.SshPriv
	domain := "ubuntu-base-vm"
	qconn, _ := lib.ConnectLibvirt()
	dom, _ := qconn.GetDomain(domain)
	vmIP, _ := dom.GetIP()

	// Read the private key file
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Unable to read private key: %v", err)
	}

	// Parse the private key file
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		t.Fatalf("Unable to parse private key: %v", err)
	}

	// Configure the SSH client
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", net.JoinHostPort(vmIP, "22"), config)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Start a session
	session, err := conn.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Run a simple command
	var b bytes.Buffer
	session.Stdout = &b

	// cmd := "ls"
	cmd := "cat kubeadm-init.log"
	//    cmd := "echo 'SSH connection successful'"

	if err := session.Run(cmd); err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	log.Println(b.String())
}
