package models

const (
	TRIAGED_LABEL = "triaged"

	// Pipeline stages stored in the issues.stage column
	StageNew         = "new" // ;; Imma make this redudant and old
	StageSpecOngoing = "spec-ongoing"
	StageSpecDone    = "spec-done"
	StageApproved    = "approved"
	StageWorkDone    = "work-done"
	StageReviewDone  = "review-done"
	StageMerged      = "merged"
	StageFailed      = "failed"
)

const (
	MaxReviewIterations = 3
)

var OUR_LABEL = []string{
	TRIAGED_LABEL,
}

// var NEW_ISSUE_LABELS = []string{}