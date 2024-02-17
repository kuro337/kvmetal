package utils

import (
	"fmt"
	"log"
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

// ExtractAndCheckComments reads the specified file and checks if the content between
// the start and end markers is fully commented out. It returns the content, a boolean
// indicating if all lines are commented, and any error encountered.
func ExtractAndCheckComments(content, startMarker, endMarker string) (string, bool, error) {
	extracting := false
	extractedContent := ""
	allCommented := true

	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, startMarker) {
			extracting = true
			continue
		}

		if strings.Contains(line, endMarker) {
			break
		}

		if extracting {
			extractedContent += line + "\n"
			if !strings.HasPrefix(strings.TrimSpace(line), "#") && strings.TrimSpace(line) != "" {
				allCommented = false
			}
		}
	}

	if extractedContent == "" {
		return "", false, fmt.Errorf("no content extracted, check your markers")
	}

	return extractedContent, allCommented, nil
}

// CommentOutFile adds a comment character to the beginning of each line in the given content.
func CommentOutFile(content, commentChar string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Only add comment character to non-empty lines
		if strings.TrimSpace(line) != "" {
			lines[i] = commentChar + " " + line
		}
	}
	return strings.Join(lines, "\n")
}

// UnCommentOutFile removes the comment character from the beginning of each line in the given content,
// if the line starts with that comment character.
func UnCommentOutFile(content, commentChar string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// Only remove comment character from lines that start with it
		if strings.HasPrefix(trimmedLine, commentChar+" ") {
			lines[i] = strings.TrimPrefix(trimmedLine, commentChar+" ")
		} else if strings.HasPrefix(trimmedLine, commentChar) { // For cases where there's no space after comment char
			lines[i] = strings.TrimPrefix(trimmedLine, commentChar)
		}
	}
	return strings.Join(lines, "\n")
}

// IsLineCommented checks if lines within a specified range are commented out.
// If startLine and endLine are both -1, it checks all lines.
// Returns true if all lines in the specified range are commented out, false otherwise.
func IsFileCommented(content, commentChar string, startLine, endLine int) bool {
	lines := strings.Split(content, "\n")

	if startLine == -1 {
		startLine = 0
		if strings.HasPrefix(lines[0], "#!/bin/bash") {
			log.Printf("Shell Script Detected : NOT TREATING #!/bin/bash as a Comment!")
			startLine = 1
		}

	}

	if endLine == -1 || endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	commentedLines := 0
	for i, line := range lines {
		if i < startLine || i > endLine {
			continue
		}

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, commentChar) {
			commentedLines++
		}
	}

	totalLinesToCheck := endLine - startLine + 1
	return commentedLines == totalLinesToCheck // Check if all lines in the range are commented or empty
}
