package utils

import (
	"path"
	"strings"
)

// MatchesExclude reports whether relPath (slash-separated, relative to input root)
// matches any of the given glob patterns.
//
// Pattern rules (same as MkDocs exclude_docs):
//   - Standard filepath.Match globs: *, ?, [abc]
//   - ** matches any number of path segments (including zero)
//   - A pattern ending with / matches the directory and everything inside it
//   - A pattern without / is matched against the file name only (basename match)
//   - A pattern containing / (other than trailing) is matched against the full path
func MatchesExclude(relPath string, patterns []string) bool {
	relPath = path.Clean(strings.ReplaceAll(relPath, "\\", "/"))
	for _, pattern := range patterns {
		if matchPattern(relPath, pattern) {
			return true
		}
	}
	return false
}

func matchPattern(relPath, pattern string) bool {
	pattern = strings.TrimSuffix(pattern, "/")
	pattern = path.Clean(strings.ReplaceAll(pattern, "\\", "/"))

	// No slash in pattern → match basename only
	if !strings.Contains(pattern, "/") {
		base := path.Base(relPath)
		if ok, _ := path.Match(pattern, base); ok {
			return true
		}
		// Also treat as directory prefix: if any path segment equals the pattern
		for _, seg := range strings.Split(relPath, "/") {
			if ok, _ := path.Match(pattern, seg); ok {
				return true
			}
		}
		return false
	}

	// Pattern contains ** — expand to match any path prefix
	if strings.Contains(pattern, "**") {
		return matchDoublestar(relPath, pattern)
	}

	// Plain glob with slash — match against full path
	ok, _ := path.Match(pattern, relPath)
	return ok
}

// matchDoublestar handles patterns containing **.
// It splits on ** and tries all combinations of segment counts.
func matchDoublestar(relPath, pattern string) bool {
	parts := strings.SplitN(pattern, "**", 2)
	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	segments := strings.Split(relPath, "/")

	// prefix must match the beginning
	if prefix != "" {
		prefixSegs := strings.Split(prefix, "/")
		if len(segments) < len(prefixSegs) {
			return false
		}
		prefixPath := strings.Join(segments[:len(prefixSegs)], "/")
		if ok, _ := path.Match(prefix, prefixPath); !ok {
			return false
		}
		segments = segments[len(prefixSegs):]
	}

	if suffix == "" {
		return true
	}

	// suffix must match some trailing portion of the remaining segments
	suffixSegs := strings.Split(suffix, "/")
	for i := 0; i <= len(segments)-len(suffixSegs); i++ {
		candidate := strings.Join(segments[i:i+len(suffixSegs)], "/")
		if ok, _ := path.Match(suffix, candidate); ok {
			return true
		}
	}
	return false
}
