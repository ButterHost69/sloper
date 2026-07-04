package runtime

import (
	"context"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/scheduler"
)

type Runtime struct {
	config	models.RuntimeConfig
}

// TODO: Add Something Related to Recovery, if an issues or a pr is stuck in between, we first resume that process before starting anything new
// TODO: Make it able to process N number of request // issues

func NewRuntime(repoPath string) (*Runtime ){
	config := models.RuntimeConfig{
		RepoPath: repoPath,
	}

	return &Runtime{
		config: config,
	}
}

func (r *Runtime) Start(ctx context.Context){
	// First make it be able to do the whole recovery thing

	// Connect to the mysql and use it as the main queue, and maybe something else, idk

	scheduler := scheduler.New(r.config.RepoPath)
	scheduler.Start(ctx)
	// Start the scheduler
		// Discovers Issues
		// Adds them to loops

	// Cleanly Exit Everything
}