package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/utils"
)

// rpcClient manages a single pi --mode rpc subprocess.
//
// Lifecycle:  NewRPCClient → (use) → Shutdown
//
// It is safe for concurrent use: SendCommand and Subscribe are
// goroutine-safe.  The internal readLoop runs until the process exits or
// Shutdown is called.
type rpcClient struct {
	cmd    *exec.Cmd
	stdin  *frameWriter
	stdout *frameReader

	// subscribers receive every event (except command responses).
	subs   map[uint64]chan models.AgentEvent
	subsMu sync.RWMutex
	subSeq atomic.Uint64

	// pending maps command id → response channel.
	pending   map[string]chan models.AgentEvent
	pendingMu sync.Mutex

	// lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{} // closed when readLoop exits

	// config (immutable after construction)
	opts models.AgentOptions

	// health
	lastEvent atomic.Int64 // UnixNano
}

// newRPCClient spawns pi --mode rpc and returns a ready client.
func newRPCClient(ctx context.Context, opts models.AgentOptions) (*rpcClient, error) {
	ctx, cancel := context.WithCancel(ctx)

	binary := opts.BinaryPath
	if binary == "" {
		binary = "pi"
	}

	args := []string{"--mode", "rpc"}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts.Thinking != "" {
		args = append(args, "--thinking", opts.Thinking)
	}
	// V1: no session persistence; add --session later for Variant C.
	args = append(args, "--no-session")

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = opts.CWD

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("agent rpc: stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("agent rpc: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("agent rpc: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("agent rpc: start %s: %w", binary, err)
	}

	c := &rpcClient{
		cmd:     cmd,
		stdin:   newFrameWriter(stdinPipe),
		stdout:  newFrameReader(stdoutPipe),
		subs:    make(map[uint64]chan models.AgentEvent),
		pending: make(map[string]chan models.AgentEvent),
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
		opts:    opts,
	}
	c.lastEvent.Store(time.Now().UnixNano())

	go c.readLoop()
	go c.drainStderr(stderrPipe)

	return c, nil
}

// SendCommand writes a command and blocks until the matching response
// arrives (or ctx is cancelled, or the process dies).
func (c *rpcClient) SendCommand(ctx context.Context, cmd rpcCmd) (*models.AgentEvent, error) {
	if cmd.ID == "" {
		cmd.ID = utils.GenerateRandomID()
	}

	respCh := make(chan models.AgentEvent, 1)
	c.pendingMu.Lock()
	c.pending[cmd.ID] = respCh
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, cmd.ID)
		c.pendingMu.Unlock()
	}()

	if err := c.stdin.Write(cmd); err != nil {
		return nil, fmt.Errorf("agent rpc: write command %q: %w", cmd.Type, err)
	}

	select {
	case resp := <-respCh:
		if resp.Success != nil && !*resp.Success {
			return &resp, fmt.Errorf("agent rpc: command %q failed: %s", cmd.Type, resp.Error)
		}
		return &resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, fmt.Errorf("agent rpc: pi process exited while waiting for %q response", cmd.Type)
	}
}

// Subscribe returns a channel that receives every event from the agent
// (except command responses, which are routed to SendCommand callers).
// Call the returned function to unsubscribe.
func (c *rpcClient) Subscribe(bufSize int) (<-chan models.AgentEvent, func()) {
	ch := make(chan models.AgentEvent, bufSize)
	id := c.subSeq.Add(1)

	c.subsMu.Lock()
	c.subs[id] = ch
	c.subsMu.Unlock()

	unsub := func() {
		c.subsMu.Lock()
		delete(c.subs, id)
		c.subsMu.Unlock()
	}
	return ch, unsub
}

// Done returns a channel that is closed when the readLoop exits.
func (c *rpcClient) Done() <-chan struct{} { return c.done }

// IsHealthy returns false if no event has been received for 60 s.
func (c *rpcClient) IsHealthy() bool {
	return time.Since(time.Unix(0, c.lastEvent.Load())) < 60*time.Second
}

// Shutdown gracefully stops the pi process.
func (c *rpcClient) Shutdown() {
	// 1. Abort any in-flight LLM call.
	_ = c.stdin.Write(newRPCCmd("abort", ""))
	time.Sleep(200 * time.Millisecond)

	// 2. Close stdin so pi knows we're done.
	if closer, ok := c.stdin.w.(io.Closer); ok {
		_ = closer.Close()
	}

	// 3. Cancel context → SIGKILL via exec.CommandContext.
	c.cancel()

	// 4. Wait for readLoop to exit.
	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
		if c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
	}

	_ = c.cmd.Wait()
}

// ─── internal goroutines ────────────────────────────────────────────

func (c *rpcClient) readLoop() {
	defer close(c.done)

	for {
		var event models.AgentEvent
		if err := c.stdout.Decode(&event); err != nil {
			if c.ctx.Err() != nil {
				return // normal shutdown
			}
			// Malformed line – log and try to continue.
			if !isEOF(err) {
				log.Printf("[agent rpc] decode error: %v", err)
			}
			return
		}

		c.lastEvent.Store(time.Now().UnixNano())

		// Route command responses to the caller that is waiting.
		if event.Command != "" && event.ID != "" {
			c.pendingMu.Lock()
			ch, ok := c.pending[event.ID]
			c.pendingMu.Unlock()
			if ok {
				select {
				case ch <- event:
				default:
				}
				continue // don't broadcast responses
			}
		}

		// Broadcast to all subscribers.
		c.subsMu.RLock()
		for _, ch := range c.subs {
			select {
			case ch <- event:
			default: // slow consumer – drop
			}
		}
		c.subsMu.RUnlock()
	}
}

func (c *rpcClient) drainStderr(r io.Reader) {
	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	if s := strings.TrimSpace(buf.String()); s != "" {
		log.Printf("[agent rpc] stderr: %s", s)
	}
}

func isEOF(err error) bool {
	return err == io.EOF || strings.Contains(err.Error(), "file already closed")
}
