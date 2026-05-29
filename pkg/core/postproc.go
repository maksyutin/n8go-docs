package core

import (
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func isExternalLink(link string) bool {
	return strings.Contains(link, "://") || strings.HasPrefix(link, "mailto:")
}

func isAnchorOnly(link string) bool {
	return len(link) > 0 && link[0] == '#'
}

// resolveHref resolves a markdown href to a correct relative output URL.
// Links to known pages are resolved via the PageIndex.
// Resource links (no .md extension, not a known page) use RootPath prefix.
func resolveHref(link string, ctx *pageContext) string {
	if isAnchorOnly(link) || isExternalLink(link) {
		return link
	}

	// Split anchor fragment
	anchor := ""
	if idx := strings.Index(link, "#"); idx != -1 {
		anchor = link[idx:]
		link = link[:idx]
	}

	if link == "" {
		return anchor
	}

	// Only attempt page resolution for .md links or links without extension (could be directory index)
	ext := strings.ToLower(filepath.Ext(link))
	isMdLink := ext == ".md" || ext == "" || link == "."

	if isMdLink {
		// Resolve the link to a clean absolute input path
		var absInput string
		if strings.HasPrefix(link, "/") {
			absInput = filepath.Clean(filepath.Join(ctx.Site.InputPath, link))
		} else {
			absInput = filepath.Clean(filepath.Join(ctx.InputDir, link))
		}
		absInput = filepath.ToSlash(absInput)

		if targetUrl, ok := ctx.Index[absInput]; ok {
			return relativeUrl(ctx.Url, targetUrl) + anchor
		}
	}

	// Fallback: resource or unresolvable link — use RootPath to anchor to site root.
	// Strip any leading ../ from the link first (theme templates may not have them,
	// but markdown can reference ../img/foo.png etc).
	stripped := filepath.ToSlash(filepath.Clean(link))
	for strings.HasPrefix(stripped, "../") {
		stripped = stripped[3:]
	}
	// "." means the site root directory.
	if stripped == "." && anchor == "" {
		root := ctx.RootPath
		if root == "" {
			return "./"
		}
		return root
	}
	// Directory-style links (no file extension) get a trailing slash for clean URLs.
	if filepath.Ext(stripped) == "" && anchor == "" {
		return ctx.RootPath + stripped + "/"
	}
	return ctx.RootPath + stripped + anchor
}

// relativeUrl returns a relative URL from currentUrl (output dir of current page)
// to targetUrl (output dir of target page). Both are directory paths like "guide/page" or ".".
func relativeUrl(currentUrl, targetUrl string) string {
	cur := currentUrl
	if cur == "." {
		cur = ""
	}
	tgt := targetUrl
	if tgt == "." {
		tgt = ""
	}

	rel, err := filepath.Rel(cur, tgt)
	if err != nil {
		if tgt == "" {
			return "./"
		}
		return tgt + "/"
	}
	rel = filepath.ToSlash(rel)
	// filepath.Rel("a/b", "") returns "../../." — strip the trailing "/.".
	rel = strings.TrimSuffix(rel, "/.")

	if rel == "" || rel == "." {
		return "./"
	}
	return rel + "/"
}

func processHtmlNode(node *html.Node, ctx *pageContext) {
	for idx, attr := range node.Attr {
		switch attr.Key {
		case "href":
			node.Attr[idx].Val = resolveHref(attr.Val, ctx)
		case "src":
			if !isExternalLink(attr.Val) {
				stripped := filepath.ToSlash(filepath.Clean(attr.Val))
				for strings.HasPrefix(stripped, "../") {
					stripped = stripped[3:]
				}
				node.Attr[idx].Val = ctx.RootPath + stripped
			}
		}
	}

	for n := node.FirstChild; n != nil; n = n.NextSibling {
		processHtmlNode(n, ctx)
	}
}

func processHtml(input io.Reader, output io.Writer, ctx *pageContext) error {
	htmlRootNode, err := html.Parse(input)
	if err != nil {
		return err
	}

	processHtmlNode(htmlRootNode, ctx)

	return html.Render(output, htmlRootNode)
}
