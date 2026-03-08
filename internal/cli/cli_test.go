//nolint:testpackage
package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/neatflowcv/recru-net/internal/app/flow"
	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestRunSync(t *testing.T) {
	// Arrange
	newSyncService = func() (*flow.Service, string, error) {
		return flow.NewService(stubPositionProvider{
			positions: []*domain.Position{
				{Title: "Platform Engineer"},
				{Title: "Backend Engineer"},
			},
			err: nil,
		}, stubPositionRepository{err: nil}), defaultDatabaseURL, nil
	}

	t.Cleanup(func() {
		newSyncService = defaultNewSyncService
	})

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	// Act
	err := Run([]string{"sync"}, &stdout, &stderr)

	// Assert
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "using DATABASE_URL="+defaultDatabaseURL)
	require.Contains(t, stdout.String(), "synced 2 positions from jumpit")
}

//nolint:paralleltest
func TestRunSyncReturnsServiceBuildError(t *testing.T) {
	newSyncService = func() (*flow.Service, string, error) {
		return nil, "", context.DeadlineExceeded
	}

	t.Cleanup(func() {
		newSyncService = defaultNewSyncService
	})

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	err := Run([]string{"sync"}, &stdout, &stderr)

	require.EqualError(t, err, "run CLI command: build sync service: context deadline exceeded")
}

//nolint:paralleltest
func TestRunList(t *testing.T) {
	// Arrange
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	// Act
	err := Run([]string{"list"}, &stdout, &stderr)
	got := stdout.String()

	// Assert
	require.NoError(t, err)
	require.Contains(t, got, "SOURCE   COMPANY      TITLE")
	require.Contains(t, got, "Backend Engineer")
}

//nolint:paralleltest
func TestRunUnknownCommand(t *testing.T) {
	// Arrange
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	// Act
	err := Run([]string{"unknown"}, &stdout, &stderr)

	// Assert
	require.Error(t, err)
}

type stubPositionProvider struct {
	positions []*domain.Position
	err       error
}

func (s stubPositionProvider) ListPositions(_ context.Context) ([]*domain.Position, error) {
	return s.positions, s.err
}

type stubPositionRepository struct {
	err error
}

func (s stubPositionRepository) UpsertPositions(_ context.Context, _ []*domain.Position) error {
	return s.err
}
