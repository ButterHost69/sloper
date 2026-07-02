package models

import (
	"time"
)

const (
	DefaultShellGracefulStop   = 5 * time.Second
	DefaultShellMaxOutputBytes = 256 * 1024
)

type ShellResult struct {
	ExitCode   int
	Duration   time.Duration
	DurationMS int64

	Stdout string
	Stderr string
}

type ShellOptions struct {
	Command          string
	Args             []string
	CWD              string
	Env              map[string]string
	Timeout          time.Duration
	GracefulShutdown time.Duration

	MaxCapturedBytes int // Will be needed for truncating agents outputs
	// Stdin            string 		** Not sure where needed ** TODO: REMOVE THIS IF NOT IN USE AFTER SOME TIME
}

type ShellCommandExecutionError struct {
	Message string
	Result  ShellResult
}

type BoundedBuffer struct {
	data  []byte
	limit int
}

func NewBoundedBuffer(limit int) *BoundedBuffer {
	if limit == 0 {
		limit = DefaultShellMaxOutputBytes
	}

	return &BoundedBuffer{
		limit: limit,
	}
}

func (b *BoundedBuffer) Write(data []byte) (int,error) {
	// NOTE: need return error - because this is an io.writer alternative

	if len(b.data) >= b.limit {
		return 0, nil
	}
	
	remaining := b.limit - len(b.data)
	if len(data) > remaining{
		data = data[:remaining]
	}

	b.data = append(b.data, data...)
	return len(data), nil
}

func (b *BoundedBuffer) String() string {
	return string(b.data)
}

func (e *ShellCommandExecutionError) Error() string { return e.Message }
