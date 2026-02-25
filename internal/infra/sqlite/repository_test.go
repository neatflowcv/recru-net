package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/neatflowcv/recru-net/internal/domain"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	repo, err := NewRepositoryWithDB(db)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestUpsertMany_InsertUpdateUnchanged(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.Close()
	ctx := context.Background()
	now := time.Now().UTC()

	job := domain.JobPosting{
		Source:      "saramin",
		ExternalID:  "123",
		URL:         domain.NormalizeURL("https://example.com/job?b=2&a=1"),
		URLHash:     domain.URLHash("https://example.com/job?b=2&a=1"),
		Title:       "Backend Engineer",
		Company:     "ACME",
		Location:    "Seoul",
		CollectedAt: now,
	}
	job.Fingerprint = domain.ComputeFingerprint(job)

	res, err := repo.UpsertMany(ctx, []domain.JobPosting{job})
	if err != nil {
		t.Fatal(err)
	}
	if res.Inserted != 1 {
		t.Fatalf("expected 1 insert got %+v", res)
	}

	res, err = repo.UpsertMany(ctx, []domain.JobPosting{job})
	if err != nil {
		t.Fatal(err)
	}
	if res.Unchanged != 1 {
		t.Fatalf("expected 1 unchanged got %+v", res)
	}

	job.Title = "Senior Backend Engineer"
	job.Fingerprint = domain.ComputeFingerprint(job)
	res, err = repo.UpsertMany(ctx, []domain.JobPosting{job})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated != 1 {
		t.Fatalf("expected 1 update got %+v", res)
	}
}

func TestExportCSV(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.Close()
	ctx := context.Background()
	now := time.Now().UTC()

	job := domain.JobPosting{
		Source:      "saramin",
		ExternalID:  "999",
		URL:         "https://example.com/job/999",
		URLHash:     domain.URLHash("https://example.com/job/999"),
		Title:       "Go Developer",
		Company:     "ACME",
		Location:    "Seoul",
		CollectedAt: now,
	}
	job.Fingerprint = domain.ComputeFingerprint(job)
	if _, err := repo.UpsertMany(ctx, []domain.JobPosting{job}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := repo.ExportCSV(ctx, domain.ListFilter{Limit: 10}, &buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected csv output")
	}
}
