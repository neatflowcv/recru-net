package flow

import (
	"context"
	"errors"
	"fmt"

	"github.com/neatflowcv/recru-net/internal/domain"
)

var (
	errPositionProviderNil   = errors.New("position provider is nil")
	errPositionRepositoryNil = errors.New("position repository is nil")
)

type Service struct {
	positionProvider   domain.PositionProvider
	positionRepository domain.PositionRepository
}

func NewService(
	positionProvider domain.PositionProvider,
	positionRepository domain.PositionRepository,
) (*Service, error) {
	if positionProvider == nil {
		return nil, errPositionProviderNil
	}

	if positionRepository == nil {
		return nil, errPositionRepositoryNil
	}

	return &Service{
		positionProvider:   positionProvider,
		positionRepository: positionRepository,
	}, nil
}

func (s *Service) Sync(ctx context.Context) ([]*domain.Position, error) {
	positions, err := s.positionProvider.ListPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}

	err = s.positionRepository.UpsertPositions(ctx, positions)
	if err != nil {
		return nil, fmt.Errorf("upsert positions: %w", err)
	}

	return positions, nil
}
