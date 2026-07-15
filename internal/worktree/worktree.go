package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ButterHost69/sloper/internal/git"
	"github.com/ButterHost69/sloper/internal/logger"
	"go.uber.org/zap"
)

const (
	DefaultWorktreeDir = ".sloper/worktrees"
)

type Manager struct {
	baseDir string
	git     *git.GitGateway
	mu      sync.Mutex
}

func NewManager(baseDir string, gitGateway *git.GitGateway) *Manager {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = os.TempDir()
		}
		baseDir = filepath.Join(homeDir, DefaultWorktreeDir)
	}
	return &Manager{
		baseDir: baseDir,
		git:     gitGateway,
	}
}

func (m *Manager) BaseDir() string { return m.baseDir }

func (m *Manager) EnsureBaseDir() error {
	return os.MkdirAll(m.baseDir, 0o755)
}

func (m *Manager) worktreePath(name string) string {
	return filepath.Join(m.baseDir, name)
}

func (m *Manager) CreateWithBranch(ctx context.Context, repoPath, branchName, base string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log := logger.Default()

	if err := m.EnsureBaseDir(); err != nil {
		return "", fmt.Errorf("worktree: create base dir: %w", err)
	}

	wtPath := m.worktreePath(branchName)

	// Clean up any stale directory (leftover from a crashed previous run)
	if _, err := os.Stat(wtPath); err == nil {
		log.Warn("worktree: stale directory found, removing",
			zap.String("path", wtPath))
		_ = m.git.RemoveWorktree(ctx, repoPath, wtPath, true)
		_ = os.RemoveAll(wtPath)
	}

	branchExists, _ := m.git.BranchExists(ctx, repoPath, branchName)

	if branchExists {
		log.Info("worktree: branch exists, creating worktree without -b",
			zap.String("branch", branchName),
			zap.String("path", wtPath))

		if err := m.git.AddWorktree(ctx, repoPath, wtPath, branchName, ""); err != nil {
			return "", fmt.Errorf("worktree: add existing branch %s: %w", branchName, err)
		}
	} else {
		log.Info("worktree: creating with new branch",
			zap.String("branch", branchName),
			zap.String("base", base),
			zap.String("path", wtPath))

		if err := m.git.AddWorktree(ctx, repoPath, wtPath, base, branchName); err != nil {
			return "", fmt.Errorf("worktree: add (branch %s from %s): %w", branchName, base, err)
		}
	}

	return wtPath, nil
}

func (m *Manager) CreateAtCommit(ctx context.Context, repoPath, name, ref string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log := logger.Default()

	if err := m.EnsureBaseDir(); err != nil {
		return "", fmt.Errorf("worktree: create base dir: %w", err)
	}

	wtPath := m.worktreePath(name)
	if _, err := os.Stat(wtPath); err == nil {
		return "", fmt.Errorf("worktree: path already exists: %s", wtPath)
	}

	log.Info("worktree: creating detached at ref",
		zap.String("ref", ref),
		zap.String("path", wtPath))

	if err := m.git.AddWorktreeDetached(ctx, repoPath, wtPath, ref); err != nil {
		return "", fmt.Errorf("worktree: add detached at %s: %w", ref, err)
	}

	return wtPath, nil
}

func (m *Manager) Remove(ctx context.Context, repoPath, wtPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log := logger.Default()
	log.Info("worktree: removing", zap.String("path", wtPath))

	if err := m.git.RemoveWorktree(ctx, repoPath, wtPath, true); err != nil {
		_ = os.RemoveAll(wtPath)
		return fmt.Errorf("worktree: remove %s: %w", wtPath, err)
	}

	return nil
}

func (m *Manager) CleanupAll(ctx context.Context, repoPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log := logger.Default()

	list, err := m.git.ListWorktrees(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("worktree: cleanup list: %w", err)
	}

	for _, line := range strings.Split(list, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			if strings.HasPrefix(path, m.baseDir) {
				log.Info("worktree: cleanup removing", zap.String("path", path))
				_ = m.git.RemoveWorktree(ctx, repoPath, path, true)
			}
		}
	}

	_ = m.git.PruneWorktrees(ctx, repoPath)

	entries, _ := os.ReadDir(m.baseDir)
	for _, entry := range entries {
		fullPath := filepath.Join(m.baseDir, entry.Name())
		if _, err := os.Stat(fullPath); err == nil {
			log.Info("worktree: cleanup removing orphan dir", zap.String("path", fullPath))
			_ = os.RemoveAll(fullPath)
		}
	}

	return nil
}

func (m *Manager) BranchNameForIssue(issueNumber int64, slug string) string {
	if slug == "" {
		return fmt.Sprintf("sloper/issue-%d", issueNumber)
	}
	return fmt.Sprintf("sloper/issue-%d-%s", issueNumber, slug)
}

func (m *Manager) ReviewWorktreeName(issueNumber int64, prNumber int64) string {
	return fmt.Sprintf("review-%d-pr-%d", issueNumber, prNumber)
}
