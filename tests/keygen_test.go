package tests

import (
	"log"
	"os"
	"testing"

	"kvmgo/configuration"
	"kvmgo/constants"
	"kvmgo/network/ssh"
)

func TestGenerateAndWriteSSHKeyPair(t *testing.T) {
	privateKeyPath := "/home/kuro/Documents/Code/Go/kvmgo/data/keys/id_rsa"
	publicKeyPath := "/home/kuro/Documents/Code/Go/kvmgo/data/keys/id_rsa.pub"

	// Generate keys
	privateKey, publicKey, err := ssh.GenerateSSHKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate SSH key pair: %v", err)
	}

	// Write keys to files
	err = ssh.WriteSSHKeyPair(privateKey, publicKey, privateKeyPath, publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to write SSH key pair to files: %v", err)
	}

	t.Log("SSH key pair generated and written successfully")
}

func TestCreateValidUserdata(t *testing.T) {
	publicKeyPath := "/home/kuro/Documents/Code/Go/kvmgo/data/keys/pub"

	pubk, _ := os.ReadFile(publicKeyPath)
	validUserdata := configuration.SubstituteHostNameAndFqdnUserdataSSHPublicKey(
		constants.DefaultUserdata,
		"testvm",
		string(pubk))

	log.Print(validUserdata)

	t.Error("trigger")
}

func TestSubstitueSSH(t *testing.T) {
	publicKeyPath := "/home/kuro/Documents/Code/Go/kvmgo/data/keys/pub"
	pubk, _ := os.ReadFile(publicKeyPath)
	userdatassh := configuration.SubstituteSSHPubKey(constants.DefaultUserdata, string(pubk))
	log.Print(userdatassh)
	t.Error("trigger")
}
