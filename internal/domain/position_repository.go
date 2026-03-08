package domain

import "context"

type PositionRepository interface {
	UpsertPositions(ctx context.Context, positions []*Position) error
}
