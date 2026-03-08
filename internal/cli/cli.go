package cli

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Sync syncCmd `cmd:"" help:"Run job sync."`
	List listCmd `cmd:"" help:"List stored jobs."`
}

type syncCmd struct{}

type listCmd struct{}

func Run(args []string, stdout, stderr io.Writer) error {
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
		return err
	}

	return ctx.Run()
}

func (c syncCmd) Run(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, "sync stub: not wired yet")

	return err
}

func (c listCmd) Run(stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, "SOURCE   COMPANY      TITLE")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, "saramin  ACME Corp    Backend Engineer")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, "jumpit   Example Labs Platform Engineer")

	return err
}
