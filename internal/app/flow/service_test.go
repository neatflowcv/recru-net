//nolint:testpackage
package flow

import (
	"context"
	"errors"
	"testing"

	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/stretchr/testify/require"
)

var errWriteFailed = errors.New("write failed")

func TestServiceSyncRequiresProvider(t *testing.T) {
	t.Parallel()

	service := NewService(nil, &stubPositionRepository{
		positions: nil,
		err:       nil,
	})

	_, err := service.Sync(context.Background())

	require.EqualError(t, err, "position provider is nil")
}

func TestServiceSyncRequiresRepository(t *testing.T) {
	t.Parallel()

	service := NewService(stubPositionProvider{
		positions: nil,
		err:       nil,
	}, nil)

	_, err := service.Sync(context.Background())

	require.EqualError(t, err, "position repository is nil")
}

func TestServiceSyncUpsertsPositions(t *testing.T) {
	t.Parallel()

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
	service := NewService(provider, repository)

	positions, err := service.Sync(context.Background())

	require.NoError(t, err)
	require.Len(t, positions, 2)
	require.Equal(t, positions, repository.positions)
}

func TestServiceSyncReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repository := stubPositionRepository{
		positions: nil,
		err:       errWriteFailed,
	}
	service := NewService(stubPositionProvider{
		positions: nil,
		err:       nil,
	}, &repository)

	_, err := service.Sync(context.Background())

	require.EqualError(t, err, "upsert positions: write failed")
}

type stubPositionProvider struct {
	positions []*domain.Position
	err       error
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
