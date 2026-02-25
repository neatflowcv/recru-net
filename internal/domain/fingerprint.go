package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func ComputeFingerprint(job JobPosting) string {
	postedAt := ""
	if job.PostedAt != nil {
		postedAt = job.PostedAt.UTC().Format(time.RFC3339)
	}

	base := strings.Join([]string{
		strings.TrimSpace(strings.ToLower(job.Title)),
		strings.TrimSpace(strings.ToLower(job.Company)),
		strings.TrimSpace(strings.ToLower(job.Location)),
		strings.TrimSpace(NormalizeURL(job.URL)),
		postedAt,
	}, "|")

	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])
}

func URLHash(rawURL string) string {
	normalized := NormalizeURL(rawURL)
	sum := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", sum)
}
