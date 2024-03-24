package utils

import (
	"log"
	"time"
)

// RetryUntil retries a function until a condition is satisfied or max attempts are reached
func RetryUntilString(fn func() (string, error), condition func(error) bool, maxAttempts int, backoff time.Duration) (string, error) {
	var ip string
	var err error
	for attempts := 0; attempts < maxAttempts; attempts++ {
		log.Printf("Attempt %d", attempts)
		ip, err = fn()
		if err == nil || !condition(err) {
			return ip, err
		}
		time.Sleep(backoff)
	}
	return "", err // Return the last error encountered
}
