package network

import (
	"strings"
)

const SampleUfwCommentedOutFile = `
#KVM_GO_START
# *nat
# :PREROUTING ACCEPT [0:0]
# -A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:8088 -m comment --comment "Expose Yarn UI on Hadoop Host at 8088 to host 9999"
# COMMIT
#KVM_GO_END
#
# rules.before
#
otherrules
mightbe: valid config here 
etc. 
etc. 
# comments 
# comments 


`

const SampleUfwActiveFile = `
#KVM_GO_START
*nat
:PREROUTING ACCEPT [0:0]
-A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:9999 -m comment --comment "Testing port 9999 of vm from ubuntu host 9999"
COMMIT
#KVM_GO_END


# all other non-local packets are dropped
-A ufw-not-local -m limit --limit 3/min --limit-burst 10 -j ufw-logging-deny
-A ufw-not-local -j DROP

# allow MULTICAST UPnP for service discovery (be sure the MULTICAST line above
# is uncommented)
-A ufw-before-input -p udp -d 239.255.255.250 --dport 1900 -j ACCEPT
	
# don't delete the 'COMMIT' line or these rules won't be processed
COMMIT
`

const SampleUfwActive = `
#KVM_GO_START
*nat
:PREROUTING ACCEPT [0:0]
-A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:9999 -m comment --comment "Testing port 9999 of vm from ubuntu host 9999"
COMMIT
#KVM_GO_END
`

const SampleUfwInActive = `
#KVM_GO_START
# *nat
# :PREROUTING ACCEPT [0:0]
# -A PREROUTING -p tcp --dport 9999 -j DNAT --to-destination 192.168.122.109:9999 -m comment --comment "Testing port 9999 of vm from ubuntu host 9999"
# COMMIT
#KVM_GO_END
`

// DisableUfwForwarding comments out the rules between #KVM_GO_START and #KVM_GO_END
func DisableUfwForwarding(content string) string {
	return ToggleUfwRules(content, true)
}

// ActivateUfwForwarding uncomments the rules between #KVM_GO_START and #KVM_GO_END
func ActivateUfwForwarding(content string) string {
	return ToggleUfwRules(content, false)
}

// Adds a new UFW rule between #KVM_GO_START and #KVM_GO_END
func AddUfwRule(content, newRule string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inSection bool

	for _, line := range lines {
		if strings.Contains(line, "#KVM_GO_END") {
			// Add the new rule right before the end marker
			// Check if the new rule should be commented based on the section state
			if IsSectionCommented(result) {
				result = append(result, "# "+newRule)
			} else {
				result = append(result, newRule)
			}
			inSection = false
		}
		if inSection {
			// Skip adding the rule if it's already present
			if strings.Contains(line, newRule) {
				continue
			}
		}
		result = append(result, line)
		if strings.Contains(line, "#KVM_GO_START") {
			inSection = true
		}
	}
	return strings.Join(result, "\n")
}

// RemoveUfwRule removes a specified rule found between the #KVM_GO_START and #KVM_GO_END markers.
func RemoveUfwRule(content, ruleToRemove string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inSection bool

	for _, line := range lines {
		if strings.Contains(line, "#KVM_GO_START") {
			inSection = true
		}
		if strings.Contains(line, "#KVM_GO_END") {
			inSection = false
		}
		if inSection && (strings.Contains(line, ruleToRemove) || strings.Contains(line, "# "+ruleToRemove)) {
			// Skip the rule to remove
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// Toggle the commenting state of the section between #KVM_GO_START and #KVM_GO_END based on the current state.
func ToggleUfwRules(content string, enable bool) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inSection bool
	sectionCommented := IsSectionCommented(lines)

	for _, line := range lines {
		if strings.Contains(line, "#KVM_GO_START") || strings.Contains(line, "#KVM_GO_END") {
			result = append(result, line)
			inSection = !inSection
			continue
		}
		if inSection {
			if enable && sectionCommented {
				line = strings.TrimPrefix(line, "# ")
			} else if !enable && !sectionCommented {
				line = "# " + line
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func IsSectionCommented(lines []string) bool {
	var inSection bool
	for _, line := range lines {
		if strings.Contains(line, "#KVM_GO_START") {
			inSection = true
			continue
		}
		if strings.Contains(line, "#KVM_GO_END") {
			break
		}
		if inSection && !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			return false
		}
	}
	return true
}
