package app

import (
	"context"
	"errors"
	"io"

	"github.com/neatflowcv/recru-net/internal/domain"
)

type QueryService struct {
	Repo domain.JobRepository
}

func (s QueryService) List(ctx context.Context, filter domain.ListFilter) ([]domain.JobPosting, error) {
	if s.Repo == nil {
		return nil, errors.New("repo is required")
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return s.Repo.List(ctx, filter)
}

func (s QueryService) ExportCSV(ctx context.Context, filter domain.ListFilter, w io.Writer) error {
	if s.Repo == nil {
		return errors.New("repo is required")
	}
	if filter.Limit <= 0 {
		filter.Limit = 1000
	}
	return s.Repo.ExportCSV(ctx, filter, w)
}
