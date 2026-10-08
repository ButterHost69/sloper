package slash

import (
	"testing"

	"github.com/ButterHost69/sloper/internal/models"
)

func TestParseCommentsAcceptsSupportedCommands(t *testing.T) {
	for _, want := range []CommandType{CmdApprove, CmdAbort, CmdStatus, CmdRetry, CmdSpec} {
		cmds := ParseComments([]models.CommentInfo{
			{ID: 1, Author: "butter", Body: "/sloper " + string(want)},
		})
		if len(cmds) != 1 {
			t.Fatalf("/sloper %s parsed into %d commands, want 1", want, len(cmds))
		}
		if cmds[0].Type != want {
			t.Errorf("/sloper %s parsed as %q", want, cmds[0].Type)
		}
	}
}

func TestParseCommentsRejectsUnsupportedCommands(t *testing.T) {
	for _, body := range []string{
		"/sloper review",
		"/sloper frobnicate",
		"just a normal comment mentioning /sloper spec",
	} {
		if cmds := ParseComments([]models.CommentInfo{{ID: 1, Author: "butter", Body: body}}); len(cmds) != 0 {
			t.Errorf("%q parsed into %+v, want no command", body, cmds)
		}
	}
}

func TestParseCommentsTreatsReviseAsSpecAlias(t *testing.T) {
	for _, body := range []string{
		"/sloper revise",
		"/sloper revise make the plan smaller",
		"/sloper Revise",
	} {
		cmds := ParseComments([]models.CommentInfo{{ID: 7, Author: "butter", Body: body}})
		if len(cmds) != 1 {
			t.Fatalf("%q parsed into %d commands, want 1", body, len(cmds))
		}
		if cmds[0].Type != CmdSpec {
			t.Errorf("%q parsed as %q, want %q", body, cmds[0].Type, CmdSpec)
		}
		if cmds[0].CommentID != 7 {
			t.Errorf("%q lost its comment ID: %+v", body, cmds[0])
		}
	}
}

func TestParseCommentsKeepsAuthorAndCommentID(t *testing.T) {
	cmds := ParseComments([]models.CommentInfo{
		{ID: 42, Author: "butter", Body: "some text\n/sloper approve\nmore text"},
	})
	if len(cmds) != 1 {
		t.Fatalf("parsed into %d commands, want 1", len(cmds))
	}
	if cmds[0].Author != "butter" || cmds[0].CommentID != 42 {
		t.Errorf("command = %+v, want author butter and comment 42", cmds[0])
	}
}
