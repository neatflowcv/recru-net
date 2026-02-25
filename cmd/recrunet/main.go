package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/neatflowcv/recru-net/internal/app"
	"github.com/neatflowcv/recru-net/internal/config"
	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/neatflowcv/recru-net/internal/infra/httpclient"
	"github.com/neatflowcv/recru-net/internal/infra/sqlite"
	"github.com/neatflowcv/recru-net/internal/source/saramin"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: recrunet <crawl|list|export>")
	}

	switch args[0] {
	case "crawl":
		return runCrawl(ctx, args[1:])
	case "list":
		return runList(ctx, args[1:])
	case "export":
		return runExport(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runCrawl(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("crawl", flag.ContinueOnError)
	configPath := fs.String("config", "./config.yaml", "config yaml path")
	source := fs.String("source", "", "source name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	repo, srcMap, cleanup, err := setupRuntime(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	svc := app.CrawlService{
		Repo:    repo,
		Sources: srcMap,
		Now:     time.Now,
		Logger:  slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	summary, err := svc.Run(ctx, cfg, *source)
	if err != nil {
		return err
	}

	for srcName, result := range summary.SourceResults {
		fmt.Printf("source=%s inserted=%d updated=%d unchanged=%d\n", srcName, result.Inserted, result.Updated, result.Unchanged)
	}
	fmt.Printf("total_fetched=%d\n", summary.TotalFetched)
	return nil
}

func runList(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	configPath := fs.String("config", "./config.yaml", "config yaml path")
	source := fs.String("source", "", "source filter")
	keyword := fs.String("keyword", "", "keyword filter in title")
	company := fs.String("company", "", "company filter")
	limit := fs.Int("limit", 50, "row limit")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	repo, _, cleanup, err := setupRuntime(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	svc := app.QueryService{Repo: repo}
	jobs, err := svc.List(ctx, domain.ListFilter{
		Source:  *source,
		Keyword: *keyword,
		Company: *company,
		Limit:   *limit,
	})
	if err != nil {
		return err
	}

	for _, job := range jobs {
		fmt.Printf("[%s] %s | %s | %s | %s\n", job.Source, job.Title, job.Company, job.Location, job.URL)
	}
	return nil
}

func runExport(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	configPath := fs.String("config", "./config.yaml", "config yaml path")
	source := fs.String("source", "", "source filter")
	keyword := fs.String("keyword", "", "keyword filter in title")
	company := fs.String("company", "", "company filter")
	limit := fs.Int("limit", 1000, "row limit")
	outPath := fs.String("out", "./jobs.csv", "csv output path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	repo, _, cleanup, err := setupRuntime(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	f, err := os.Create(*outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	svc := app.QueryService{Repo: repo}
	if err := svc.ExportCSV(ctx, domain.ListFilter{
		Source:  *source,
		Keyword: *keyword,
		Company: *company,
		Limit:   *limit,
	}, f); err != nil {
		return err
	}

	fmt.Printf("exported %s\n", *outPath)
	return nil
}

func setupRuntime(cfg config.Config) (*sqlite.Repository, map[string]domain.Source, func(), error) {
	repo, err := sqlite.NewRepository(cfg.Storage.SQLitePath)
	if err != nil {
		return nil, nil, nil, err
	}

	httpClient := httpclient.New(
		time.Duration(cfg.HTTP.TimeoutSec)*time.Second,
		cfg.HTTP.RetryCount,
		time.Duration(cfg.HTTP.RetryBackoffMS)*time.Millisecond,
		cfg.HTTP.RateLimitPerSec,
	)

	sources := map[string]domain.Source{
		"saramin": saramin.New(httpClient),
	}

	cleanup := func() {
		httpClient.Close()
		_ = repo.Close()
	}

	return repo, sources, cleanup, nil
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var cfgErr config.ValidationError
	if errors.As(err, &cfgErr) {
		return 2
	}
	return 1
}
