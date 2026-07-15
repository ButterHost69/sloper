package models

const (
	TRIAGED_LABEL = "triaged"

	// Pipeline stages stored in the issues.stage column
	StageNew        = "new"
	StageSpecDone   = "spec-done"
	StageApproved   = "approved"
	StageWorkDone   = "work-done"
	StageReviewDone = "review-done"
	StageMerged     = "merged"
	StageFailed     = "failed"
)

const (
	MaxReviewIterations = 3
)
