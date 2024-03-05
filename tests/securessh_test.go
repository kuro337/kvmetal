package tests

import (
	"bytes"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"kvmgo/constants"
	"kvmgo/lib"

	"golang.org/x/crypto/ssh"
)

func TestSSHConnection(t *testing.T) {
	privateKeyPath := constants.SshPriv
	domain := "test"
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
	if err := session.Run("echo 'SSH connection successful'"); err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	log.Println(b.String())
}
