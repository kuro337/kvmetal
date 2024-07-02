package network

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"kvmgo/utils"
)

/*

/etc/libvirt/hooks/qemu

*/

const SAMPLE_UFW_BEFORE = `#KVM_GO_START
# *nat
# :PREROUTING ACCEPT [0:0]
# -A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:8088 -m comment --comment "Expose Yarn UI on Hadoop Host at 8088 to host 9999"
# COMMIT
#KVM_GO_END`

// ExposeVM gets a VM name (domain) and exposes it on a Port to external Traffic
func ExposeVM(vmname, vmPort, hostPort string) {
	// step 1. figure out VM's IP address and hostname

	vmIP, _ := GetVMIPAddr(vmname)

	log.Printf("VM:%s\nIP Addr: %s", vmname, vmIP)

	// step 2. Get Host IP
	hostIP, _ := GetHostIP()
	log.Printf("Host IP:%s", hostIP.StringWithSubnet())

	// AFTER READING write both to data/network/qemuhooks/ and data/network/ufw/

	/* QEMU HOOKS */
	// step 3. Create the Qemu Hooks File - 1 time thing as Host IP and Libvirt Subnet stays Static
	// Dynamic is simply toggling it (commenting out/in)
	qemuHooksFile := CreateQemuHooksFile()
	utils.LogOffwhite("QEMU HOOKS FILE")

	utils.LogDottedLineDelimitedText(qemuHooksFile)

	/* UFW RULES */
	// now - we will construct the UFW before Rule for it - goes in /etc/ufw/before.rules

	// For Qemu Hooks : we have issues if its active - and we launch a VM with the same name
	// We want to comment it out - if we dont need port forwarding anymore

	utils.LogOffwhite("CURRENT UFW RULES:")
	currentUfwRules, _ := GetCurrentUfwRules()
	utils.LogDottedLineDelimitedText(currentUfwRules)

	ufwBeforeRule := CreateUfwBeforeRule(vmIP.StringWithSubnet(), vmPort, hostPort, "Rule to expose Yarn UI")

	/* UFW: We want to add the Rule here - for each new VM - and delete it once we're done /etc/ufw/before.rules */
	// If we have no more Active VM's : we will delete the Rule and also Comment out Qemu Hooks

	log.Printf("Generated Rule:\n%s", ufwBeforeRule)

	// check if the VM is already exposed

	running := isVMExposed(currentUfwRules, "", vmIP.StringWithSubnet())

	if running {
		log.Printf("VM is already Exposed")
	} else {
		log.Printf("VM is not Exposed")
		log.Printf("Adding UFW Rule:")
		newRuleAdded := AddUfwRule(currentUfwRules, ufwBeforeRule)
		utils.LogOffwhite("Added new UFW Rule:")
		utils.LogDottedLineDelimitedText(newRuleAdded)

	}
}

// isVMExposed doesnt require the Private IP - it will extract it if required. But for performance pass it only the IP will work too
func isVMExposed(ufwFileContent, vmName, ip string) bool {
	var vmIP string
	if ip == "" {
		if vmName == "" {
			log.Printf("One of VMName or IP must be explicity passed to check if its exposed")
			return false
		}
		ip, _ := GetVMIPAddr(vmName)
		vmIP = ip.StringWithSubnet()
	}

	content, active, _ := CheckUfwBeforeHooksActive(ufwFileContent)

	fmt.Printf("Content from ufw before:\n%s", content)

	if strings.Contains(content, vmIP) {
		if active {

			log.Printf("VM is active and included in UFW Before Rules for Port Forwarding.")
			return true
		}
		log.Printf("VM IP is included in Rules but rules are currently inactive")
	}

	log.Printf("VM is not included in Port Forwarding")
	return false
}

// Checks if all lines are commented out
func QemuHooksCheck() {
	filePath := "/etc/libvirt/hooks/qemu"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open qemu hooks file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	commentedLineCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") || len(trimmedLine) == 0 {
			commentedLineCount++
		}
	}

	if lineCount == commentedLineCount {
		fmt.Println("All lines in the qemu hooks file are commented out or empty.")
	} else {
		fmt.Println("Some lines in the qemu hooks file are not commented out.")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %s", err)
	}
}

// sudo vi  /etc/ufw/before.rules
// sudo cat  /etc/ufw/before.rules

/*
#KVM_GO_START
# *nat
# :PREROUTING ACCEPT [0:0]
# -A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:8088 -m comment --comment "Expose Yarn UI on Hadoop Host at 8088 to host 9999"
# COMMIT
#KVM_GO_END
*/
func CheckUfwBeforeHooksActive(ufwFileContent string) (string, bool, error) {
	content, commentedOut, err := utils.ExtractAndCheckComments(ufwFileContent, "#KVM_GO_START", "#KVM_GO_END")
	if err != nil {
		log.Printf("Err parsing and Checking UFW rules %s", err)
		return content, false, err
	}

	if commentedOut {
		log.Printf("UFW Rules are currently inactive")
		return content, false, err

	}

	log.Printf("UFW Rules are ACTIVE")

	log.Printf("Content from Ufw Rules:\n%s", content)

	return content, true, err
}

/*
func AddUfwBeforeRule(vmIP, vmPort, hostPort, description string) error {
	rule := fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s -m comment --comment \"%s\"", hostPort, vmIP, vmPort, description)
	command := fmt.Sprintf("echo -e \"#\\n*nat\\n:PREROUTING ACCEPT [0:0]\\n%s\\nCOMMIT\\n#\" | sudo tee -a /etc/ufw/before.rules", rule)
	_, err := exec.Command("/bin/bash", "-c", command).Output()
	if err != nil {
		return fmt.Errorf("failed to add UFW before rule: %v", err)
	}
	return nil
}

func ReloadUfw() error {
	if err := exec.Command("sudo", "ufw", "reload").Run(); err != nil {
		return fmt.Errorf("failed to reload UFW: %v", err)
	}
	return nil
}
*/

func RemoveUfwBeforeRule(ufwBeforeRule string) error {
	// find line with matching IP - in the rules
	return nil
}

func ToggleQemuHooks(content string, enable bool) error {
	var newContent string
	if enable {
		// Logic to uncomment lines
	} else {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			newContent += "#" + line + "\n"
		}
	}

	return nil
}
