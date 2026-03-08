package flow

import (
	"context"
	"fmt"

	"github.com/neatflowcv/recru-net/internal/domain"
)

type Service struct {
	positionProvider domain.PositionProvider
}

func NewService(positionProvider domain.PositionProvider) *Service {
	return &Service{
		positionProvider: positionProvider,
	}
}

func (s *Service) Sync(ctx context.Context) ([]*domain.Position, error) {
	if s.positionProvider == nil {
		return nil, fmt.Errorf("position provider is nil")
	}

	return s.positionProvider.ListPositions(ctx)
}
