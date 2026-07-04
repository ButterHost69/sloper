package models

import "time"

// RetryPolicy decides whether to retry a failed command and how long to wait.
// Maybe we make it global, in the models file ; might be reused later
type RetryPolicy interface {
	ShouldRetry(args []string, err error, attempt int) bool // Custom Logic, allows for checking locks and such, can return true if no checking and stuff
	Delay(attempt int) time.Duration
	MaxAttempts() int
}
