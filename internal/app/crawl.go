package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/neatflowcv/recru-net/internal/config"
	"github.com/neatflowcv/recru-net/internal/domain"
)

type CrawlService struct {
	Repo    domain.JobRepository
	Sources map[string]domain.Source
	Now     func() time.Time
	Logger  *slog.Logger
}

type CrawlSummary struct {
	SourceResults map[string]domain.UpsertResult
	TotalFetched  int
}

func (s CrawlService) Run(ctx context.Context, cfg config.Config, onlySource string) (CrawlSummary, error) {
	if s.Repo == nil {
		return CrawlSummary{}, errors.New("repo is required")
	}
	if len(s.Sources) == 0 {
		return CrawlSummary{}, errors.New("sources are required")
	}
	if s.Now == nil {
		s.Now = time.Now
	}

	summary := CrawlSummary{SourceResults: make(map[string]domain.UpsertResult)}
	queries := cfg.ToDomainQueries()
	sources := make([]string, 0, len(cfg.Sources))
	for _, src := range cfg.Sources {
		if onlySource != "" && src != onlySource {
			continue
		}
		sources = append(sources, src)
	}
	sort.Strings(sources)
	if len(sources) == 0 {
		return summary, fmt.Errorf("no source selected")
	}

	for _, srcName := range sources {
		src, ok := s.Sources[srcName]
		if !ok {
			return summary, fmt.Errorf("source not registered: %s", srcName)
		}

		all := make([]domain.JobPosting, 0, 256)
		for _, q := range queries {
			jobs, err := src.Fetch(ctx, q)
			if err != nil {
				return summary, fmt.Errorf("fetch source=%s query=%s: %w", srcName, q.Name, err)
			}
			for i := range jobs {
				jobs[i].Source = srcName
				jobs[i].CollectedAt = s.Now().UTC()
				jobs[i].URL = domain.NormalizeURL(jobs[i].URL)
				jobs[i].URLHash = domain.URLHash(jobs[i].URL)
				jobs[i].Fingerprint = domain.ComputeFingerprint(jobs[i])
			}
			all = append(all, jobs...)
			summary.TotalFetched += len(jobs)
		}

		result, err := s.Repo.UpsertMany(ctx, all)
		if err != nil {
			return summary, fmt.Errorf("upsert source=%s: %w", srcName, err)
		}
		summary.SourceResults[srcName] = result
		if s.Logger != nil {
			s.Logger.Info("crawl completed", "source", srcName, "fetched", len(all), "inserted", result.Inserted, "updated", result.Updated, "unchanged", result.Unchanged)
		}
	}

	return summary, nil
}
