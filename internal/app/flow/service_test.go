package flow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/neatflowcv/recru-net/internal/app/flow"
	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/stretchr/testify/require"
)

var errWriteFailed = errors.New("write failed")

func TestServiceSyncRequiresProvider(t *testing.T) {
	t.Parallel()

	// Arrange
	repository := &stubPositionRepository{
		positions: nil,
		err:       nil,
	}

	// Act
	service, err := flow.NewService(nil, repository)

	// Assert
	require.EqualError(t, err, "position provider is nil")
	require.Nil(t, service)
}

func TestServiceSyncRequiresRepository(t *testing.T) {
	t.Parallel()

	// Arrange
	provider := stubPositionProvider{
		positions: nil,
		err:       nil,
	}

	// Act
	service, err := flow.NewService(provider, nil)

	// Assert
	require.EqualError(t, err, "position repository is nil")
	require.Nil(t, service)
}

func TestServiceSyncUpsertsPositions(t *testing.T) {
	t.Parallel()

	// Arrange
	provider := stubPositionProvider{
		positions: []*domain.Position{
			{Title: "Platform Engineer"},
			{Title: "Backend Engineer"},
		},
		err: nil,
	}
	repository := &stubPositionRepository{
		positions: nil,
		err:       nil,
	}
	service := newTestService(t, provider, repository)

	// Act
	positions, err := service.Sync(context.Background())

	// Assert
	require.NoError(t, err)
	require.Len(t, positions, 2)
	require.Equal(t, positions, repository.positions)
}

func TestServiceSyncReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	// Arrange
	provider := stubPositionProvider{
		positions: nil,
		err:       nil,
	}
	repository := stubPositionRepository{
		positions: nil,
		err:       errWriteFailed,
	}
	service := newTestService(t, provider, &repository)

	// Act
	_, err := service.Sync(context.Background())

	// Assert
	require.EqualError(t, err, "upsert positions: write failed")
}

type stubPositionProvider struct {
	positions []*domain.Position
	err       error
}

func newTestService(
	t *testing.T,
	provider domain.PositionProvider,
	repository domain.PositionRepository,
) *flow.Service {
	t.Helper()

	service, err := flow.NewService(provider, repository)
	require.NoError(t, err)

	return service
}

func (s stubPositionProvider) ListPositions(_ context.Context) ([]*domain.Position, error) {
	return s.positions, s.err
}

type stubPositionRepository struct {
	positions []*domain.Position
	err       error
}

func (s *stubPositionRepository) UpsertPositions(_ context.Context, positions []*domain.Position) error {
	s.positions = positions

	return s.err
}
