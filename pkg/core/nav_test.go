package core

import (
	"os"
	"path/filepath"
	"testing"

	"n8go-docs/manifest"
)

func TestBuildPageIndexCloneAndMarkActiveNodes(t *testing.T) {
	contexts := []pageContext{
		{Page: pageInfo{FilePath: filepath.Join("docs", "index.md")}, Url: "."},
		{Page: pageInfo{FilePath: filepath.Join("docs", "guide", "page.md")}, Url: "guide/page"},
	}

	index := buildPageIndex(contexts)
	if index["docs/index.md"] != "." {
		t.Fatalf("root page was not indexed: %#v", index)
	}
	if index["docs/guide/page.md"] != "guide/page" {
		t.Fatalf("nested page was not indexed: %#v", index)
	}

	original := []*navNode{{
		Name: "Guide",
		Children: []*navNode{{
			Name: "Page",
			Url:  "guide/page",
		}},
	}}
	cloned := cloneNavTree(original)
	if !markActiveNodes(cloned, "guide/page") {
		t.Fatal("expected active node to be found")
	}
	if !cloned[0].Active || !cloned[0].Children[0].Active {
		t.Fatalf("expected active leaf and ancestor: %#v", cloned)
	}
	if original[0].Active || original[0].Children[0].Active {
		t.Fatalf("clone should not mutate original tree: %#v", original)
	}
	if markActiveNodes(cloned, "missing") {
		t.Fatal("unexpected active match for missing URL")
	}
}

func TestBuildNavFromConfigBuildsNestedContextsAndRootPath(t *testing.T) {
	root := t.TempDir()
	inputDir := filepath.Join(root, "docs")
	writeTestFile(t, filepath.Join(inputDir, "index.md"), "# Home")
	writeTestFile(t, filepath.Join(inputDir, "guide", "page.md"), "# Page")

	var contexts []pageContext
	nodes, err := buildNavFromConfig([]manifest.NavItem{
		{Title: "Home", File: "index.md"},
		{Title: "Guide", Children: []manifest.NavItem{
			{Title: "Page", File: "guide/page.md"},
		}},
	}, manifest.SiteManifest{InputPath: inputDir}, testThemeManifest(), &contexts)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 2 || len(nodes[1].Children) != 1 {
		t.Fatalf("unexpected nav tree: %#v", nodes)
	}
	if len(contexts) != 2 {
		t.Fatalf("expected 2 contexts, got %d", len(contexts))
	}
	if contexts[0].Url != "." || contexts[0].RootPath != "" {
		t.Fatalf("unexpected root context: %#v", contexts[0])
	}
	if contexts[1].Url != "guide/page" || contexts[1].RootPath != "../../" {
		t.Fatalf("unexpected nested context: %#v", contexts[1])
	}
}

func TestIndexAndReadmeHelpers(t *testing.T) {
	root := t.TempDir()
	indexPath := filepath.Join(root, "index.md")
	readmePath := filepath.Join(root, "README.md")

	writeTestFile(t, indexPath, "# Index")
	writeTestFile(t, readmePath, "# Readme")

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if !isIndexFile(indexPath) {
		t.Fatal("index.md should be recognized as index file")
	}
	if !hasIndexFile(entries) {
		t.Fatal("directory should report index.md")
	}
	if isReadmeAsIndex(readmePath, entries) {
		t.Fatal("README.md should not act as index when index.md exists")
	}

	onlyReadmeDir := filepath.Join(root, "only-readme")
	onlyReadmePath := filepath.Join(onlyReadmeDir, "README.md")
	writeTestFile(t, onlyReadmePath, "# Readme")
	onlyReadmeEntries, err := os.ReadDir(onlyReadmeDir)
	if err != nil {
		t.Fatal(err)
	}
	if !isReadmeAsIndex(onlyReadmePath, onlyReadmeEntries) {
		t.Fatal("README.md should act as index when no index.md exists")
	}
}
