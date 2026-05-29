package core

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"n8go-docs/diagnostics"
	"n8go-docs/manifest"
	"n8go-docs/utils"
)


func isIndexFile(filePath string) bool {
	return utils.GetFileName(filePath) == IndexFileName
}

// hasIndexFile reports whether dirPath contains an index.md file.
func hasIndexFile(dirEntries []os.DirEntry) bool {
	for _, e := range dirEntries {
		if !e.IsDir() && utils.GetFileName(e.Name()) == IndexFileName {
			return true
		}
	}
	return false
}

// isReadmeAsIndex reports whether filePath is README.md acting as the index
// for its directory (i.e. no index.md exists alongside it).
func isReadmeAsIndex(filePath string, dirEntries []os.DirEntry) bool {
	return utils.GetFileName(filePath) == ReadmeFileName && !hasIndexFile(dirEntries)
}

func sortTree(node *navNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].Url == "." {
			return true
		}
		return len(node.Children[i].Children) < len(node.Children[j].Children)
	})
	for _, c := range node.Children {
		sortTree(c)
	}
}

func prepareDocumentationTree(dirPath string, rootDirPrefix string, parentNode *navNode, siteManifest manifest.SiteManifest, themeManifest manifest.ThemeManifest, contexts *[]pageContext) error {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, dirent := range dir {
		childPath := filepath.Join(dirPath, dirent.Name())
		relPath := filepath.ToSlash(childPath[len(siteManifest.InputPath)+1:])
		if utils.MatchesExclude(relPath, siteManifest.ExcludeDocs) {
			continue
		}

		newNode := &navNode{
			Name: strings.ReplaceAll(dirent.Name(), "-", " "),
			Url:  "",
		}

		if dirent.IsDir() {
			err := prepareDocumentationTree(childPath, rootDirPrefix+"../", newNode, siteManifest, themeManifest, contexts)
			if err != nil {
				return err
			}
			parentNode.Children = append(parentNode.Children, newNode)
		} else if filepath.Ext(dirent.Name()) == ".md" {
			diagnostics.Debug(func() {
				log.Println("processing:\n", childPath)
			})

			relForLog := childPath[len(siteManifest.InputPath)+1:]
			diagnostics.Info("Parsing '%s'", relForLog)

			treatsAsIndex := isIndexFile(childPath) || isReadmeAsIndex(childPath, dir)

			currentRootDirPrefix := rootDirPrefix
			if !treatsAsIndex {
				currentRootDirPrefix = "../" + currentRootDirPrefix
			}

			context, err := createPageContext(childPath, currentRootDirPrefix, siteManifest, themeManifest)
			if err != nil {
				return err
			}

			newNode.Name = context.Page.Title
			newNode.Url = context.Url
			*contexts = append(*contexts, context)
			parentNode.Children = append(parentNode.Children, newNode)
		}
	}

	sortTree(parentNode)

	return nil
}

func buildPageIndex(contexts []pageContext) PageIndex {
	index := make(PageIndex, len(contexts))
	for _, ctx := range contexts {
		index[filepath.ToSlash(ctx.Page.FilePath)] = ctx.Url
	}
	return index
}

// buildNavFromConfig constructs a nav tree and page contexts from the explicit nav config.
// Each leaf entry is a .md file; sections are recursed.
// cloneNavTree deep-copies a nav tree so each page can independently set Active flags.
func cloneNavTree(nodes []*navNode) []*navNode {
	result := make([]*navNode, len(nodes))
	for i, n := range nodes {
		clone := &navNode{Name: n.Name, Url: n.Url, Active: false}
		if len(n.Children) > 0 {
			clone.Children = cloneNavTree(n.Children)
		}
		result[i] = clone
	}
	return result
}

// markActiveNodes sets Active=true on the node that owns the current page URL,
// and on all ancestor nodes up the tree (so tabs and sections can highlight correctly).
// Returns true if any node in this subtree matched.
func markActiveNodes(nodes []*navNode, currentUrl string) bool {
	for _, n := range nodes {
		if n.Url != "" && n.Url == currentUrl {
			n.Active = true
			return true
		}
		if markActiveNodes(n.Children, currentUrl) {
			n.Active = true
			return true
		}
	}
	return false
}

func buildNavFromConfig(items []manifest.NavItem, siteManifest manifest.SiteManifest, themeManifest manifest.ThemeManifest, contexts *[]pageContext) ([]*navNode, error) {
	var nodes []*navNode
	for _, item := range items {
		node := &navNode{Name: item.Title}
		if item.File != "" {
			mdFile := filepath.Join(siteManifest.InputPath, filepath.FromSlash(item.File))

			relForLog := filepath.ToSlash(item.File)
			diagnostics.Info("Parsing '%s'", relForLog)

			// Compute rootPath: depth of the page's output directory determines how many "../" needed.
			// We need createPageContext which handles this — call with empty rootPath first to get Url,
			// then recompute rootPath from the actual output Url depth.
			ctx, err := createPageContext(mdFile, "", siteManifest, themeManifest)
			if err != nil {
				return nil, err
			}
			// rootPath = one "../" per path segment in the output URL
			depth := 0
			if ctx.Url != "" && ctx.Url != "." {
				depth = len(strings.Split(ctx.Url, "/"))
			}
			rootPath := strings.Repeat("../", depth)
			ctx.RootPath = rootPath

			node.Url = ctx.Url
			*contexts = append(*contexts, ctx)
		} else if len(item.Children) > 0 {
			children, err := buildNavFromConfig(item.Children, siteManifest, themeManifest, contexts)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
