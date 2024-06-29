package ssh

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"kvmgo/constants"
	"kvmgo/lib"

	"golang.org/x/crypto/ssh"
)

func EstablishSsh(domain string) (*ssh.Session, error) {
	privateKeyPath := constants.SshPriv
	qconn, _ := lib.ConnectLibvirt()
	dom, _ := qconn.GetDomain(domain)
	vmIP, _ := dom.GetIP()

	// Read the private key file
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read private key: %v", err)
	}

	// Parse the private key file
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse private key: %v", err)
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
		return nil, fmt.Errorf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Start a session
	session, err := conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %v", err)
	}

	// defer session.Close()

	// Run a simple command
	var b bytes.Buffer
	session.Stdout = &b

	//	cmd := "ls"
	cmd := "echo 'SSH connection successful'"

	if err := session.Run(cmd); err != nil {
		return nil, fmt.Errorf("Failed to run command: %v", err)
	}

	log.Println(b.String())

	return session, nil
}