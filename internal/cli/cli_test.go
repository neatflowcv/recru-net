package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/neatflowcv/recru-net/internal/app/flow"
	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestRunSync(t *testing.T) {
	// Arrange
	newSyncService = func() *flow.Service {
		return flow.NewService(stubPositionProvider{
			positions: []*domain.Position{
				{Title: "Platform Engineer"},
				{Title: "Backend Engineer"},
			},
		})
	}
	t.Cleanup(func() {
		newSyncService = defaultNewSyncService
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Act
	err := Run([]string{"sync"}, &stdout, &stderr)

	// Assert
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "synced 2 positions from jumpit")
}

func TestRunList(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Act
	err := Run([]string{"list"}, &stdout, &stderr)
	got := stdout.String()

	// Assert
	require.NoError(t, err)
	require.Contains(t, got, "SOURCE   COMPANY      TITLE")
	require.Contains(t, got, "Backend Engineer")
}

func TestRunUnknownCommand(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer
	var stderr bytes.Buffer

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
