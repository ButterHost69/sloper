package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

func DecodeJSONObject(value string) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, invalidJSONError(value, err)
	}
	if out == nil {
		return map[string]any{}, nil
	}
	return out, nil
}

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

func DecodeJSONArrayOrPages(value string) ([]map[string]any, error) {
	rows, err := DecodeJSONArray(value)
	if err == nil {
		return rows, nil
	}
	var pages [][]map[string]any
	if pageErr := json.Unmarshal([]byte(value), &pages); pageErr != nil {
		return nil, err
	}
	for _, page := range pages {
		rows = append(rows, page...)
	}
	if rows == nil {
		return []map[string]any{}, nil
	}
	return rows, nil
}

func ToObjectSlice(value any) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]any); ok {
			out = append(out, row)
		}
	}
	return out
}