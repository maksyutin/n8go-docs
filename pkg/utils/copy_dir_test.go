package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func setupSrcDir(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	files := map[string]string{
		"index.md":        "# Index",
		"style.css":       "body {}",
		"img/photo.png":   "PNG",
		"img/diagram.svg": "SVG",
		"sub/page.md":     "# Sub",
		"sub/script.js":   "var x=1",
		"sub/data.json":   `{"k":"v"}`,
	}
	for rel, content := range files {
		path := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return src
}

func TestCopyDirContents_CopiesMatchingFiles(t *testing.T) {
	src := setupSrcDir(t)
	dst := t.TempDir()

	// Copy everything except .md
	err := CopyDirContents(src, dst, func(ext string) bool {
		return ext != ".md"
	})
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{"style.css", "img/photo.png", "img/diagram.svg", "sub/script.js", "sub/data.json"}
	for _, rel := range expect {
		if _, err := os.Stat(filepath.Join(dst, rel)); err != nil {
			t.Errorf("expected file %s to exist in dst", rel)
		}
	}

	notExpect := []string{"index.md", "sub/page.md"}
	for _, rel := range notExpect {
		if _, err := os.Stat(filepath.Join(dst, rel)); err == nil {
			t.Errorf("expected file %s to NOT exist in dst", rel)
		}
	}
}

func TestCopyDirContents_PreservesContent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	content := "body { color: red; }"
	if err := os.WriteFile(filepath.Join(src, "style.css"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyDirContents(src, dst, func(ext string) bool { return true }); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %q, want %q", string(got), content)
	}
}

func TestCopyDirContentsWithHook_CallsOnCopy(t *testing.T) {
	src := setupSrcDir(t)
	dst := t.TempDir()

	var copied []string
	err := CopyDirContentsWithHook(src, dst, func(ext string) bool {
		return ext == ".css" || ext == ".js"
	}, func(relPath string) bool {
		copied = append(copied, relPath)
		return true
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(copied) != 2 {
		t.Errorf("expected 2 onCopy calls, got %d: %v", len(copied), copied)
	}
	for _, p := range copied {
		if filepath.Ext(p) != ".css" && filepath.Ext(p) != ".js" {
			t.Errorf("unexpected path in onCopy: %s", p)
		}
	}
}

func TestCopyDirContentsWithHook_SkipsWhenOnCopyReturnsFalse(t *testing.T) {
	src := setupSrcDir(t)
	dst := t.TempDir()

	// Skip style.css specifically
	err := CopyDirContentsWithHook(src, dst, func(ext string) bool {
		return ext == ".css" || ext == ".js"
	}, func(relPath string) bool {
		return relPath != "style.css"
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "style.css")); err == nil {
		t.Error("style.css should have been skipped")
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "script.js")); err != nil {
		t.Error("sub/script.js should have been copied")
	}
}

func TestCopyDirContentsWithHook_NilHook(t *testing.T) {
	src := setupSrcDir(t)
	dst := t.TempDir()

	// Must not panic with nil hook
	err := CopyDirContentsWithHook(src, dst, func(ext string) bool { return ext == ".css" }, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCopyDirContents_PredicateFalse_NothingCopied(t *testing.T) {
	src := setupSrcDir(t)
	dst := t.TempDir()

	err := CopyDirContents(src, dst, func(ext string) bool { return false })
	if err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dst)
	if len(entries) != 0 {
		t.Errorf("expected empty dst, got %d entries", len(entries))
	}
}

func TestCopyDirContents_NonExistentSrc(t *testing.T) {
	err := CopyDirContents("/nonexistent/src", t.TempDir(), func(_ string) bool { return true })
	if err == nil {
		t.Error("expected error for non-existent src, got nil")
	}
}

// ── ScanDir ──────────────────────────────────────────────────────────────────

func TestScanDir_FindsMatchingExtension(t *testing.T) {
	dir := t.TempDir()
	files := []string{"a.md", "b.md", "c.html", "d.txt", "sub/e.md"}
	for _, f := range files {
		path := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte(""), 0644)
	}

	got, err := ScanDir(dir, ".md")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 .md files, got %d: %v", len(got), got)
	}
}

func TestScanDir_NoMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "page.html"), []byte(""), 0644)

	got, err := ScanDir(dir, ".md")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 results, got %d", len(got))
	}
}

func TestScanDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := ScanDir(dir, ".md")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 results in empty dir, got %d", len(got))
	}
}
