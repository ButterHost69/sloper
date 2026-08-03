-- Comment reply tracking (issue comment replies via GitHub Reply button)
ALTER TABLE issue_comments ADD COLUMN in_reply_to_id INTEGER NOT NULL DEFAULT 0; -- parent comment ID for Reply-button replies
ALTER TABLE issue_comments ADD COLUMN replied_by_bot INTEGER NOT NULL DEFAULT 0;  -- 1 = bot has replied to this comment
