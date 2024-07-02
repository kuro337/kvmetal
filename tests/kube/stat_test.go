package tests

import (
	"regexp"
	"strings"
	"testing"

	"kvmgo/kube"
	"kvmgo/kube/join"
)

// SplitSoutLineDelimited splits a string by newlines and then splits each line by a delimiter
// If each line has at least the colsExpected - it includes the line values in the response
// Any invalid lines are included in the second return value invalid[]
// Pass "" and -1 as default params to simply split by newlines and by " " as the delimiter
func SplitSoutLineDelimited(stdout, lineDelim string, colsExpected int) ([][]string, []string) {
	if lineDelim == "" {
		lineDelim = " "
	}

	re := regexp.MustCompile(`\s+`) //  Regex to replace spaces

	var data [][]string

	var invalid []string

	lines := strings.Split(stdout, "\n") //  split each line

	for _, line := range lines {
		result := re.ReplaceAllString(line, " ")
		cols := strings.Split(result, lineDelim)
		if colsExpected == -1 || len(cols) >= colsExpected {
			data = append(data, cols[0:colsExpected])
		} else {
			invalid = append(invalid, line)
		}

	}

	return data, invalid
}

func TestNodesCtrl(t *testing.T) {
	controlDomain := "control"

	workers := []string{"worker"}
	control, err := kube.NewControl(controlDomain)
	if err != nil {
		t.Errorf("Error:%s", err)
	}

	if _, err := join.VerifyNodes(control, workers); err != nil {
		t.Logf("error from verify nodes: %s\n", err)
	}
}
