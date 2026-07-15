package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ButterHost69/sloper/internal/logger"
	"github.com/ButterHost69/sloper/internal/models"
	"go.uber.org/zap"
)

type AgentGateway struct {
	opts models.AgentOptions
}

func NewAgentGateway(opts models.AgentOptions) *AgentGateway {
	if opts.BinaryPath == "" {
		opts.BinaryPath = "pi"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Minute
	}
	return &AgentGateway{opts: opts}
}

type StageOutput struct {
	Text     string
	Thinking string
}

func (g *AgentGateway) RunStage(ctx context.Context, prompt string) (*StageOutput, error) {
	return g.RunStageWithCWD(ctx, prompt, "", "")
}

func (g *AgentGateway) RunStageWithCWD(ctx context.Context, prompt, cwdOverride, sessionID string) (*StageOutput, error) {
	log := logger.Default()

	stageCtx, cancel := context.WithTimeout(ctx, g.opts.Timeout)
	defer cancel()

	opts := g.opts
	if cwdOverride != "" {
		opts.CWD = cwdOverride
	}
	if sessionID != "" {
		opts.SessionID = sessionID
	}

	log.Info("agent: starting pi process",
		zap.String("binary", opts.BinaryPath),
		zap.String("cwd", opts.CWD),
		zap.String("model", opts.Model),
		zap.String("provider", opts.Provider),
		zap.String("session_id", opts.SessionID),
		zap.Bool("has_api_key", opts.APIKey != ""))

	client, err := newRPCClient(stageCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("agent: start: %w", err)
	}
	defer client.Shutdown()

	events, unsub := client.Subscribe(256)
	defer unsub()

	if _, err := client.SendCommand(stageCtx, rpcCmd{Type: "prompt", Message: prompt}); err != nil {
		return nil, fmt.Errorf("agent: send prompt: %w", err)
	}

	log.Info("agent: prompt accepted, waiting for response...")
	return collectUntilSettled(stageCtx, events, client.Done())
}

func collectUntilSettled(
	ctx context.Context,
	events <-chan models.AgentEvent,
	done <-chan struct{},
) (*StageOutput, error) {
	log := logger.Default()
	var text, thinking strings.Builder
	var messageEndText string
	var sawSettled bool
	var eventCount int

	for {
		select {
		case evt, ok := <-events:
			if !ok {
				log.Warn("agent: event channel closed before settled",
					zap.Int("events_seen", eventCount),
					zap.Int("text_len", text.Len()))
				return &StageOutput{Text: text.String(), Thinking: thinking.String()},
					fmt.Errorf("agent: pi process exited unexpectedly")
			}
			eventCount++

			log.Debug("agent: event",
				zap.Int("n", eventCount),
				zap.String("type", evt.Type),
				zap.Bool("is_text_delta", evt.IsTextDelta()),
				zap.Bool("is_settled", evt.IsSettled()),
				zap.Bool("is_error", evt.IsError || evt.Error != ""))

			switch evt.Type {
			case "agent_start", "turn_start", "turn_end", "agent_end", "agent_settled":
				log.Info("agent: " + evt.Type)
				if evt.Type == "agent_end" && evt.WillRetry {
					log.Info("agent: will_retry")
				}
			case "message_start":
				if evt.Message != nil {
					log.Info("agent: message_start", zap.String("role", evt.Message.Role))
				}
			case "message_end":
				if evt.Message != nil {
					log.Info("agent: message_end",
						zap.String("role", evt.Message.Role),
						zap.String("stop_reason", evt.Message.StopReason))
					if evt.Message.Role == "assistant" {
						msgText := extractMessageText(evt.Message)
						if msgText != "" {
							log.Warn("agent: assistant message_end text",
								zap.String("stop_reason", evt.Message.StopReason),
								zap.String("text", truncate(msgText, 2000)))
							if text.Len() == 0 {
								messageEndText = msgText
							}
						}
						if evt.Message.StopReason == "error" {
							evtJSON, _ := json.Marshal(evt)
							log.Error("agent: assistant error — raw event",
								zap.String("raw", string(evtJSON)))
							if evt.Error != "" && messageEndText == "" {
								messageEndText = evt.Error
							}
						}
					}
				}
			}

			// Capture streaming text deltas
			if evt.IsTextDelta() {
				text.WriteString(evt.TextDelta())
			}
			// Capture thinking deltas
			if evt.Type == "message_update" &&
				evt.AssistantMessageEvent != nil &&
				evt.AssistantMessageEvent.Type == "thinking_delta" {
				thinking.WriteString(evt.AssistantMessageEvent.Delta)
			}
			// Log errors from events
			if evt.IsError || evt.Error != "" {
				log.Error("agent: error event",
					zap.String("type", evt.Type),
					zap.String("error", evt.Error))
			}
			// Settled = agent is done
			if evt.IsSettled() {
				sawSettled = true
				finalText := text.String()
				if finalText == "" && messageEndText != "" {
					finalText = messageEndText
				}
				log.Info("agent: settled",
					zap.Int("events_seen", eventCount),
					zap.Int("text_len", len(finalText)),
					zap.Int("thinking_len", thinking.Len()))
				return &StageOutput{Text: finalText, Thinking: thinking.String()}, nil
			}

		case <-ctx.Done():
			log.Warn("agent: context cancelled",
				zap.Int("events_seen", eventCount),
				zap.Int("text_len", text.Len()),
				zap.Bool("saw_settled", sawSettled))
			return &StageOutput{Text: text.String(), Thinking: thinking.String()}, ctx.Err()

		case <-done:
			log.Warn("agent: process done channel closed",
				zap.Int("events_seen", eventCount),
				zap.Int("text_len", text.Len()),
				zap.Bool("saw_settled", sawSettled))
			return &StageOutput{Text: text.String(), Thinking: thinking.String()},
				fmt.Errorf("agent: pi process exited before settled")
		}
	}
}

func extractMessageText(msg *models.AgentMessage) string {
	if msg == nil {
		return ""
	}
	var parts []string
	for _, part := range msg.Content {
		if part.Type == "text" && part.Text != "" {
			parts = append(parts, part.Text)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}
	// Content is empty — for errors, marshal the whole message as fallback
	if msg.StopReason == "error" {
		if b, err := json.Marshal(msg); err == nil {
			return string(b)
		}
	}
	return ""
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func (g *AgentGateway) Start(ctx context.Context) error {
	return nil
}

func (g *AgentGateway) RunStagePersist(ctx context.Context, prompt string) (*StageOutput, error) {
	return nil, fmt.Errorf("not implemented: use RunStage (one-shot) for now")
}

func (g *AgentGateway) Steer(ctx context.Context, msg string) error {
	return fmt.Errorf("not implemented")
}

func (g *AgentGateway) Abort() error {
	return fmt.Errorf("not implemented")
}

func (g *AgentGateway) Stop() {}
