package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

// ErrRepoNotFound is returned when no repo row matches the request.
var ErrRepoNotFound = errors.New("storage: repo not found")

// ErrRepoLinkExists is returned when the same owner/repo is attached twice.
var ErrRepoLinkExists = errors.New("storage: repo already attached")

// RepoRecord is one row of the repos table: a repo the controller keeps a
// docker instance for.
type RepoRecord struct {
	ID             int64
	Link           string
	Name           string
	CloneURL       string
	ContainerName  string
	ContainerID    string
	Image          string
	HostPort       int
	APIToken       string
	DesiredState   string
	Status         string
	StatusMessage  string
	ContainerState string
	RestartCount   int
	AttachedAt     string
	UpdatedAt      string
}

const repoSelect = `
	SELECT id, link, name, clone_url, container_name, container_id, image, host_port,
	       api_token, desired_state, status, status_message, container_state, restart_count,
	       attached_at, updated_at
	FROM repos`

// InsertRepo stores a newly attached repo and returns its id. Attaching the
// same link twice fails with ErrRepoLinkExists; the conflict is resolved by
// SQLite rather than by a read-then-write in the caller.
func (r *Repositories) InsertRepo(ctx context.Context, rec RepoRecord) (int64, error) {
	if strings.TrimSpace(rec.DesiredState) == "" {
		rec.DesiredState = models.RepoDesiredRunning
	}
	if strings.TrimSpace(rec.Status) == "" {
		rec.Status = models.RepoStatusPending
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO repos (link, name, clone_url, container_name, container_id, image,
		                   host_port, api_token, desired_state, status, status_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(link) DO NOTHING`,
		rec.Link, rec.Name, rec.CloneURL, rec.ContainerName, rec.ContainerID, rec.Image,
		rec.HostPort, rec.APIToken, rec.DesiredState, rec.Status, rec.StatusMessage)
	if err != nil {
		return 0, fmt.Errorf("storage: insert repo %s: %w", rec.Link, err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("storage: insert repo %s: %w", rec.Link, err)
	}
	if inserted == 0 {
		return 0, fmt.Errorf("%w: %s", ErrRepoLinkExists, rec.Link)
	}
	return res.LastInsertId()
}

// ListRepos returns every attached repo in attach order.
func (r *Repositories) ListRepos(ctx context.Context) ([]RepoRecord, error) {
	rows, err := r.db.QueryContext(ctx, repoSelect+` ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("storage: list repos: %w", err)
	}
	defer rows.Close()

	out := make([]RepoRecord, 0)
	for rows.Next() {
		rec, err := scanRepo(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// GetRepo returns one repo by id.
func (r *Repositories) GetRepo(ctx context.Context, id int64) (*RepoRecord, error) {
	row := r.db.QueryRowContext(ctx, repoSelect+` WHERE id = ?`, id)
	rec, err := scanRepo(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: id %d", ErrRepoNotFound, id)
		}
		return nil, err
	}
	return &rec, nil
}

// GetRepoByLink returns one repo by its owner/repo key.
func (r *Repositories) GetRepoByLink(ctx context.Context, link string) (*RepoRecord, error) {
	row := r.db.QueryRowContext(ctx, repoSelect+` WHERE link = ?`, link)
	rec, err := scanRepo(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", ErrRepoNotFound, link)
		}
		return nil, err
	}
	return &rec, nil
}

// UpdateRepoRuntime writes back what the last reconcile observed: container
// identity, status and any error message. Columns are written explicitly rather
// than through a map so a typo cannot silently drop a field.
func (r *Repositories) UpdateRepoRuntime(ctx context.Context, rec RepoRecord) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE repos
		SET container_name = ?, container_id = ?, image = ?, host_port = ?,
		    status = ?, status_message = ?, container_state = ?, restart_count = ?,
		    updated_at = ?
		WHERE id = ?`,
		rec.ContainerName, rec.ContainerID, rec.Image, rec.HostPort,
		rec.Status, rec.StatusMessage, rec.ContainerState, rec.RestartCount,
		now(), rec.ID)
	if err != nil {
		return fmt.Errorf("storage: update repo %d: %w", rec.ID, err)
	}
	return nil
}

// SetRepoDesiredState records whether the operator wants this repo's container
// running. The reconciler is what actually starts or stops it.
func (r *Repositories) SetRepoDesiredState(ctx context.Context, id int64, desired string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE repos SET desired_state = ?, status = ?, status_message = '', updated_at = ? WHERE id = ?`,
		desired, statusForDesired(desired), now(), id)
	if err != nil {
		return fmt.Errorf("storage: set desired state for repo %d: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: id %d", ErrRepoNotFound, id)
	}
	return nil
}

func statusForDesired(desired string) string {
	if desired == models.RepoDesiredStopped {
		return models.RepoStatusStopped
	}
	return models.RepoStatusPending
}

// DeleteRepo removes the row. The container and its volumes are the caller's
// responsibility, because deleting a row must never orphan a running container.
func (r *Repositories) DeleteRepo(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM repos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete repo %d: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: id %d", ErrRepoNotFound, id)
	}
	return nil
}

// rowScanner covers both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanRepo(row rowScanner) (RepoRecord, error) {
	var rec RepoRecord
	if err := row.Scan(&rec.ID, &rec.Link, &rec.Name, &rec.CloneURL, &rec.ContainerName,
		&rec.ContainerID, &rec.Image, &rec.HostPort, &rec.APIToken, &rec.DesiredState,
		&rec.Status, &rec.StatusMessage, &rec.ContainerState, &rec.RestartCount,
		&rec.AttachedAt, &rec.UpdatedAt); err != nil {
		return RepoRecord{}, err
	}
	return rec, nil
}
