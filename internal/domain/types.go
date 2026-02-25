package domain

import (
	"net/url"
	"sort"
	"strings"
	"time"
)

type Query struct {
	Name             string
	Keywords         []string
	Locations        []string
	ExperienceLevels []string
	PageLimit        int
}

type JobPosting struct {
	ID          int64
	Source      string
	ExternalID  string
	URL         string
	URLHash     string
	Title       string
	Company     string
	Location    string
	PostedAt    *time.Time
	CollectedAt time.Time
	Fingerprint string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListFilter struct {
	Source  string
	Keyword string
	Company string
	Limit   int
}

type UpsertResult struct {
	Inserted  int
	Updated   int
	Unchanged int
}

func NormalizeURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}

	u.Fragment = ""
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	q := u.Query()
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	normalizedQ := url.Values{}
	for _, k := range keys {
		vals := q[k]
		sort.Strings(vals)
		for _, v := range vals {
			normalizedQ.Add(k, v)
		}
	}
	u.RawQuery = normalizedQ.Encode()

	return u.String()
}
