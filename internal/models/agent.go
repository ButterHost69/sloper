package models

import "time"

// AgentOptions configures the pi-coding-agent RPC client.
type AgentOptions struct {
	BinaryPath string        // path to pi, defaults to "pi"
	CWD        string        // repo working directory
	Model      string        // e.g. "claude-sonnet-4:high"
	Thinking   string        // off | minimal | low | medium | high | xhigh | max
	APIKey     string        // API key for the provider (passed as --api-key)
	Provider   string        // provider name (passed as --provider, e.g. "anthropic", "openai")
	SessionID  string        // deterministic session ID (passed as --session-id)
	SessionDir string        // directory for session storage (passed as --session-dir)
	Timeout    time.Duration // per-prompt deadline; 0 = no deadline
}

// AgentEvent mirrors a JSONL line from pi --mode rpc stdout.
type AgentEvent struct {
	Type                  string         `json:"type"`
	AssistantMessageEvent *AgentMsgDelta `json:"assistantMessageEvent,omitempty"`
	Message               *AgentMessage  `json:"message,omitempty"`
	ToolName              string         `json:"toolName,omitempty"`
	ToolCallID            string         `json:"toolCallId,omitempty"`
	IsError               bool           `json:"isError,omitempty"`
	Error                 string         `json:"error,omitempty"`
	ID                    string         `json:"id,omitempty"`
	Command               string         `json:"command,omitempty"`
	Success               *bool          `json:"success,omitempty"`
	WillRetry             bool           `json:"willRetry,omitempty"`
}

// AgentMessage is the message object inside message_start / message_end events.
type AgentMessage struct {
	Role       string             `json:"role"` // user | assistant | toolResult | custom
	Content    []AgentContentPart `json:"content"`
	StopReason string             `json:"stopReason,omitempty"`
}

// AgentContentPart is one part of a message's content array.
type AgentContentPart struct {
	Type string `json:"type"` // text | tool_call | etc.
	Text string `json:"text,omitempty"`
}

// AgentMsgDelta is the streaming delta inside message_update events.
type AgentMsgDelta struct {
	Type    string `json:"type"` // text_delta | thinking_delta | toolcall_start | …
	Delta   string `json:"delta,omitempty"`
	Content string `json:"content,omitempty"`
}

// IsSettled returns true when this event signals the agent has fully settled.
func (e AgentEvent) IsSettled() bool { return e.Type == "agent_settled" }

// IsTextDelta returns true when this event carries a text chunk.
func (e AgentEvent) IsTextDelta() bool {
	return e.Type == "message_update" &&
		e.AssistantMessageEvent != nil &&
		e.AssistantMessageEvent.Type == "text_delta"
}

// TextDelta returns the delta string when IsTextDelta is true, "" otherwise.
func (e AgentEvent) TextDelta() string {
	if e.AssistantMessageEvent != nil {
		return e.AssistantMessageEvent.Delta
	}
	return ""
}

// IsTerminal returns true for events that signal no more agent work
// (agent_settled, agent_end without willRetry, or a fatal error).
func (e AgentEvent) IsTerminal() bool {
	return e.Type == "agent_settled"
}

// ─── Pipeline stages ────────────────────────────────────────────────

// PipelineStage identifies which phase of issue processing we are in.
type PipelineStage int

const (
	StageSpec PipelineStage = iota
	StageWork
	StageReview
	StageFix
	StageMerge
)

func (s PipelineStage) String() string {
	switch s {
	case StageSpec:
		return "spec"
	case StageWork:
		return "work"
	case StageReview:
		return "review"
	case StageFix:
		return "fix"
	case StageMerge:
		return "merge"
	default:
		return "unknown"
	}
}

// ─── Structured stage outputs ───────────────────────────────────────

// SpecResult is the agent's output from the SPEC stage.
type SpecResult struct {
	Summary            string   `json:"summary"`
	FilesToChange      []string `json:"files_to_change"`
	ImplementationPlan string   `json:"implementation_plan"`
	RawOutput          string   `json:"-"`
}

// WorkResult is the agent's output from the WORK stage.
type WorkResult struct {
	BranchName string `json:"branch_name"`
	PRNumber   int64  `json:"pr_number"`
	Diff       string `json:"diff"`
	CommitMsg  string `json:"commit_msg"`
	RawOutput  string `json:"-"`
}

// ReviewResult is the agent's output from the REVIEW stage.
type ReviewResult struct {
	Approved    bool     `json:"approved"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	RawOutput   string   `json:"-"`
}
