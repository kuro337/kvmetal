package tests

// #hostname: _HOSTNAME_

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"kvmgo/cli"
	"kvmgo/configuration/presets"
	"kvmgo/constants"
	"kvmgo/utils"
)

func TestMinimal(t *testing.T) {
	if err := cli.TestLaunchConf("control"); err != nil {
		t.Logf("ERROR :%s\n", err)
	}
}

func TestCloudInitValidSchema(t *testing.T) {
	hadoop_userdata := presets.CreateKafkaUserData("ubuntu",
		"password",
		"kafka",
		utils.ReadFileFatal(constants.SshPub))

	//	os.WriteFile("testfile.yaml", []byte(hadoop_userdata), 0o644)
	tmpfile, err := os.CreateTemp("", "testfile.yaml")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(hadoop_userdata)

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write([]byte(hadoop_userdata)); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Build the command to run with sudo
	cmd := exec.Command("sudo", "cloud-init", "schema", "--config-file", tmpfile.Name())

	// Capture the output or error of the command
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n%s", err, stderr.String())
	}

	log.Print(hadoop_userdata)
	// Check command output for success message
	if !strings.Contains(out.String(), "Valid cloud-config") {
		t.Errorf("cloud-init schema validation failed: %s", out.String())
	}

	// Your additional validation logic here
	if !strings.Contains(hadoop_userdata, "hadoop") {
		t.Errorf("Parsing and Substitution Failed")
	}
	t.Error("trigger")
	// Your additional validation logic here
	if !strings.Contains(hadoop_userdata, "hadoop.kuro.com") {
		t.Errorf("Parsing and Substitution Failed")
	}
}

// go test -v
// go test
// go test circle_test.go
// go test -v ./mypackage -run TestMyFunction
