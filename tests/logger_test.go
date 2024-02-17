package tests

import (
	"fmt"
	"testing"

	"kvmgo/utils"
)

func TestLogicGate(t *testing.T) {
	utils.MockANSIPrint()
	fmt.Println(utils.TurnBoldBlueDelimited("TEST THIS SECTION BLUE"))

	fmt.Println(utils.AddDelimiter("TEST THIS SECTION BLUE"))

	fmt.Println(utils.TurnBlueDelimited("TEST THIS SECTION BLUE"))
}

// go test -v
// go test
// go test circle_test.go
// go test -v ./mypackage -run TestMyFunction
