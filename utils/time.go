package utils

import (
	"fmt"
	"time"
)

// HumanizeDuration converts a time.Duration to a human-readable string.
func HumanizeDuration(duration time.Duration) string {
	if duration.Hours() >= 24 {
		days := duration / (24 * time.Hour)
		return fmt.Sprintf("%d days", days)
	} else if duration.Hours() >= 1 {
		hours := duration / time.Hour
		return fmt.Sprintf("%d hours", hours)
	} else if duration.Minutes() >= 1 {
		minutes := duration / time.Minute
		return fmt.Sprintf("%d minutes", minutes)
	} else {
		seconds := duration / time.Second
		return fmt.Sprintf("%d seconds", seconds)
	}
}
