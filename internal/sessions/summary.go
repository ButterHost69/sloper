package sessions

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
)

// ToolRef is one tool call in the transcript, with the status of its result.
type ToolRef struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Preview string `json:"preview,omitempty"`
	Status  string `json:"status"` // pending | ok | error
}

// Current is the console's answer to "what is happening right now", derived
// from the newest entries: a pending tool call, a fresh model turn, or nothing.
type Current struct {
	Kind  string    `json:"kind"` // starting | waiting | thinking | tool | error
	Text  string    `json:"text,omitempty"`
	Since string    `json:"since,omitempty"`
	Tools []ToolRef `json:"tools,omitempty"`
}

// Summary is the cheap per-session view used by the sessions list.
type Summary struct {
	Title             string    `json:"title,omitempty"`
	Model             string    `json:"model,omitempty"`
	Provider          string    `json:"provider,omitempty"`
	Messages          int       `json:"messages"`
	UserMessages      int       `json:"user_messages"`
	AssistantMessages int       `json:"assistant_messages"`
	ToolResults       int       `json:"tool_results"`
	ToolCalls         int       `json:"tool_calls"`
	ToolErrors        int       `json:"tool_errors"`
	ThinkingChars     int       `json:"thinking_chars"`
	OutputChars       int       `json:"output_chars"`
	TotalTokens       int64     `json:"total_tokens"`
	CostUSD           float64   `json:"cost_usd"`
	Compactions       int       `json:"compactions"`
	StartedAt         string    `json:"started_at,omitempty"`
	LastEntryAt       string    `json:"last_entry_at,omitempty"`
	LastText          string    `json:"last_text,omitempty"`
	Current           *Current  `json:"current,omitempty"`
	Tools             []ToolRef `json:"tools,omitempty"`
}

// Summarize folds a session's entries into the list-view summary. It is O(n)
// over the transcript, so callers cache it by file size and mtime.
func Summarize(entries []Entry) Summary {
	var s Summary

	tools := make(map[string]*ToolRef)
	var lastTools []ToolRef
	var lastRole string
	var lastEntryAt string

	for i := range entries {
		e := entries[i]
		if e.Timestamp != "" {
			lastEntryAt = e.Timestamp
		}
		switch e.Type {
		case "session":
			s.StartedAt = e.Timestamp
		case "model_change":
			if e.Provider != "" || e.ModelID != "" {
				s.Provider, s.Model = e.Provider, e.ModelID
			}
		case "compaction":
			s.Compactions++
		case "message":
			if e.Message == nil {
				continue
			}
			lastRole = e.Message.Role
			s.Messages++
			if s.StartedAt == "" {
				s.StartedAt = e.Timestamp
			}
			switch e.Message.Role {
			case "user":
				s.UserMessages++
				if s.Title == "" {
					s.Title = truncate(firstLine(messageText(e.Message)), 140)
				}
			case "assistant":
				s.AssistantMessages++
				if e.Message.Model != "" {
					s.Model = e.Message.Model
				}
				if e.Message.Provider != "" {
					s.Provider = e.Message.Provider
				}
				applyUsage(&s, e.Message.Usage)
				lastTools = lastTools[:0]
				for _, part := range e.Message.Content {
					switch part.Type {
					case "text":
						s.OutputChars += len(part.Text)
						if text := strings.TrimSpace(part.Text); text != "" {
							s.LastText = truncate(text, 400)
						}
					case "thinking":
						s.ThinkingChars += len(part.Thinking)
					case "toolCall":
						s.ToolCalls++
						ref := ToolRef{
							ID:      part.ID,
							Name:    part.Name,
							Preview: ToolPreview(part.Name, part.Arguments),
							Status:  "pending",
						}
						lastTools = append(lastTools, ref)
						if part.ID != "" {
							copied := ref
							tools[part.ID] = &copied
						}
					}
				}
			case "toolResult":
				s.ToolResults++
				if e.Message.IsError {
					s.ToolErrors++
				}
				if ref, ok := tools[e.Message.ToolCallID]; ok {
					if e.Message.IsError {
						ref.Status = "error"
					} else {
						ref.Status = "ok"
					}
				}
			}
		}
	}

	s.LastEntryAt = lastEntryAt
	for _, ref := range lastTools {
		if known, ok := tools[ref.ID]; ok {
			ref.Status = known.Status
		}
		s.Tools = append(s.Tools, ref)
	}
	s.Current = currentActivity(lastRole, lastTools, tools, entries)
	return s
}

func applyUsage(s *Summary, usage *Usage) {
	if usage == nil {
		return
	}
	s.TotalTokens += usage.TotalTokens
	if usage.Cost != nil {
		s.CostUSD += usage.Cost.Total
	}
}

// currentActivity describes the newest step. pi writes an assistant message
// (with its tool calls) before the tools run, so a tool call without a matching
// result is exactly "this tool is running now".
func currentActivity(lastRole string, lastTools []ToolRef, tools map[string]*ToolRef, entries []Entry) *Current {
	if lastRole == "" {
		if len(entries) == 0 {
			return nil
		}
		return &Current{Kind: "starting", Text: "Session created, no messages yet"}
	}

	since := ""
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Type == "message" {
			since = entries[i].Timestamp
			break
		}
	}

	switch lastRole {
	case "user":
		return &Current{Kind: "waiting", Text: "Prompt sent, waiting for the model", Since: since}
	case "assistant":
		if last := lastMessage(entries); last != nil && last.Message != nil {
			if last.Message.StopReason == "error" {
				text := last.Message.ErrorMessage
				if text == "" {
					text = "The model returned an error"
				}
				return &Current{Kind: "error", Text: truncate(text, 300), Since: since}
			}
		}
		pending := make([]ToolRef, 0, len(lastTools))
		for _, ref := range lastTools {
			status := ref.Status
			if known, ok := tools[ref.ID]; ok {
				status = known.Status
			}
			if status == "pending" {
				pending = append(pending, ref)
			}
		}
		if len(pending) > 0 {
			return &Current{Kind: "tool", Tools: pending, Since: since}
		}
		return &Current{Kind: "thinking", Text: "Model turn finished", Since: since}
	default: // toolResult and anything else
		return &Current{Kind: "thinking", Text: "Reading tool results", Since: since}
	}
}

func lastMessage(entries []Entry) *Entry {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Type == "message" {
			return &entries[i]
		}
	}
	return nil
}

// messageText joins the text parts of a message.
func messageText(m *Message) string {
	var parts []string
	for _, part := range m.Content {
		if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// ToolPreview renders a short, single-line hint of what a tool call is doing.
func ToolPreview(name string, args json.RawMessage) string {
	if len(args) == 0 {
		return ""
	}
	var fields map[string]any
	if err := json.Unmarshal(args, &fields); err != nil {
		return ""
	}
	pick := func(keys ...string) string {
		for _, key := range keys {
			if value, ok := fields[key].(string); ok && strings.TrimSpace(value) != "" {
				return value
			}
		}
		return ""
	}

	var preview string
	switch name {
	case "bash":
		preview = pick("command")
	case "read", "write", "edit":
		preview = pick("path", "file_path", "filePath")
	}
	if preview == "" {
		preview = pick("path", "file_path", "filePath", "command", "query", "pattern", "url", "prompt", "message")
	}
	if preview == "" {
		preview = firstStringValue(fields)
	}
	return truncate(firstLine(preview), 200)
}

// firstStringValue returns the shortest string field, which reads better as a
// preview than a JSON blob.
func firstStringValue(fields map[string]any) string {
	values := make([]string, 0, len(fields))
	for _, value := range fields {
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			values = append(values, text)
		}
	}
	sort.Slice(values, func(i, j int) bool { return len(values[i]) < len(values[j]) })
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max]) + "…"
}

// Cache memoizes summaries by file identity (size + mtime) so polling the
// sessions list does not re-parse every transcript.
type Cache struct {
	mu     sync.Mutex
	byPath map[string]cacheItem
}

type cacheItem struct {
	size    int64
	modNano int64
	summary Summary
}

func NewCache() *Cache {
	return &Cache{byPath: make(map[string]cacheItem)}
}

// SummaryFor returns the cached summary, recomputing it when the file changed.
func (c *Cache) SummaryFor(f File) (Summary, error) {
	key := f.Path
	modNano := f.ModTime.UnixNano()

	c.mu.Lock()
	item, ok := c.byPath[key]
	c.mu.Unlock()
	if ok && item.size == f.Size && item.modNano == modNano {
		return item.summary, nil
	}

	chunk, err := ReadAll(f.Path)
	if err != nil {
		return Summary{}, err
	}
	summary := Summarize(chunk.Entries)
	c.Store(f, summary)
	return summary, nil
}

// Store records a summary that a caller already computed.
func (c *Cache) Store(f File, summary Summary) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byPath[f.Path] = cacheItem{size: f.Size, modNano: f.ModTime.UnixNano(), summary: summary}
}

// Retain drops cached summaries for files that are no longer on disk.
func (c *Cache) Retain(paths []string) {
	keep := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		keep[path] = struct{}{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for path := range c.byPath {
		if _, ok := keep[path]; !ok {
			delete(c.byPath, path)
		}
	}
}
