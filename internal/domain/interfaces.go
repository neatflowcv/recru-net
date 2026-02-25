package domain

import (
	"context"
	"io"
)

type Source interface {
	Name() string
	Fetch(ctx context.Context, q Query) ([]JobPosting, error)
}

type JobRepository interface {
	UpsertMany(ctx context.Context, jobs []JobPosting) (UpsertResult, error)
	List(ctx context.Context, f ListFilter) ([]JobPosting, error)
	ExportCSV(ctx context.Context, f ListFilter, w io.Writer) error
}
