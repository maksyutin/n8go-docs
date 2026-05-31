package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"n8go-docs/manifest"

	"github.com/microcosm-cc/bluemonday"
)

type SearchIndexEntry struct {
	Title   string
	Url     string
	Content string
}

type SearchIndex struct {
	entries []SearchIndexEntry
}

func NewSearchIndex() *SearchIndex {
	return &SearchIndex{}
}

func (index *SearchIndex) Add(entry SearchIndexEntry) {
	index.entries = append(index.entries, sanitizeSearchIndexEntry(entry))
}

func (index *SearchIndex) Write(siteManifest manifest.SiteManifest) error {
	indexPath := filepath.Join(siteManifest.OutputPath, "search", "index.json")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(index.entries)
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0o644)
}

func sanitizeSearchIndexEntry(entry SearchIndexEntry) SearchIndexEntry {
	policy := bluemonday.StrictPolicy()

	return SearchIndexEntry{
		Title:   normalizeSearchText(policy.Sanitize(entry.Title)),
		Url:     sanitizeSearchURL(entry.Url),
		Content: normalizeSearchText(policy.Sanitize(entry.Content)),
	}
}

func normalizeSearchText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func sanitizeSearchURL(rawURL string) string {
	url := strings.TrimSpace(filepath.ToSlash(rawURL))
	url = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, url)

	if hasSchemePrefix(url) {
		return "./" + url
	}
	return url
}

func hasSchemePrefix(url string) bool {
	colon := strings.IndexRune(url, ':')
	if colon <= 0 {
		return false
	}

	slash := strings.IndexRune(url, '/')
	return slash == -1 || colon < slash
}
