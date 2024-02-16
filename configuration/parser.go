package configuration

import (
	"fmt"
	"strings"
)

// Relies on the provided String having a line #hostname: _HOSTNAME_ - replaces this with provided hostname
func SubstituteHostnameUserData(yamlTemplate, hostname string) string {
	userData := strings.Replace(yamlTemplate,
		"#hostname: _HOSTNAME_",
		fmt.Sprintf("hostname: %s", hostname),
		1)
	return userData
}
