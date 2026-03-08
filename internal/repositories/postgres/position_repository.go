package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/neatflowcv/recru-net/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const positionBatchSize = 100
const minimumPositionsForDeduplication = 2

var errPositionNil = errors.New("position is nil")

type PositionRepository struct {
	db *gorm.DB
}

func NewPositionRepository(dsn string) (*PositionRepository, error) {
	//nolint:exhaustruct
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	//nolint:exhaustruct
	err = gormDB.AutoMigrate(&positionRecord{})
	if err != nil {
		return nil, fmt.Errorf("auto-migrate positions: %w", err)
	}

	return &PositionRepository{db: gormDB}, nil
}

func (r *PositionRepository) UpsertPositions(ctx context.Context, positions []*domain.Position) error {
	if len(positions) == 0 {
		return nil
	}

	records := make([]positionRecord, 0, len(positions))
	dedupedPositions := dedupePositionsBySourceExternalID(positions)

	for index, position := range dedupedPositions {
		if position == nil {
			return fmt.Errorf("%w: index %d", errPositionNil, index)
		}

		records = append(records, newPositionRecord(position))
	}

	err := gorm.G[positionRecord](
		r.db,
		//nolint:exhaustruct
		clause.OnConflict{
			Columns: []clause.Column{
				//nolint:exhaustruct
				{Name: "source"},
				//nolint:exhaustruct
				{Name: "external_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"id",
				"title",
				"company_name",
				"job_categories",
				"tech_stacks",
				"locations",
				"career_min_years",
				"career_max_years",
				"career_entry_level",
				"closes_at",
				"updated_at",
			}),
		},
	).CreateInBatches(ctx, &records, positionBatchSize)
	if err != nil {
		return fmt.Errorf("upsert positions: %w", err)
	}

	return nil
}

func dedupePositionsBySourceExternalID(positions []*domain.Position) []*domain.Position {
	if len(positions) < minimumPositionsForDeduplication {
		return positions
	}

	lastIndexByKey := make(map[string]int, len(positions))

	for index, position := range positions {
		if position == nil {
			continue
		}

		lastIndexByKey[positionKey(position)] = index
	}

	deduped := make([]*domain.Position, 0, len(lastIndexByKey))

	for index, position := range positions {
		if position == nil {
			deduped = append(deduped, nil)

			continue
		}

		if lastIndexByKey[positionKey(position)] != index {
			continue
		}

		deduped = append(deduped, position)
	}

	return deduped
}

type positionRecord struct {
	ID               string         `gorm:"primaryKey"`
	Source           string         `gorm:"not null;uniqueIndex:idx_positions_source_external_id"`
	ExternalID       string         `gorm:"not null;uniqueIndex:idx_positions_source_external_id"`
	Title            string         `gorm:"not null"`
	CompanyName      string         `gorm:"not null"`
	JobCategories    pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	TechStacks       pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	Locations        pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	CareerMinYears   *int
	CareerMaxYears   *int
	CareerEntryLevel bool `gorm:"not null;default:false"`
	ClosesAt         *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (positionRecord) TableName() string {
	return "positions"
}

func newPositionRecord(position *domain.Position) positionRecord {
	return positionRecord{
		ID:               buildPositionID(position),
		Source:           string(position.Source),
		ExternalID:       position.ExternalID,
		Title:            position.Title,
		CompanyName:      position.CompanyName,
		JobCategories:    toStringArray(position.JobCategories),
		TechStacks:       toStringArray(position.TechStacks),
		Locations:        toStringArray(position.Locations),
		CareerMinYears:   position.Career.MinYears,
		CareerMaxYears:   position.Career.MaxYears,
		CareerEntryLevel: position.Career.EntryLevel,
		ClosesAt:         position.ClosesAt,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
	}
}

func buildPositionID(position *domain.Position) string {
	if position.ID != "" {
		return position.ID
	}

	return fmt.Sprintf("%s:%s", position.Source, position.ExternalID)
}

func positionKey(position *domain.Position) string {
	return fmt.Sprintf("%s:%s", position.Source, position.ExternalID)
}

func toStringArray(values []string) pq.StringArray {
	if len(values) == 0 {
		return pq.StringArray{}
	}

	items := make(pq.StringArray, len(values))
	copy(items, values)

	return items
}
