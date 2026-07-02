package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)


func DecodeJSONArray(value string) ([]map[string]any, error) {
	var out []map[string]any
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, invalidJSONError(value, err)
	}
	if out == nil {
		return []map[string]any{}, nil
	}
	return out, nil
}

func invalidJSONError(stdout string, err error) error {
	message := "Invalid gh JSON payload"
	errText := ""
	if err != nil {
		errText = err.Error()
		message += ": " + errText
	}
	message += fmt.Sprintf("; stdoutBytes=%d", len(stdout))
	if sample := summarizeInvalidJSONPayload(stdout); sample != "" {
		message += "; stdoutSample=" + strconv.Quote(sample)
	}

	// TODO: Why use the ShellCommandExecutionError ? wouldnt a normal error.New not work ? maybe something related to maybe needing this result or something.
	return &models.ShellCommandExecutionError{Message: message, Result: models.ShellResult{ExitCode: 0, Stdout: stdout, Stderr: errText}}
}

func summarizeInvalidJSONPayload(stdout string) string {
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return ""
	}
	stdout = strings.Join(strings.Fields(stdout), " ")
	const maxSampleBytes = 240
	if len(stdout) <= maxSampleBytes {
		return stdout
	}
	return strings.TrimSpace(stdout[:maxSampleBytes]) + "…"
}