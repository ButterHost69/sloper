// Runs commands in shell
// Like gh, pi etc ...
package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ButterHost69/sloper/internal/logger"
	"github.com/ButterHost69/sloper/internal/models"
	"go.uber.org/zap"
)

func Run(ctx context.Context, options models.ShellOptions) (models.ShellResult, error) {
	if strings.TrimSpace(options.Command) == "" {
		return models.ShellResult{}, fmt.Errorf("command is required")
	}

	startedTime := time.Now()
	maxCapturedBytes := options.MaxCapturedBytes
	if maxCapturedBytes <= 0 {
		maxCapturedBytes = models.DefaultShellMaxOutputBytes
	}

	gracefulShutdown := options.GracefulShutdown
	if gracefulShutdown <= 0 {
		gracefulShutdown = models.DefaultShellGracefulStop
	}

	cmd := exec.Command(options.Command, options.Args...)
	cmd.Dir = options.CWD
	if len(options.Env) > 0 {
		cmd.Env = mapEnv(options.Env)
	}

	if options.Stdin != "" {
		cmd.Stdin = strings.NewReader(options.Stdin)
	}

	stdOutBuffer := models.NewBoundedBuffer(maxCapturedBytes)
	stdErrBuffer := models.NewBoundedBuffer(maxCapturedBytes)

	cmd.Stdout = stdOutBuffer
	cmd.Stderr = stdErrBuffer

	if err := cmd.Start(); err != nil {
		return models.ShellResult{}, fmt.Errorf("error occured in executing cmd: %w", err) // TODO: Provide the Entire Object when returning error so we can neatly print/log it
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	var (
		waitErr          error
		timedOut         bool
		canceledErr      error
		terminationStart <-chan time.Time
		killAt           <-chan time.Time
		terminatOnce     sync.Once
	)

	terminateFunc := func() {
		terminatOnce.Do(func() {
			if cmd.Process == nil {
				return
			}

			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil && !isProcessDone(err) {
				_ = cmd.Process.Kill()
				return
			}

			if gracefulShutdown <= 0 {
				_ = cmd.Process.Kill()
				return
			}

			killAt = time.After(gracefulShutdown)
		})
	}

	if options.Timeout > 0 {
		terminationStart = time.After(options.Timeout)
	}

	waiting := true
	for waiting {
		select {
		case waitErr = <-waitCh:
			waiting = false
		case <-terminationStart:
			timedOut = true
			terminationStart = nil
			terminateFunc()
		case <-killAt:
			killAt = nil
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		case <-ctx.Done():
			if canceledErr == nil {
				canceledErr = ctx.Err()
				terminateFunc()
			}
		}
	}

	duration := time.Since(startedTime)
	result := models.ShellResult{
		ExitCode:   exitCode(cmd),
		Stdout:     stdOutBuffer.String(),
		Stderr:     stdErrBuffer.String(),
		Duration:   duration,
		DurationMS: duration.Milliseconds(),
	}

	logShellCommand(options, result, duration, timedOut, canceledErr)

	if timedOut {
		// TODO: Find Why are there duplicates being sent ? also why is this sent as a pointer ??
		return result, &models.ShellCommandExecutionError{Message: "Command Timed Out", Result: result}
	}

	if canceledErr != nil {
		return result, canceledErr
	}

	if result.ExitCode != 0 {
		return result, &models.ShellCommandExecutionError{Message: commandFailureMessage(result), Result: result}
	}

	if waitErr != nil {
		return result, waitErr
	}

	return result, nil
}

// Converts {key:value} -> "key=value"
func mapEnv(envMap map[string]string) []string {
	envs := make([]string, len(envMap))
	idx := 0
	for key, value := range envMap {
		envs[idx] = fmt.Sprintf("%s=%s", key, value)
	}

	return envs
}

func isProcessDone(err error) bool {
	return err == nil || err == os.ErrProcessDone
}

func exitCode(cmd *exec.Cmd) int {
	if cmd == nil || cmd.ProcessState == nil {
		return -1
	}
	return cmd.ProcessState.ExitCode()
}

func commandFailureMessage(result models.ShellResult) string {
	message := fmt.Sprintf("Command exited with code %d", result.ExitCode)
	stderr := strings.TrimSpace(result.Stderr)
	stdout := strings.TrimSpace(result.Stdout)
	if stderr != "" {
		message += ": " + stderr
	}
	if stdout != "" {
		if stderr != "" {
			message += "\nstdout: " + stdout
		} else {
			message += ": " + stdout
		}
	}
	return message
}

func logShellCommand(options models.ShellOptions, result models.ShellResult, duration time.Duration, timedOut bool, canceledErr error) {
	log := logger.Default()
	if log == nil {
		return
	}

	args := strings.Join(options.Args, " ")
	fields := []zap.Field{
		zap.String("command", options.Command+" "+args),
		zap.Int("exit_code", result.ExitCode),
		zap.Duration("duration", duration),
	}
	if options.CWD != "" {
		fields = append(fields, zap.String("cwd", options.CWD))
	}
	if result.Stdout != "" {
		stdout := strings.TrimSpace(result.Stdout)
		if len(stdout) > 500 {
			stdout = stdout[:500] + "...(truncated)"
		}
		fields = append(fields, zap.String("stdout", stdout))
	}
	if result.Stderr != "" {
		stderr := strings.TrimSpace(result.Stderr)
		if len(stderr) > 500 {
			stderr = stderr[:500] + "...(truncated)"
		}
		fields = append(fields, zap.String("stderr", stderr))
	}
	if timedOut {
		fields = append(fields, zap.Bool("timed_out", true))
	}
	if canceledErr != nil {
		fields = append(fields, zap.String("canceled", canceledErr.Error()))
	}

	if result.ExitCode == 0 && !timedOut && canceledErr == nil {
		log.Debug("shell: command completed", fields...)
	} else {
		log.Warn("shell: command failed", fields...)
	}
}
