package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"n8go-docs/diagnostics"
	"n8go-docs/manifest"
)

type SearchIndexEntry struct {
	Title   string
	Url     string
	Content string
}

func AddToSearchIndex(siteManifest manifest.SiteManifest, entry SearchIndexEntry) {
	indexPath := filepath.Join(siteManifest.OutputPath, "search", "index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		writeNewSearchIndex(indexPath, entry)
	} else {
		appendToSearchIndex(indexPath, entry)
	}
}

func writeNewSearchIndex(indexPath string, entry SearchIndexEntry) {
	if err := os.MkdirAll(filepath.Dir(indexPath), os.ModePerm); err != nil {
		diagnostics.PrintError(err, "failed to create search directory")
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		diagnostics.PrintError(err, "failed to marshal search content")
		return
	}

	if err := os.WriteFile(indexPath, append([]byte("[ "), append(data, []byte(" ]")...)...), os.ModePerm); err != nil {
		diagnostics.PrintError(err, "failed to write search index")
	}
}

func appendToSearchIndex(indexPath string, entry SearchIndexEntry) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		diagnostics.PrintError(err, "failed to read search index file")
		return
	}

	var existing []SearchIndexEntry
	if err := json.Unmarshal(data, &existing); err != nil {
		diagnostics.PrintError(err, "failed to unmarshal search index file")
		return
	}

	existing = append(existing, entry)

	jsonData, err := json.Marshal(existing)
	if err != nil {
		diagnostics.PrintError(err, "failed to marshal search content")
		return
	}

	if err := os.WriteFile(indexPath, jsonData, os.ModePerm); err != nil {
		diagnostics.PrintError(err, "failed to write search index")
	}
}
