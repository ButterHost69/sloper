package session

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ButterHost69/sloper/internal/logger"
	"go.uber.org/zap"
)

func CleanupSessions(sessionDir, issuePattern string) error {
	if sessionDir == "" || issuePattern == "" {
		return nil
	}

	log := logger.Default()

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var deleted int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		if strings.Contains(name, issuePattern) {
			path := filepath.Join(sessionDir, name)
			if err := os.Remove(path); err != nil {
				log.Warn("session: failed to delete session file",
					zap.String("path", path),
					zap.Error(err))
			} else {
				log.Info("session: deleted session file",
					zap.String("path", path))
				deleted++
			}
		}
	}

	if deleted > 0 {
		log.Info("session: cleaned up sessions",
			zap.Int("deleted", deleted),
			zap.String("pattern", issuePattern))
	}

	return nil
}
