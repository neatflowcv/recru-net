package jumpit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/neatflowcv/recru-net/internal/domain"
)

const defaultBaseURL = "https://jumpit-api.saramin.co.kr"

var errUnexpectedStatus = errors.New("unexpected jumpit status")

type Provider struct {
	baseURL    string
	httpClient *http.Client
}

func NewProvider(httpClient *http.Client) *Provider {
	return NewProviderWithBaseURL(defaultBaseURL, httpClient)
}

func NewProviderWithBaseURL(baseURL string, httpClient *http.Client) *Provider {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &Provider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (p *Provider) ListPositions(ctx context.Context) ([]*domain.Position, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/positions", nil)
	if err != nil {
		return nil, fmt.Errorf("build jumpit request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request jumpit positions: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errUnexpectedStatus, resp.StatusCode)
	}

	var payload listPositionsResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("decode jumpit positions: %w", err)
	}

	positions := make([]*domain.Position, 0, len(payload.Result.Positions))

	for _, item := range payload.Result.Positions {
		position, err := item.toDomainPosition()
		if err != nil {
			return nil, err
		}

		positions = append(positions, position)
	}

	return positions, nil
}

type listPositionsResponse struct {
	Result struct {
		Positions []positionDTO `json:"positions"`
	} `json:"result"`
}

type positionDTO struct {
	ID          int64    `json:"id"`
	JobCategory string   `json:"jobCategory"`
	Title       string   `json:"title"`
	CompanyName string   `json:"companyName"`
	TechStacks  []string `json:"techStacks"`
	Locations   []string `json:"locations"`
	Newcomer    bool     `json:"newcomer"`
	MinCareer   int      `json:"minCareer"`
	MaxCareer   int      `json:"maxCareer"`
	ClosedAt    string   `json:"closedAt"`
}

func (p positionDTO) toDomainPosition() (*domain.Position, error) {
	var closesAt *time.Time

	if p.ClosedAt != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05", p.ClosedAt)
		if err != nil {
			return nil, fmt.Errorf("parse jumpit closedAt %q: %w", p.ClosedAt, err)
		}

		closesAt = &parsed
	}

	position := &domain.Position{
		ID:            strconv.FormatInt(p.ID, 10),
		Source:        domain.PositionSourceJumpit,
		ExternalID:    strconv.FormatInt(p.ID, 10),
		Title:         p.Title,
		CompanyName:   p.CompanyName,
		JobCategories: splitCSV(p.JobCategory),
		TechStacks:    append([]string(nil), p.TechStacks...),
		Locations:     append([]string(nil), p.Locations...),
		Career: domain.CareerRange{
			MinYears:   nil,
			MaxYears:   nil,
			EntryLevel: p.Newcomer,
		},
		ClosesAt: closesAt,
	}

	if p.MinCareer > 0 || (!p.Newcomer && p.MinCareer == 0 && p.MaxCareer > 0) {
		position.Career.MinYears = intPtr(p.MinCareer)
	}

	if p.MaxCareer > 0 {
		position.Career.MaxYears = intPtr(p.MaxCareer)
	}

	return position, nil
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")

	items := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		items = append(items, trimmed)
	}

	return items
}

func intPtr(value int) *int {
	return &value
}
