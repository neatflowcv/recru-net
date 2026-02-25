package sqlite

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/neatflowcv/recru-net/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(path string) (*Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	repo := &Repository{db: db}
	if err := repo.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return repo, nil
}

func NewRepositoryWithDB(db *sql.DB) (*Repository, error) {
	repo := &Repository{db: db}
	if err := repo.init(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) init(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS job_postings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source TEXT NOT NULL,
  external_id TEXT,
  url TEXT NOT NULL,
  url_hash TEXT NOT NULL,
  title TEXT NOT NULL,
  company TEXT NOT NULL,
  location TEXT,
  posted_at DATETIME,
  fingerprint TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_job_source_external ON job_postings(source, external_id) WHERE external_id IS NOT NULL AND external_id <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_job_source_url_hash ON job_postings(source, url_hash);
CREATE INDEX IF NOT EXISTS idx_job_source_updated ON job_postings(source, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_job_company ON job_postings(company);
`
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (r *Repository) UpsertMany(ctx context.Context, jobs []domain.JobPosting) (domain.UpsertResult, error) {
	result := domain.UpsertResult{}
	if len(jobs) == 0 {
		return result, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, job := range jobs {
		now := job.CollectedAt
		if now.IsZero() {
			now = time.Now().UTC()
		}
		existingID, existingFP, found, err := findExisting(ctx, tx, job)
		if err != nil {
			return result, err
		}

		if !found {
			_, err = tx.ExecContext(ctx, `
INSERT INTO job_postings (source, external_id, url, url_hash, title, company, location, posted_at, fingerprint, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				job.Source,
				nullIfEmpty(job.ExternalID),
				job.URL,
				job.URLHash,
				job.Title,
				job.Company,
				job.Location,
				toNullableTime(job.PostedAt),
				job.Fingerprint,
				now,
				now,
			)
			if err != nil {
				return result, fmt.Errorf("insert job: %w", err)
			}
			result.Inserted++
			continue
		}

		if existingFP == job.Fingerprint {
			result.Unchanged++
			continue
		}

		_, err = tx.ExecContext(ctx, `
UPDATE job_postings
SET external_id = COALESCE(NULLIF(?, ''), external_id),
    url = ?,
    url_hash = ?,
    title = ?,
    company = ?,
    location = ?,
    posted_at = ?,
    fingerprint = ?,
    updated_at = ?
WHERE id = ?`,
			job.ExternalID,
			job.URL,
			job.URLHash,
			job.Title,
			job.Company,
			job.Location,
			toNullableTime(job.PostedAt),
			job.Fingerprint,
			now,
			existingID,
		)
		if err != nil {
			return result, fmt.Errorf("update job: %w", err)
		}
		result.Updated++
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit tx: %w", err)
	}
	return result, nil
}

func findExisting(ctx context.Context, tx *sql.Tx, job domain.JobPosting) (id int64, fingerprint string, found bool, err error) {
	if strings.TrimSpace(job.ExternalID) != "" {
		err = tx.QueryRowContext(ctx, `SELECT id, fingerprint FROM job_postings WHERE source = ? AND external_id = ?`, job.Source, job.ExternalID).
			Scan(&id, &fingerprint)
		if err == nil {
			return id, fingerprint, true, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, "", false, fmt.Errorf("query existing by external_id: %w", err)
		}
	}

	err = tx.QueryRowContext(ctx, `SELECT id, fingerprint FROM job_postings WHERE source = ? AND url_hash = ?`, job.Source, job.URLHash).
		Scan(&id, &fingerprint)
	if err == nil {
		return id, fingerprint, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", false, nil
	}
	return 0, "", false, fmt.Errorf("query existing by url_hash: %w", err)
}

func (r *Repository) List(ctx context.Context, f domain.ListFilter) ([]domain.JobPosting, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	where := make([]string, 0, 3)
	args := make([]any, 0, 4)
	if f.Source != "" {
		where = append(where, "source = ?")
		args = append(args, f.Source)
	}
	if f.Company != "" {
		where = append(where, "LOWER(company) LIKE ?")
		args = append(args, "%"+strings.ToLower(f.Company)+"%")
	}
	if f.Keyword != "" {
		where = append(where, "LOWER(title) LIKE ?")
		args = append(args, "%"+strings.ToLower(f.Keyword)+"%")
	}

	query := `
SELECT id, source, COALESCE(external_id, ''), url, url_hash, title, company, COALESCE(location, ''), posted_at, fingerprint, created_at, updated_at
FROM job_postings`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY updated_at DESC LIMIT ?"
	args = append(args, f.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	result := make([]domain.JobPosting, 0, f.Limit)
	for rows.Next() {
		var job domain.JobPosting
		var postedAt sql.NullTime
		if err := rows.Scan(
			&job.ID,
			&job.Source,
			&job.ExternalID,
			&job.URL,
			&job.URLHash,
			&job.Title,
			&job.Company,
			&job.Location,
			&postedAt,
			&job.Fingerprint,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		if postedAt.Valid {
			pt := postedAt.Time
			job.PostedAt = &pt
		}
		result = append(result, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}
	return result, nil
}

func (r *Repository) ExportCSV(ctx context.Context, f domain.ListFilter, w io.Writer) error {
	jobs, err := r.List(ctx, f)
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"source", "external_id", "title", "company", "location", "url", "posted_at", "updated_at"}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, job := range jobs {
		postedAt := ""
		if job.PostedAt != nil {
			postedAt = job.PostedAt.UTC().Format(time.RFC3339)
		}
		if err := cw.Write([]string{
			job.Source,
			job.ExternalID,
			job.Title,
			job.Company,
			job.Location,
			job.URL,
			postedAt,
			job.UpdatedAt.UTC().Format(time.RFC3339),
		}); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func toNullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC()
}
