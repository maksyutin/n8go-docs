package core

import "n8go-docs/manifest"

type navNode struct {
	Name     string
	Url      string
	Children []*navNode
	Active   bool
}

type tocEntry struct {
	Id    string
	Name  string
	Level int
}

type generatorInfo struct {
	Name    string
	Version string
}

type pageInfo struct {
	FilePath string
	FileName string
	Title    string
	Body     string
	Toc      []tocEntry
}

// PageIndex maps clean input file path → output URL for all pages in the site.
type PageIndex map[string]string

type pageContext struct {
	Page      pageInfo
	Generator generatorInfo
	Now       string
	Site      manifest.SiteManifest
	Nav       []*navNode
	RootPath  string
	Url       string    // output directory path relative to site root
	InputDir  string    // directory of the source .md file (for resolving relative links)
	Index     PageIndex // site-wide mapping inputFile → outputUrl
}
