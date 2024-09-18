package configuration

import (
	"fmt"
	"strings"
)

/* Adds the Hostname, fqdn, & SSH key to the Userdata */
func SubstituteHostNameAndFqdnUserdataSSHPublicKey(yamlTemplate, hostname, sshPubKey string) string {
	hostNameAdded := SubstituteHostnameUserData(yamlTemplate, hostname)
	fqdnAdded := SubstituteFqdnUserData(hostNameAdded, hostname)
	sshKeyAdded := SubstituteSSHPubKey(fqdnAdded, sshPubKey)
	return sshKeyAdded
}

// Relies on the provided String having a line #hostname: _HOSTNAME_ - replaces this with provided hostname
func SubstituteHostNameAndFqdnUserdata(yamlTemplate, hostname string) string {
	hostNameAdded := SubstituteHostnameUserData(yamlTemplate, hostname)
	return SubstituteFqdnUserData(hostNameAdded, hostname)
}

// Relies on the provided String having a line #hostname: _HOSTNAME_ - replaces this with provided hostname
func SubstituteHostnameUserData(yamlTemplate, hostname string) string {
	userData := strings.Replace(yamlTemplate,
		"#hostname: _HOSTNAME_",
		fmt.Sprintf("hostname: %s", hostname),
		1)
	return userData
}

// SubstituteFqdnUserData adds the fqdn so DHCP can find it
func SubstituteFqdnUserData(yamlTemplate, hostname string) string {
	userData := strings.Replace(yamlTemplate,
		"#fqdn: _FQDN_",
		fmt.Sprintf("fqdn: %s", hostname+".kuro.com"),
		1)
	return userData
}

func SubstituteSSHPubKey(yamlTemplate, sshPubKey string) string {
	repl := "#  - ssh-rsa $SSH_PUB"

	ans := strings.Replace(yamlTemplate, repl, fmt.Sprintf("  - %s", sshPubKey), 1)
	ans = strings.Replace(ans, "#ssh_authorized_keys:", "ssh_authorized_keys:", 1)

	return ans
}
