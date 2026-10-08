package slash

import (
	"regexp"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

type CommandType string

const (
	CmdApprove CommandType = "approve"
	CmdAbort   CommandType = "abort"
	CmdStatus  CommandType = "status"
	CmdRetry   CommandType = "retry"
	CmdSpec    CommandType = "spec"
	// CmdRevise is a legacy alias for CmdSpec: the spec rewrite already folds in
	// the whole conversation, so targeted feedback needs no separate path.
	CmdRevise CommandType = "revise"
)

type Command struct {
	Type      CommandType
	Author    string
	CommentID int64
	Body      string
}

var slashPattern = regexp.MustCompile(`(?im)^\s*/sloper\s+(\w+)\s*(.*)$`)

func ParseComments(comments []models.CommentInfo) []Command {
	var cmds []Command
	for _, c := range comments {
		cmd := parseComment(c)
		if cmd != nil {
			cmds = append(cmds, *cmd)
		}
	}
	return cmds
}

func parseComment(c models.CommentInfo) *Command {
	match := slashPattern.FindStringSubmatch(c.Body)
	if match == nil {
		return nil
	}

	cmdType := CommandType(strings.ToLower(strings.TrimSpace(match[1])))
	if cmdType == CmdRevise {
		cmdType = CmdSpec
	}

	switch cmdType {
	case CmdApprove, CmdAbort, CmdStatus, CmdRetry, CmdSpec:
		return &Command{
			Type:      cmdType,
			Author:    c.Author,
			CommentID: c.ID,
			Body:      c.Body,
		}
	default:
		return nil
	}
}

func IsValidCommand(body string) bool {
	return slashPattern.MatchString(body)
}
