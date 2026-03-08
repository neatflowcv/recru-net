package domain

import "context"

type PositionProvider interface {
	ListPositions(ctx context.Context) ([]*Position, error)
}
