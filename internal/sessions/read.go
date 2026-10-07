package sessions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	// maxTailBytes bounds how much of a session file a single read pulls into
	// memory for the backwards ("load earlier") path.
	maxTailBytes = 8 << 20
	// maxEntryBytes guards against a single pathological line.
	maxEntryBytes = 4 << 20
	// DefaultPageSize is the transcript window returned when a caller does not
	// ask for a specific size.
	DefaultPageSize = 400
	// MaxPageSize caps one response.
	MaxPageSize = 2000
)

// Entry is one JSONL line of a pi session file, normalized enough for the
// console to render it without knowing every version's schema. Raw keeps the
// original line so unknown entry types can still be inspected.
type Entry struct {
	Type       string          `json:"type"`
	ID         string          `json:"id,omitempty"`
	ParentID   string          `json:"parent_id,omitempty"`
	Timestamp  string          `json:"timestamp,omitempty"`
	Message    *Message        `json:"message,omitempty"`
	Provider   string          `json:"provider,omitempty"`
	ModelID    string          `json:"model_id,omitempty"`
	Level      string          `json:"level,omitempty"`
	Summary    string          `json:"summary,omitempty"`
	Name       string          `json:"name,omitempty"`
	Label      string          `json:"label,omitempty"`
	TargetID   string          `json:"target_id,omitempty"`
	CustomType string          `json:"custom_type,omitempty"`
	FromID     string          `json:"from_id,omitempty"`
	Raw        json.RawMessage `json:"raw,omitempty"`
}

// Message is the nested agent message of a "message" entry.
type Message struct {
	Role         string        `json:"role"`
	Content      []ContentPart `json:"content,omitempty"`
	Provider     string        `json:"provider,omitempty"`
	Model        string        `json:"model,omitempty"`
	Usage        *Usage        `json:"usage,omitempty"`
	StopReason   string        `json:"stopReason,omitempty"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
	ToolCallID   string        `json:"toolCallId,omitempty"`
	ToolName     string        `json:"toolName,omitempty"`
	IsError      bool          `json:"isError,omitempty"`
	Timestamp    int64         `json:"timestamp,omitempty"`
	// bashExecution messages (a user running "!cmd") carry their own fields.
	Command   string `json:"command,omitempty"`
	Output    string `json:"output,omitempty"`
	ExitCode  *int   `json:"exitCode,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

// ContentPart is one block of a message: text, thinking, a tool call, or an
// image. Fields that do not apply to a part's type stay empty.
type ContentPart struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	MimeType  string          `json:"mimeType,omitempty"`
	Data      string          `json:"data,omitempty"`
}

// Usage mirrors pi's per-assistant-message accounting.
type Usage struct {
	Input       int64 `json:"input"`
	Output      int64 `json:"output"`
	CacheRead   int64 `json:"cacheRead"`
	CacheWrite  int64 `json:"cacheWrite"`
	Reasoning   int64 `json:"reasoning"`
	TotalTokens int64 `json:"totalTokens"`
	Cost        *Cost `json:"cost,omitempty"`
}

// Cost is the provider-reported cost breakdown.
type Cost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}

// UnmarshalJSON accepts both the array form and the plain-string form that
// some pi versions write for user messages.
func (m *Message) UnmarshalJSON(data []byte) error {
	type messageAlias Message
	var wire struct {
		*messageAlias
		Content json.RawMessage `json:"content"`
	}
	wire.messageAlias = (*messageAlias)(m)
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if len(wire.Content) == 0 {
		return nil
	}
	m.Content = nil
	if err := json.Unmarshal(wire.Content, &m.Content); err == nil {
		return nil
	}
	var text string
	if err := json.Unmarshal(wire.Content, &text); err == nil {
		m.Content = []ContentPart{{Type: "text", Text: text}}
	}
	return nil
}

// Chunk is one incremental read of a session file.
type Chunk struct {
	Entries     []Entry `json:"entries"`
	StartOffset int64   `json:"start_offset"`
	NextOffset  int64   `json:"next_offset"`
	Size        int64   `json:"size_bytes"`
	ModifiedAt  string  `json:"modified_at"`
	// Reset means the requested offset is past the end of the file (it was
	// rewritten or replaced); the caller should restart from zero.
	Reset bool `json:"reset"`
	// HasMore means the file had more entries than the requested page.
	HasMore bool `json:"has_more"`
	// HasMoreBefore means entries exist before StartOffset.
	HasMoreBefore bool `json:"has_more_before"`
	Skipped       int  `json:"skipped"`
}

type entryWire struct {
	Type          string   `json:"type"`
	ID            string   `json:"id"`
	ParentID      *string  `json:"parentId"`
	Timestamp     string   `json:"timestamp"`
	Message       *Message `json:"message"`
	Provider      string   `json:"provider"`
	ModelID       string   `json:"modelId"`
	ThinkingLevel string   `json:"thinkingLevel"`
	Summary       string   `json:"summary"`
	Name          string   `json:"name"`
	Label         string   `json:"label"`
	TargetID      string   `json:"targetId"`
	CustomType    string   `json:"customType"`
	FromID        string   `json:"fromId"`
}

func parseEntry(line []byte) (Entry, bool) {
	var w entryWire
	if err := json.Unmarshal(line, &w); err != nil || w.Type == "" {
		return Entry{}, false
	}
	raw := make([]byte, len(line))
	copy(raw, line)

	e := Entry{
		Type:       w.Type,
		ID:         w.ID,
		Timestamp:  w.Timestamp,
		Message:    w.Message,
		Provider:   w.Provider,
		ModelID:    w.ModelID,
		Level:      w.ThinkingLevel,
		Summary:    w.Summary,
		Name:       w.Name,
		Label:      w.Label,
		TargetID:   w.TargetID,
		CustomType: w.CustomType,
		FromID:     w.FromID,
		Raw:        raw,
	}
	if w.ParentID != nil {
		e.ParentID = *w.ParentID
	}
	return e, true
}

// ReadAfter returns up to limit entries that start at byte offset. A trailing
// partial line is left unconsumed so the next poll re-reads it once pi has
// finished writing it.
func ReadAfter(path string, offset int64, limit int) (*Chunk, error) {
	return readAfter(path, offset, clampLimit(limit))
}

// ReadAll returns the whole file. It is used for summaries, which need every
// message to total tokens and tool calls correctly.
func ReadAll(path string) (*Chunk, error) {
	return readAfter(path, 0, 0)
}

// readAfter is the unbounded reader; limit <= 0 means "until EOF".
func readAfter(path string, offset int64, limit int) (*Chunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("sessions: open %s: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("sessions: stat %s: %w", path, err)
	}
	chunk := &Chunk{
		Entries:    make([]Entry, 0),
		Size:       info.Size(),
		ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
	}
	if offset < 0 {
		offset = 0
	}
	if offset > info.Size() {
		chunk.Reset = true
		chunk.NextOffset = 0
		chunk.HasMoreBefore = false
		return chunk, nil
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("sessions: seek %s: %w", path, err)
	}

	chunk.StartOffset = offset
	chunk.NextOffset = offset

	reader := bufio.NewReaderSize(f, 64<<10)
	pos := offset
	for limit <= 0 || len(chunk.Entries) < limit {
		line, readErr := reader.ReadBytes('\n')
		if len(line) == 0 {
			break
		}
		if line[len(line)-1] != '\n' {
			// pi has not finished this entry yet.
			break
		}
		pos += int64(len(line))
		chunk.NextOffset = pos

		trimmed := bytes.TrimRight(line[:len(line)-1], "\r")
		if len(bytes.TrimSpace(trimmed)) == 0 {
			continue
		}
		if len(trimmed) > maxEntryBytes {
			chunk.Skipped++
			continue
		}
		entry, ok := parseEntry(trimmed)
		if !ok {
			chunk.Skipped++
			continue
		}
		chunk.Entries = append(chunk.Entries, entry)
		if readErr != nil {
			break
		}
	}
	chunk.HasMore = limit > 0 && len(chunk.Entries) >= limit && chunk.NextOffset < info.Size()
	return chunk, nil
}

// ReadBefore returns up to limit entries that end at byte offset (exclusive),
// for scrolling back through a transcript.
func ReadBefore(path string, before int64, limit int) (*Chunk, error) {
	limit = clampLimit(limit)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("sessions: stat %s: %w", path, err)
	}
	size := info.Size()
	if before <= 0 || before > size {
		before = size
	}

	chunk := &Chunk{
		Entries:    make([]Entry, 0),
		NextOffset: before,
		Size:       size,
		ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
	}

	start := int64(0)
	if before > maxTailBytes {
		start = before - maxTailBytes
	}
	data := make([]byte, before-start)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("sessions: open %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return nil, fmt.Errorf("sessions: seek %s: %w", path, err)
	}
	if _, err := io.ReadFull(f, data); err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("sessions: read %s: %w", path, err)
	}

	lines := completeLines(data)
	// The first line of a mid-file window is almost certainly cut in half.
	if start > 0 && len(lines) > 0 {
		lines = lines[1:]
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	if len(lines) == 0 {
		chunk.StartOffset = before
		chunk.HasMoreBefore = start > 0
		return chunk, nil
	}

	chunk.StartOffset = start + int64(lines[0].start)
	chunk.HasMoreBefore = chunk.StartOffset > 0
	// Stop the forward cursor at the last complete line so a partial tail is
	// picked up by the next ReadAfter call.
	chunk.NextOffset = start + int64(lines[len(lines)-1].end)
	for _, span := range lines {
		raw := bytes.TrimRight(data[span.start:span.end], "\r\n")
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		if len(raw) > maxEntryBytes {
			chunk.Skipped++
			continue
		}
		entry, ok := parseEntry(raw)
		if !ok {
			chunk.Skipped++
			continue
		}
		chunk.Entries = append(chunk.Entries, entry)
	}
	return chunk, nil
}

// ReadLast returns the tail of a session: the newest entries plus the offsets
// needed to page backwards and to keep polling forwards.
func ReadLast(path string, limit int) (*Chunk, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("sessions: stat %s: %w", path, err)
	}
	return ReadBefore(path, info.Size(), limit)
}

type lineSpan struct {
	start int
	end   int
}

// completeLines returns the offsets of every newline-terminated line in data.
func completeLines(data []byte) []lineSpan {
	spans := make([]lineSpan, 0, bytes.Count(data, []byte{'\n'}))
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}
		spans = append(spans, lineSpan{start: start, end: i + 1})
		start = i + 1
	}
	return spans
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return DefaultPageSize
	}
	if limit > MaxPageSize {
		return MaxPageSize
	}
	return limit
}
