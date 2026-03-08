package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/recru-net/internal/app/flow"
	"github.com/neatflowcv/recru-net/internal/providers/jumpit"
)

type CLI struct {
	Sync syncCmd `cmd:"" help:"Run job sync."`
	List listCmd `cmd:"" help:"List stored jobs."`
}

type syncCmd struct{}

type listCmd struct{}

func defaultNewSyncService() *flow.Service {
	return flow.NewService(jumpit.NewProvider(nil))
}

var newSyncService = defaultNewSyncService

func Run(args []string, stdout, stderr io.Writer) error {
	root := &CLI{}
	syncService := newSyncService()

	parser, err := kong.New(
		root,
		kong.Name("recru"),
		kong.Description("Recruitment scraper CLI."),
		kong.Writers(stdout, stderr),
		kong.BindTo(stdout, (*io.Writer)(nil)),
		kong.Bind(syncService),
	)
	if err != nil {
		return fmt.Errorf("build CLI: %w", err)
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return fmt.Errorf("parse CLI args: %w", err)
	}

	err = ctx.Run()
	if err != nil {
		return fmt.Errorf("run CLI command: %w", err)
	}

	return nil
}

func (c syncCmd) Run(service *flow.Service, stdout io.Writer) error {
	return runSync(service, stdout)
}

func runSync(service *flow.Service, stdout io.Writer) error {
	positions, err := service.Sync(context.Background())
	if err != nil {
		return fmt.Errorf("sync positions: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "synced %d positions from jumpit\n", len(positions))
	if err != nil {
		return fmt.Errorf("write sync result: %w", err)
	}

	return nil
}

func (c listCmd) Run(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, "SOURCE   COMPANY      TITLE")
	if err != nil {
		return fmt.Errorf("write list header: %w", err)
	}

	_, err = fmt.Fprintln(stdout, "saramin  ACME Corp    Backend Engineer")
	if err != nil {
		return fmt.Errorf("write saramin list row: %w", err)
	}

	_, err = fmt.Fprintln(stdout, "jumpit   Example Labs Platform Engineer")
	if err != nil {
		return fmt.Errorf("write jumpit list row: %w", err)
	}

	return nil
}
