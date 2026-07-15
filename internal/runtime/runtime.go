package runtime

import (
	"context"
	"os"

	"github.com/ButterHost69/sloper/internal/git"
	"github.com/ButterHost69/sloper/internal/logger"
	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/scheduler"
	"github.com/ButterHost69/sloper/internal/storage"
	"github.com/ButterHost69/sloper/internal/worktree"
	"go.uber.org/zap"
)

type Runtime struct {
	config models.RuntimeConfig
}

func NewRuntime(repoPath string) *Runtime {
	config := models.RuntimeConfig{
		RepoPath: repoPath,
	}
	return &Runtime{config: config}
}

func (r *Runtime) Start(ctx context.Context) {
	log := logger.Default()

	if err := logger.Init(""); err != nil {
		log.Warn("runtime: failed to init file logger, using console only", zap.Error(err))
	}
	log = logger.Default()

	log.Info("runtime: starting sloper", zap.String("repo", r.config.RepoPath))

	dbPath := os.Getenv("SLOPER_DB_PATH")
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		log.Error("runtime: open database failed", zap.Error(err))
		return
	}
	defer db.Close()

	if err := storage.Migrate(ctx, db); err != nil {
		log.Error("runtime: database migration failed", zap.Error(err))
		return
	}

	repos := storage.NewRepositories(db)

	r.runRecovery(ctx, repos, log)

	wtMgr := worktree.NewManager("", git.New(models.GitGatewayOptions{}))
	if err := wtMgr.CleanupAll(ctx, r.config.RepoPath); err != nil {
		log.Warn("runtime: worktree cleanup failed", zap.Error(err))
	}

	sched := scheduler.New(r.config.RepoPath, repos)
	sched.Start(ctx)

	logger.Sync()
}

func (r *Runtime) runRecovery(ctx context.Context, repos *storage.Repositories, log *zap.Logger) {
	n, err := repos.MarkRunsInterrupted(ctx)
	if err != nil {
		log.Warn("runtime: failed to mark interrupted runs", zap.Error(err))
	} else if n > 0 {
		log.Info("runtime: marked interrupted runs", zap.Int("count", n))
	}

	interrupted, err := repos.GetInterruptedRuns(ctx)
	if err != nil {
		log.Warn("runtime: failed to get interrupted runs", zap.Error(err))
		return
	}

	for _, run := range interrupted {
		log.Info("runtime: recovering interrupted run",
			logger.WithIssue(run.IssueNumber),
			zap.String("stage", run.Stage),
			zap.Int64("run_id", run.ID))

		repos.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: run.IssueNumber,
			EventType:   "recovery.interrupted_run",
			Stage:       run.Stage,
			Message:     "run was interrupted, will retry on next tick",
		})
	}

	log.Info("runtime: recovery complete")
}
