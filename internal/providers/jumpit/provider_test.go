package jumpit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neatflowcv/recru-net/internal/domain"
	"github.com/neatflowcv/recru-net/internal/providers/jumpit"
)

func TestProviderListPositions(t *testing.T) {
	t.Parallel()

	// Arrange
	var requestedPages []int

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		page, ok := validatePositionsRequest(t, responseWriter, request)
		if !ok {
			return
		}

		requestedPages = append(requestedPages, page)

		writePositionsResponse(responseWriter, page)
	}))
	defer server.Close()

	provider := jumpit.NewProviderWithBaseURL(server.URL, server.Client())

	// Act
	positions, err := provider.ListPositions(context.Background())

	// Assert
	require.NoError(t, err)
	require.Equal(t, []int{1, 2}, requestedPages)
	require.Len(t, positions, 2)
	assertPosition(t, positions[0])
	require.Equal(t, "456", positions[1].ExternalID)
	require.Equal(t, "Backend Engineer", positions[1].Title)
}

func validatePositionsRequest(
	t *testing.T,
	responseWriter http.ResponseWriter,
	request *http.Request,
) (int, bool) {
	t.Helper()

	if request.URL.Path != "/api/positions" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		t.Errorf("unexpected path: %s", request.URL.Path)

		return 0, false
	}

	if request.URL.Query().Get("sort") != "reg_dt" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		t.Errorf("unexpected sort query: %s", request.URL.Query().Get("sort"))

		return 0, false
	}

	pageValue := request.URL.Query().Get("page")
	if pageValue == "" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		t.Error("missing page query")

		return 0, false
	}

	page, err := strconv.Atoi(pageValue)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		t.Errorf("invalid page query %q: %v", pageValue, err)

		return 0, false
	}

	return page, true
}

func writePositionsResponse(responseWriter http.ResponseWriter, page int) {
	responseWriter.Header().Set("Content-Type", "application/json")

	switch page {
	case 1:
		_, _ = responseWriter.Write([]byte(`{
			"result": {
				"totalCount": 2,
				"page": 1,
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
	case 2:
		_, _ = responseWriter.Write([]byte(`{
			"result": {
				"totalCount": 2,
				"page": 2,
				"positions": [
					{
						"id": 456,
						"jobCategory": "서버/백엔드 개발자",
						"title": "Backend Engineer",
						"companyName": "Second Labs",
						"techStacks": ["Go"],
						"locations": ["서울 마포구"],
						"newcomer": false,
						"minCareer": 2,
						"maxCareer": 5,
						"closedAt": "2026-03-12T23:59:59"
					}
				]
			}
		}`))
	default:
		responseWriter.WriteHeader(http.StatusBadRequest)
	}
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
