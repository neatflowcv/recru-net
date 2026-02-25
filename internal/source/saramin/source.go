package saramin

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/neatflowcv/recru-net/internal/domain"
)

type HTTPGetter interface {
	Get(ctx context.Context, reqURL string) ([]byte, error)
}

type Source struct {
	Client  HTTPGetter
	BaseURL string
}

func New(client HTTPGetter) *Source {
	return &Source{
		Client:  client,
		BaseURL: "https://www.saramin.co.kr",
	}
}

func (s *Source) Name() string {
	return "saramin"
}

func (s *Source) Fetch(ctx context.Context, q domain.Query) ([]domain.JobPosting, error) {
	if s.Client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if q.PageLimit <= 0 {
		q.PageLimit = 1
	}

	out := make([]domain.JobPosting, 0, 128)
	for page := 1; page <= q.PageLimit; page++ {
		pageURL := s.buildSearchURL(q, page)
		body, err := s.Client.Get(ctx, pageURL)
		if err != nil {
			return nil, fmt.Errorf("fetch search page %d: %w", page, err)
		}

		items, err := parseListHTML(body, s.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("parse search page %d: %w", page, err)
		}

		for i := range items {
			if items[i].URL == "" {
				continue
			}
			detailBody, err := s.Client.Get(ctx, items[i].URL)
			if err != nil {
				out = append(out, items[i])
				continue
			}
			detail, err := parseDetailHTML(detailBody)
			if err == nil {
				if detail.Company != "" {
					items[i].Company = detail.Company
				}
				if detail.Location != "" {
					items[i].Location = detail.Location
				}
				if detail.PostedAt != nil {
					items[i].PostedAt = detail.PostedAt
				}
			}
			out = append(out, items[i])
		}
	}

	return out, nil
}

func (s *Source) buildSearchURL(q domain.Query, page int) string {
	values := url.Values{}
	if len(q.Keywords) > 0 {
		values.Set("searchword", strings.Join(q.Keywords, " "))
	}
	if len(q.Locations) > 0 {
		values.Set("loc_mcd", strings.Join(q.Locations, ","))
	}
	values.Set("recruitPage", fmt.Sprintf("%d", page))
	values.Set("recruitSort", "relation")
	return fmt.Sprintf("%s/zf_user/search/recruit?%s", s.BaseURL, values.Encode())
}

var recIDPattern = regexp.MustCompile(`rec_idx=(\d+)`)

func extractExternalID(link string) string {
	if m := recIDPattern.FindStringSubmatch(link); len(m) > 1 {
		return m[1]
	}
	return ""
}

func parseKoreanDate(v string) *time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	layouts := []string{
		"2006-01-02",
		"2006.01.02",
		"2006/01/02",
	}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, v, time.Local)
		if err == nil {
			utc := t.UTC()
			return &utc
		}
	}
	return nil
}
