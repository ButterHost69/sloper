package agent

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ButterHost69/sloper/internal/models"
)

func TestParseAgentMessageErrorMessage(t *testing.T) {
	raw := `{"type":"message_end","message":{"role":"assistant","content":[],"stopReason":"error","errorMessage":"401: {\"type\":\"CreditsError\",\"message\":\"Insufficient balance. Manage your billing here: https://opencode.ai/workspace/wrk_test/billing\"}"}}`
	var evt models.AgentEvent
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if evt.Message == nil {
		t.Fatal("message is nil")
	}
	if evt.Message.StopReason != "error" {
		t.Fatalf("stop reason = %q, want error", evt.Message.StopReason)
	}
	if !strings.Contains(evt.Message.ErrorMessage, "Insufficient balance") {
		t.Fatalf("errorMessage = %q, want it to contain the provider error", evt.Message.ErrorMessage)
	}
}
