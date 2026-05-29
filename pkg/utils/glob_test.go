package utils

import "testing"

// TestMatchesExclude_TestConfig verifies the exact patterns used in n8go-docs.test.yaml.
func TestMatchesExclude_TestConfig(t *testing.T) {
	patterns := []string{"drafts/", "**/secret-*.md", "wip.md"}

	excluded := []string{
		"drafts/draft-feature.md",
		"drafts/draft-api.md",
		"user-guide/secret-internal.md",
		"dev-guide/secret-notes.md",
		"wip.md",
	}
	for _, p := range excluded {
		if !MatchesExclude(p, patterns) {
			t.Errorf("expected %q to be excluded by %v", p, patterns)
		}
	}

	included := []string{
		"index.md",
		"getting-started.md",
		"user-guide/installation.md",
		"user-guide/configuration.md",
		"user-guide/writing-your-docs.md",
		"dev-guide/api.md",
		"dev-guide/themes.md",
		"dev-guide/plugins.md",
		"about/contributing.md",
		"about/release-notes.md",
		"about/license.md",
		"css/extra.css",
	}
	for _, p := range included {
		if MatchesExclude(p, patterns) {
			t.Errorf("expected %q NOT to be excluded by %v", p, patterns)
		}
	}
}

func TestMatchesExclude(t *testing.T) {
	cases := []struct {
		path     string
		patterns []string
		want     bool
	}{
		// Basename match (no slash in pattern)
		{"draft.md", []string{"draft.md"}, true},
		{"docs/draft.md", []string{"draft.md"}, true},
		{"docs/other.md", []string{"draft.md"}, false},

		// Wildcard basename
		{"secret-notes.md", []string{"secret-*.md"}, true},
		{"docs/secret-api.md", []string{"secret-*.md"}, true},
		{"docs/public.md", []string{"secret-*.md"}, false},

		// All tmp files
		{"cache/foo.tmp", []string{"*.tmp"}, true},
		{"foo.tmp", []string{"*.tmp"}, true},
		{"foo.md", []string{"*.tmp"}, false},

		// Directory name match (no slash → matches any segment)
		{"drafts/page.md", []string{"drafts"}, true},
		{"drafts/img/photo.png", []string{"drafts"}, true},
		{"other/page.md", []string{"drafts"}, false},

		// Pattern with trailing slash treated same as without
		{"drafts/page.md", []string{"drafts/"}, true},
		{"private/doc.md", []string{"private/"}, true},

		// Pattern with slash — full path match
		{"private/page.md", []string{"private/*.md"}, true},
		{"private/sub/page.md", []string{"private/*.md"}, false},
		{"public/page.md", []string{"private/*.md"}, false},

		// Double-star
		{"a/b/work-in-progress.md", []string{"**/work-in-progress.md"}, true},
		{"work-in-progress.md", []string{"**/work-in-progress.md"}, true},
		{"a/b/c/work-in-progress.md", []string{"**/work-in-progress.md"}, true},
		{"a/b/other.md", []string{"**/work-in-progress.md"}, false},

		// Double-star with prefix
		{"internal/api/secret.md", []string{"internal/**"}, true},
		{"internal/secret.md", []string{"internal/**"}, true},
		{"other/secret.md", []string{"internal/**"}, false},

		// Multiple patterns — any match
		{"draft.md", []string{"private/*.md", "draft.md"}, true},
		{"public/page.md", []string{"private/*.md", "draft.md"}, false},

		// Empty patterns → never excluded
		{"anything.md", []string{}, false},
		{"anything.md", nil, false},
	}

	for _, tc := range cases {
		got := MatchesExclude(tc.path, tc.patterns)
		if got != tc.want {
			t.Errorf("MatchesExclude(%q, %v) = %v, want %v", tc.path, tc.patterns, got, tc.want)
		}
	}
}
