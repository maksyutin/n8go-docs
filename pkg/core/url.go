package core

import (
	"path/filepath"
	"strings"
)

func resolvePublicURL(siteURL string, rawPath string) string {
	if siteURL == "" || isAnchorOnly(rawPath) || isExternalLink(rawPath) {
		return rawPath
	}

	path, anchor := splitAnchor(rawPath)
	hasTrailingSlash := strings.HasSuffix(path, "/") && path != "/"
	path = filepath.ToSlash(filepath.Clean(path))

	if path == "." || path == "/" {
		return siteURL + "/" + anchor
	}

	for strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}
	path = strings.TrimPrefix(path, "/")
	if (hasTrailingSlash || filepath.Ext(path) == "") && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return siteURL + "/" + path + anchor
}

func isPublicSiteURL(link string, siteURL string) bool {
	return siteURL != "" && (link == siteURL || strings.HasPrefix(link, siteURL+"/"))
}

func splitAnchor(link string) (string, string) {
	if idx := strings.Index(link, "#"); idx != -1 {
		return link[:idx], link[idx:]
	}
	return link, ""
}
