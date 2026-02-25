package saramin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseListHTML(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "saramin", "list.html"))
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := parseListHTML(data, "https://www.saramin.co.kr")
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs got %d", len(jobs))
	}
	if jobs[0].ExternalID == "" || jobs[0].Title == "" {
		t.Fatalf("missing parsed fields: %+v", jobs[0])
	}
}

func TestParseDetailHTML(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "saramin", "detail.html"))
	if err != nil {
		t.Fatal(err)
	}
	job, err := parseDetailHTML(data)
	if err != nil {
		t.Fatal(err)
	}
	if job.Company == "" || job.Location == "" || job.PostedAt == nil {
		t.Fatalf("missing parsed detail fields: %+v", job)
	}
}
