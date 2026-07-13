package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
)

// AgentGateway is the high-level interface to the pi-coding-agent.
//
// Phase 1 (Variant B): each call to RunStage spawns a fresh pi --mode rpc
// process, sends one prompt, collects the response, and shuts down.
//
// Phase 2 (Variant C): keep a single rpcClient alive across multiple
// RunStage calls for session continuity.  Achieved by adding Start/Stop
// methods that hold a persistent client.
type AgentGateway struct {
	opts models.AgentOptions

	// ── Variant C fields (unused in Phase 1) ──
	// client *rpcClient // set by Start(); nil in Variant B
}

// NewAgentGateway creates an agent gateway with the given options.
func NewAgentGateway(opts models.AgentOptions) *AgentGateway {
	if opts.BinaryPath == "" {
		opts.BinaryPath = "pi"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Minute // generous default for coding tasks
	}
	return &AgentGateway{opts: opts}
}

// ─── Variant B: one-shot per stage ──────────────────────────────────

// StageOutput holds the collected text from a single agent run.
type StageOutput struct {
	Text     string // accumulated text_delta content
	Thinking string // accumulated thinking_delta content (if any)
}

// RunStage spawns pi, sends one prompt, streams events until settled,
// then kills the process.  Returns the collected assistant text.
func (g *AgentGateway) RunStage(ctx context.Context, prompt string) (*StageOutput, error) {
	stageCtx, cancel := context.WithTimeout(ctx, g.opts.Timeout)
	defer cancel()

	client, err := newRPCClient(stageCtx, g.opts)
	if err != nil {
		return nil, fmt.Errorf("agent: start: %w", err)
	}
	defer client.Shutdown()

	events, unsub := client.Subscribe(256)
	defer unsub()

	if _, err := client.SendCommand(stageCtx, rpcCmd{Type: "prompt", Message: prompt}); err != nil {
		return nil, fmt.Errorf("agent: send prompt: %w", err)
	}

	return collectUntilSettled(stageCtx, events, client.Done())
}

// collectUntilSettled reads events until agent_settled, the context is
// cancelled, or the pi process dies.  Accumulates text and thinking.
func collectUntilSettled(
	ctx context.Context,
	events <-chan models.AgentEvent,
	done <-chan struct{},
) (*StageOutput, error) {
	var text, thinking strings.Builder

	for {
		select {
		case evt, ok := <-events:
			if !ok {
				// channel closed without settled → process died
				return &StageOutput{Text: text.String(), Thinking: thinking.String()},
					fmt.Errorf("agent: pi process exited unexpectedly")
			}
			if evt.IsTextDelta() {
				text.WriteString(evt.TextDelta())
			}
			if evt.Type == "message_update" &&
				evt.AssistantMessageEvent != nil &&
				evt.AssistantMessageEvent.Type == "thinking_delta" {
				thinking.WriteString(evt.AssistantMessageEvent.Delta)
			}
			if evt.IsSettled() {
				return &StageOutput{Text: text.String(), Thinking: thinking.String()}, nil
			}
		case <-ctx.Done():
			return &StageOutput{Text: text.String(), Thinking: thinking.String()}, ctx.Err()
		case <-done:
			return &StageOutput{Text: text.String(), Thinking: thinking.String()},
				fmt.Errorf("agent: pi process exited before settled")
		}
	}
}

// ─── Variant C: persistent session (Phase 2) ────────────────────────

// Start begins a persistent pi session.  Subsequent calls to RunStagePersist
// reuse the same process.
func (g *AgentGateway) Start(ctx context.Context) error {
	// Phase 2 — not yet implemented.
	return nil
}

// RunStagePersist sends a prompt on the persistent session.  Requires
// Start() to have been called first.
func (g *AgentGateway) RunStagePersist(ctx context.Context, prompt string) (*StageOutput, error) {
	// Phase 2 — not yet implemented.
	return nil, fmt.Errorf("not implemented: use RunStage (one-shot) for now")
}

// Steer queues a steering message on the persistent session.
func (g *AgentGateway) Steer(ctx context.Context, msg string) error {
	// Phase 2.
	return fmt.Errorf("not implemented")
}

// Abort cancels the in-flight LLM call on the persistent session.
func (g *AgentGateway) Abort() error {
	// Phase 2.
	return fmt.Errorf("not implemented")
}

// Stop gracefully shuts down the persistent session.
func (g *AgentGateway) Stop() {
	// Phase 2.
}
