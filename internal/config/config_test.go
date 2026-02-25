package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ConfigDefaultsAndValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
sources: [saramin]
storage:
  sqlite_path: ./jobs.db
queries:
  - name: backend
    keywords: [go]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.HTTP.TimeoutSec == 0 || cfg.Queries[0].PageLimit == 0 {
		t.Fatalf("defaults not applied")
	}
}

func TestLoad_ValidationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("sources: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(ValidationError); !ok {
		t.Fatalf("expected ValidationError got %T", err)
	}
}
