package tests

import (
	"fmt"
	"testing"

	"kvmgo/configuration/presets"
)

func TestConfigParse(t *testing.T) {
	ans := presets.CreateKafkaUserData("ubuntu", "password", "customdomain", "1234xxx444")
	fmt.Println(ans)

	t.Error(ans)
}
