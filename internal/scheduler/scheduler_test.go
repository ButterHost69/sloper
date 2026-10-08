package scheduler

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/storage"
)

func newTestScheduler(t *testing.T) *Scheduler {
	t.Helper()

	db, err := storage.OpenDB(filepath.Join(t.TempDir(), "sloper.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &Scheduler{db: storage.NewRepositories(db)}
}

func seedIssue(t *testing.T, s *Scheduler, issueNumber int64, spec *models.SpecResult) {
	t.Helper()

	rec := storage.IssueRecordFromModel(models.IssueDetail{
		Number:    issueNumber,
		Title:     "spec rewrite",
		UpdatedAt: "2026-08-04T21:39:37Z",
	}, models.StageSpecDone)
	rec.SpecJSON = storage.SpecJSON(spec)
	if err := s.db.UpsertIssue(context.Background(), rec); err != nil {
		t.Fatalf("upsert issue: %v", err)
	}
}

func TestStoredSpecReturnsTheCompleteStoredSpec(t *testing.T) {
	s := newTestScheduler(t)
	seedIssue(t, s, 5, &models.SpecResult{
		Summary:            "Spec v1 summary",
		FilesToChange:      []string{"internal/scheduler/scheduler.go"},
		ImplementationPlan: "Spec v1 implementation plan",
	})

	spec := s.storedSpec(context.Background(), 5)
	if spec == nil {
		t.Fatal("storedSpec returned nil for a complete stored spec")
	}
	if spec.Summary != "Spec v1 summary" || spec.ImplementationPlan != "Spec v1 implementation plan" {
		t.Fatalf("storedSpec returned the wrong spec: %+v", spec)
	}
	if !s.hasValidSpec(context.Background(), 5) {
		t.Error("hasValidSpec = false for a complete stored spec")
	}
}

func TestStoredSpecIsNilWhenMissingOrIncomplete(t *testing.T) {
	s := newTestScheduler(t)
	ctx := context.Background()

	if spec := s.storedSpec(ctx, 404); spec != nil {
		t.Errorf("storedSpec = %+v for an unknown issue, want nil", spec)
	}

	seedIssue(t, s, 5, &models.SpecResult{Summary: "summary without a plan"})
	if spec := s.storedSpec(ctx, 5); spec != nil {
		t.Errorf("storedSpec = %+v for an incomplete spec, want nil", spec)
	}
	if s.hasValidSpec(ctx, 5) {
		t.Error("hasValidSpec = true for an incomplete spec")
	}
}

func TestFormatSpecCommentMarksRewrites(t *testing.T) {
	spec := &models.SpecResult{
		Summary:            "Spec v2 summary",
		FilesToChange:      []string{"internal/scheduler/scheduler.go"},
		ImplementationPlan: "Spec v2 implementation plan",
	}

	first := formatSpecComment(spec, 5, false)
	if strings.Contains(first, "supersedes") {
		t.Error("first spec comment claims to supersede a previous spec")
	}
	if !strings.Contains(first, "/sloper approve") {
		t.Error("spec comment is missing the approve hint")
	}

	rewritten := formatSpecComment(spec, 5, true)
	if !strings.Contains(rewritten, "supersedes the previous spec") {
		t.Error("rewritten spec comment does not say it supersedes the previous spec")
	}
	if !strings.Contains(rewritten, "Spec v2 implementation plan") {
		t.Error("rewritten spec comment is missing the implementation plan")
	}
}
