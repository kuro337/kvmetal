package utils

import (
	"fmt"
	"strings"

	constants "kvmgo/constants/shell"
)

// GenerateDefaultCloudInit generates Cloud Init Data for a provided hostname
func GenerateDefaultCloudInitZshKernelUpgrade(hostname string) string {
	userData := strings.Replace(constants.Userdata_Literal_zsh_kernelupgrade,
		"#hostname: _HOSTNAME_",
		fmt.Sprintf("hostname: %s", hostname),
		1)
	return userData
}
