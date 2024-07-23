package utils

import "strings"

// AddExtensionIfRequired adds an extension to the name if it is not present
// @Usage
// AddExtensionIfRequired("kafka",".img") | AddExtensionIfRequired("kafka.img",".img")
// returns "kafka.img" for both
func AddExtensionIfRequired(name, extension string) string {
	if strings.HasSuffix(name, extension) {
		return name
	}
	return name + extension
}
