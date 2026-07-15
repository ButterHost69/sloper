package slash

import (
	"regexp"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

type CommandType string

const (
	CmdApprove CommandType = "approve"
	CmdRevise  CommandType = "revise"
	CmdAbort   CommandType = "abort"
	CmdStatus  CommandType = "status"
	CmdRetry   CommandType = "retry"
)

type Command struct {
	Type      CommandType
	Feedback  string
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
	rest := strings.TrimSpace(match[2])

	switch cmdType {
	case CmdApprove, CmdAbort, CmdStatus, CmdRetry:
		return &Command{
			Type:      cmdType,
			Author:    c.Author,
			CommentID: c.ID,
			Body:      c.Body,
		}
	case CmdRevise:
		feedback := rest
		if feedback == "" {
			feedback = strings.TrimSpace(strings.TrimPrefix(c.Body, match[0]))
		}
		if feedback == "" {
			feedback = "(no specific feedback provided)"
		}
		return &Command{
			Type:      cmdType,
			Feedback:  feedback,
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
