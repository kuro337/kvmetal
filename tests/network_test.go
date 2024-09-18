package tests

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"kvmgo/lib"
	"kvmgo/network"
	"kvmgo/utils"
)

/* Uses Both Methods to get the IP Address of a running Domain */
func TestDomainIPExtractionAligned(t *testing.T) {
	domain := "test"

	vmIpAddr, err := network.GetVMIPAddr(domain)
	if err != nil {
		t.Errorf("Failed to get Nat Subnet")
	}

	vmIPSyscall := vmIpAddr.IP.String()
	conn, _ := lib.ConnectLibvirt()
	results, _ := conn.GetIPFromDHCPLeases(domain)

	if len(results) == 0 {
		t.Error("Libvirt Failed to get any IP")
	}
	found := false

	for _, x := range results {
		if x == vmIPSyscall {
			found = true
		}
	}

	if found == false {
		t.Errorf("Misalignment in IP Extraction.\nSystem IP :%s\nLibVirt IP :%s\n", vmIPSyscall, results[0])
	}
}

// Gets the IP using a Syscall
func TestGetVMIPAddress(t *testing.T) {
	domain := "consul"

	vmIpAddr, err := network.GetVMIPAddr(domain)
	if err != nil {
		t.Errorf("Failed to get Nat Subnet")
	}
	log.Printf("Domain %s IP is %s\n", domain, vmIpAddr)
}

func TestAddUfwRule(t *testing.T) {
	content := network.SampleUfwCommentedOutFile
	newRule := "-A PREROUTING -p tcp --dport 8888 -j DNAT --to-destination 192.168.122.110:1111 -m comment --comment \"New Rule Testing VM port 8888 to 1111\""

	updatedContent := network.AddUfwRule(content, newRule)

	if !strings.Contains(updatedContent, newRule) {
		fmt.Println(updatedContent)
		t.Errorf("New rule was not added as expected")
	}
}

func TestUfwAddRemoveRule(t *testing.T) {
	content := network.SampleUfwCommentedOutFile
	newRule := "-A PREROUTING -p tcp --dport 8888 -j DNAT --to-destination 192.168.122.110:1111 -m comment --comment \"New Rule Testing VM port 8888 to 1111\""

	updatedContent := network.AddUfwRule(content, newRule)

	if !strings.Contains(updatedContent, newRule) {
		fmt.Println(updatedContent)
		t.Errorf("New Rule was not added as expected")
	}

	updatedContent = network.RemoveUfwRule(content, newRule)

	if strings.Contains(updatedContent, newRule) {
		t.Errorf("Rule was not removed as expected")
	}
}

func TestRemoveUfwRule(t *testing.T) {
	content := network.SampleUfwCommentedOutFile
	oldRule := "-A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:8088 -m comment --comment \"Expose Yarn UI on Hadoop Host at 8088 to host 9999\""

	updatedContent := network.RemoveUfwRule(content, oldRule)

	if strings.Contains(updatedContent, oldRule) {
		t.Errorf("Rule was not removed as expected")
		fmt.Println(updatedContent)
	}
}

// func TestCreateQemu(t *testing.T) {
// 	qemuConf := network.CreateQemuHooksFile()

// 	utils.LogDottedLineDelimitedText(qemuConf)

// 	t.Errorf("trigger")
// }

func TestQemuInteraction(t *testing.T) {
	network.ExposeVM("hadoop", "8088", "5555")

	network.QemuHooksCheck()

	qemuConf := network.CreateQemuHooksFile()

	commented := utils.CommentOutFile(qemuConf, "#")

	allCommented := utils.IsFileCommented(commented, "#", -1, -1)

	if allCommented == false {
		t.Errorf("Commenting Out Logic Failure - after CommentingOut it should be True")
	}

	uncommented := utils.UnCommentOutFile(commented, "#")
	allCommented = utils.IsFileCommented(uncommented, "#", -1, -1)

	fmt.Printf("EXPECTED ALL UNCOMMENTED:\n%s\nResult:\n%v", uncommented, allCommented)

	if allCommented == true {
		t.Errorf("UnCommenting Logic Failure")
	}

	// t.Errorf("trigger")
}

// go test -v
// go test
// go test circle_test.go
// go test -v ./mypackage -run TestMyFunction
