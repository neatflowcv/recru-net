//nolint:testpackage
package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testDatabaseURLEnv = "DATABASE_URL"

//nolint:paralleltest
func TestPositionRepositoryUpsertPositions(t *testing.T) {
	repository := newTestRepository(t)

	ctx := context.Background()

	first := &domain.Position{
		ID:            "jumpit-1",
		Source:        domain.PositionSourceJumpit,
		ExternalID:    "1",
		Title:         "Platform Engineer",
		CompanyName:   "Example Labs",
		JobCategories: []string{"Backend", "Platform"},
		TechStacks:    []string{"Go", "PostgreSQL"},
		Locations:     []string{"Seoul"},
		Career: domain.CareerRange{
			MinYears:   intPtr(2),
			MaxYears:   intPtr(5),
			EntryLevel: false,
		},
		ClosesAt: nil,
	}

	second := &domain.Position{
		ID:            "jumpit-1",
		Source:        domain.PositionSourceJumpit,
		ExternalID:    "1",
		Title:         "Senior Platform Engineer",
		CompanyName:   "Example Labs",
		JobCategories: []string{"Platform"},
		TechStacks:    []string{"Go", "PostgreSQL", "Kubernetes"},
		Locations:     []string{"Seoul", "Remote"},
		Career: domain.CareerRange{
			MinYears:   intPtr(4),
			MaxYears:   intPtr(8),
			EntryLevel: false,
		},
		ClosesAt: nil,
	}

	require.NoError(t, repository.UpsertPositions(ctx, []*domain.Position{first}))
	require.NoError(t, repository.UpsertPositions(ctx, []*domain.Position{second}))

	records, err := gorm.G[positionRecord](repository.db).
		Where("source = ? AND external_id = ?", string(domain.PositionSourceJumpit), "1").
		Find(ctx)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "Senior Platform Engineer", records[0].Title)
	require.Equal(t, []string{"Go", "PostgreSQL", "Kubernetes"}, []string(records[0].TechStacks))
	require.Equal(t, []string{"Seoul", "Remote"}, []string(records[0].Locations))
	require.Equal(t, 4, derefInt(records[0].CareerMinYears))
	require.Equal(t, 8, derefInt(records[0].CareerMaxYears))
}

//nolint:paralleltest
func TestPositionRepositoryUpsertPositionsDedupesConflictingEntriesInSingleBatch(t *testing.T) {
	repository := newTestRepository(t)

	ctx := context.Background()

	first := &domain.Position{
		ID:            "jumpit-1",
		Source:        domain.PositionSourceJumpit,
		ExternalID:    "1",
		Title:         "Platform Engineer",
		CompanyName:   "Example Labs",
		JobCategories: []string{"Backend"},
		TechStacks:    []string{"Go"},
		Locations:     []string{"Seoul"},
		Career: domain.CareerRange{
			MinYears:   intPtr(1),
			MaxYears:   intPtr(3),
			EntryLevel: false,
		},
		ClosesAt: nil,
	}

	second := &domain.Position{
		ID:            "jumpit-1",
		Source:        domain.PositionSourceJumpit,
		ExternalID:    "1",
		Title:         "Senior Platform Engineer",
		CompanyName:   "Example Labs",
		JobCategories: []string{"Platform"},
		TechStacks:    []string{"Go", "PostgreSQL"},
		Locations:     []string{"Seoul", "Remote"},
		Career: domain.CareerRange{
			MinYears:   intPtr(4),
			MaxYears:   intPtr(8),
			EntryLevel: false,
		},
		ClosesAt: nil,
	}

	require.NoError(t, repository.UpsertPositions(ctx, []*domain.Position{first, second}))

	records, err := gorm.G[positionRecord](repository.db).
		Where("source = ? AND external_id = ?", string(domain.PositionSourceJumpit), "1").
		Find(ctx)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "Senior Platform Engineer", records[0].Title)
	require.Equal(t, []string{"Go", "PostgreSQL"}, []string(records[0].TechStacks))
}

//nolint:paralleltest
func TestPositionRepositoryUpsertPositionsRejectsNilEntry(t *testing.T) {
	repository := newTestRepository(t)

	err := repository.UpsertPositions(context.Background(), []*domain.Position{nil})

	require.EqualError(t, err, "position is nil: index 0")
}

func newTestRepository(t *testing.T) *PositionRepository {
	t.Helper()

	repository, err := NewPositionRepository(testDatabaseURL())
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}

	err = repository.db.Exec("TRUNCATE TABLE positions").Error
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupErr := repository.db.Exec("TRUNCATE TABLE positions").Error
		require.NoError(t, cleanupErr)
	})

	return repository
}

func testDatabaseURL() string {
	if value := os.Getenv(testDatabaseURLEnv); value != "" {
		return value
	}

	return fmt.Sprintf(
		"postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		"recur",
		"recur",
		5432,
		"recur",
	)
}

func intPtr(value int) *int {
	return &value
}

func derefInt(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}

func TestPositionRepositoryNewPositionRepositoryReturnsConnectionError(t *testing.T) {
	t.Parallel()

	_, err := NewPositionRepository("postgres://invalid:invalid@127.0.0.1:1/recur?sslmode=disable")

	require.Error(t, err)
	require.NotErrorIs(t, err, gorm.ErrRecordNotFound)
}
