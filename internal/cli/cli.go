package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/recru-net/internal/app/flow"
	"github.com/neatflowcv/recru-net/internal/providers/jumpit"
	pgstore "github.com/neatflowcv/recru-net/internal/repositories/postgres"
)

type CLI struct {
	Sync syncCmd `cmd:"" help:"Run job sync."`
	List listCmd `cmd:"" help:"List stored jobs."`
}

type syncCmd struct{}

type listCmd struct{}

const databaseURLEnv = "DATABASE_URL"

//nolint:gosec // Local development default DSN.
const defaultDatabaseURL = "postgres://recur:recur@localhost:5432/recur?sslmode=disable"

func defaultNewSyncService() (*flow.Service, string, error) {
	databaseURL := os.Getenv(databaseURLEnv)
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	positionRepository, err := pgstore.NewPositionRepository(databaseURL)
	if err != nil {
		return nil, "", fmt.Errorf("create position repository: %w", err)
	}

	service, err := flow.NewService(jumpit.NewProvider(nil), positionRepository)
	if err != nil {
		return nil, "", fmt.Errorf("create flow service: %w", err)
	}

	return service, databaseURL, nil
}

//nolint:gochecknoglobals
var newSyncService = defaultNewSyncService

func Run(args []string, stdout, stderr io.Writer) error {
	//nolint:exhaustruct
	root := &CLI{}

	parser, err := kong.New(
		root,
		kong.Name("recru"),
		kong.Description("Recruitment scraper CLI."),
		kong.Writers(stdout, stderr),
		kong.BindTo(stdout, (*io.Writer)(nil)),
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

func (c syncCmd) Run(stdout io.Writer) error {
	service, databaseURL, err := newSyncService()
	if err != nil {
		return fmt.Errorf("build sync service: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "using DATABASE_URL=%s\n", databaseURL)
	if err != nil {
		return fmt.Errorf("write database url: %w", err)
	}

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
