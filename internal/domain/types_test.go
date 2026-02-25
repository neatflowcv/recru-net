package domain

import "testing"

func TestNormalizeURL_SortsQueryAndRemovesFragment(t *testing.T) {
	raw := "HTTPS://Example.com/path?b=2&a=1#frag"
	got := NormalizeURL(raw)
	want := "https://example.com/path?a=1&b=2"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestURLHash_StableForEquivalentURL(t *testing.T) {
	a := URLHash("https://example.com/job?b=2&a=1")
	b := URLHash("https://example.com/job?a=1&b=2")
	if a != b {
		t.Fatalf("expected stable hash for equivalent urls")
	}
}
