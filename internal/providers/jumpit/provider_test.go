package jumpit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/neatflowcv/recru-net/internal/providers/jumpit"
	"github.com/stretchr/testify/require"
)

func TestProviderListPositions(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/positions", r.URL.Path)

		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`{
			"result": {
				"positions": [
					{
						"id": 123,
						"jobCategory": "서버/백엔드 개발자, devops/시스템 엔지니어",
						"title": "Platform Engineer",
						"companyName": "Example Labs",
						"techStacks": ["Go", "PostgreSQL"],
						"locations": ["서울 강남구"],
						"newcomer": true,
						"minCareer": 0,
						"maxCareer": 3,
						"closedAt": "2026-03-11T23:59:59"
					}
				]
			}
		}`))
	}))
	defer server.Close()

	provider := jumpit.NewProviderWithBaseURL(server.URL, server.Client())

	// Act
	positions, err := provider.ListPositions(context.Background())

	// Assert
	require.NoError(t, err)
	require.Len(t, positions, 1)
	assertPosition(t, positions[0])
}

func assertPosition(t *testing.T, got *domain.Position) {
	t.Helper()

	require.Equal(t, "123", got.ExternalID)
	require.Equal(t, domain.PositionSourceJumpit, got.Source)
	require.Len(t, got.JobCategories, 2)
	require.True(t, got.Career.EntryLevel)
	require.Nil(t, got.Career.MinYears)
	require.NotNil(t, got.Career.MaxYears)
	require.Equal(t, 3, *got.Career.MaxYears)

	wantClose, _ := time.Parse("2006-01-02T15:04:05", "2026-03-11T23:59:59")
	require.NotNil(t, got.ClosesAt)
	require.True(t, got.ClosesAt.Equal(wantClose))
}
