package core

import (
	"bytes"
	stdhtml "html"
	"os"
	"path/filepath"

	"n8go-docs/manifest"
	"n8go-docs/utils"

	hhtml "github.com/alecthomas/chroma/v2/formatters/html"
	headingid "github.com/jkboxomine/goldmark-headingid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type bootstrapCodePreWrapper struct {
	language string
}

func (w bootstrapCodePreWrapper) Start(code bool, styleAttr string) string {
	if !code {
		return `<pre tabindex="0"` + styleAttr + `>`
	}

	return `<pre tabindex="0"` + styleAttr + `><code` + codeLanguageAttrs(w.language) + `>`
}

func (w bootstrapCodePreWrapper) End(code bool) string {
	if code {
		return `</code></pre>`
	}

	return `</pre>`
}

func scanTree(node ast.Node, consumer func(node ast.Node)) {
	consumer(node)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		scanTree(child, consumer)
	}
}

func analyzeDocument(astRoot ast.Node, source []byte, pageInfo *pageInfo) {
	ids := headingid.NewIDs()
	scanTree(astRoot, func(node ast.Node) {
		switch node.Kind() {
		case ast.KindHeading:
			heading := node.(*ast.Heading)

			// Find the page title
			if heading.Level == 1 && pageInfo.Title == "" {
				pageInfo.Title = string(heading.Text(source))
			}

			// Build the table of contents
			if heading.Level < 3 {
				headingName := heading.Text(source)
				pageInfo.Toc = append(pageInfo.Toc, tocEntry{
					Id:    string(ids.Generate(headingName, ast.KindHeading)),
					Name:  string(headingName),
					Level: heading.Level,
				})
			}
		}
	})
}

// htmlElements is the set of standard HTML elements on which Bootstrap classes
// and ARIA attributes are permitted. Using OnElements instead of Globally avoids
// a bluemonday bug where Globally() duplicates already-present class attributes
// when the same attribute appears on the same element twice in the output.
var htmlElements = []string{
	"div", "span", "p", "pre", "code", "blockquote",
	"ul", "ol", "li",
	"table", "thead", "tbody", "tfoot", "tr", "th", "td", "caption",
	"h1", "h2", "h3", "h4", "h5", "h6",
	"a", "img", "figure", "figcaption",
	"section", "article", "aside", "nav", "header", "footer", "main",
	"button", "input", "select", "textarea", "form", "label", "fieldset", "legend",
	"small", "strong", "em", "del", "ins", "mark", "sub", "sup",
	"details", "summary", "dialog",
	"n8go-alert",
}

func markdownHTMLPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowStyling()

	// Bootstrap: class attribute for utility classes, component classes, grid.
	// OnElements avoids the bluemonday Globally() attribute-duplication bug.
	policy.AllowAttrs("class").OnElements(htmlElements...)

	// Bootstrap 5 JS hooks: data-bs-* drive modals, dropdowns, tooltips, carousels.
	// AllowDataAttributes permits all data-* — safe because the content is
	// rendered as static HTML; no server-side evaluation occurs.
	policy.AllowDataAttributes()

	// ARIA attributes required by Bootstrap's accessible components.
	policy.AllowAttrs(
		"aria-label", "aria-labelledby", "aria-describedby",
		"aria-expanded", "aria-controls", "aria-current",
		"aria-haspopup", "aria-modal", "aria-hidden",
		"aria-selected", "aria-disabled",
	).OnElements(htmlElements...)
	policy.AllowAttrs("role").OnElements(htmlElements...)
	policy.AllowAttrs("tabindex").OnElements(htmlElements...)

	policy.AllowAttrs("type", "message").OnElements("n8go-alert")
	policy.AllowAttrs("data-lang").OnElements("code")
	return policy
}

func codeBlockLanguage(context highlighting.CodeBlockContext) string {
	language, ok := context.Language()
	if !ok {
		return ""
	}

	return string(language)
}

func codeLanguageAttrs(language string) string {
	if language == "" {
		return ""
	}

	escapedLanguage := stdhtml.EscapeString(language)
	return ` class="language-` + escapedLanguage + `" data-lang="` + escapedLanguage + `"`
}

func renderHighlightWrapper(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
	if entering {
		_, _ = w.WriteString(`<div class="highlight">`)
		if !context.Highlighted() {
			_, _ = w.WriteString(`<pre><code` + codeLanguageAttrs(codeBlockLanguage(context)) + `>`)
		}
		return
	}

	if !context.Highlighted() {
		_, _ = w.WriteString(`</code></pre>`)
	}
	_, _ = w.WriteString(`</div>`)
}

func renderMarkdownPage(mdFile string, theme manifest.ThemeManifest, siteManifest manifest.SiteManifest) (pageInfo, error) {
	result := pageInfo{
		FilePath: filepath.Clean(mdFile),
		FileName: utils.GetFileName(mdFile),
	}

	source, err := os.ReadFile(mdFile)
	if err != nil {
		return result, err
	}

	// Create Markdown parser
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle(theme.Highlighting.Style),
				highlighting.WithFormatOptions(
					hhtml.WithLineNumbers(theme.Highlighting.LineNumbers),
					hhtml.WithClasses(true),
				),
				highlighting.WithWrapperRenderer(renderHighlightWrapper),
				highlighting.WithCodeBlockOptions(func(context highlighting.CodeBlockContext) []hhtml.Option {
					return []hhtml.Option{
						hhtml.WithPreWrapper(bootstrapCodePreWrapper{language: codeBlockLanguage(context)}),
					}
				}),
			),
			emoji.New(
				emoji.WithRenderingMethod(emoji.Twemoji),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	// Parse Markdown
	reader := text.NewReader(source)
	context := parser.NewContext(parser.WithIDs(headingid.NewIDs()))
	astRoot := md.Parser().Parse(reader, parser.WithContext(context))
	analyzeDocument(astRoot, source, &result)
	if result.Title == "" {
		result.Title = utils.PrettifyTitle(mdFile)
	}

	// Render to HTML
	var buf bytes.Buffer
	err = md.Renderer().Render(&buf, source, astRoot)
	if err == nil {
		result.Body = markdownHTMLPolicy().Sanitize(buf.String())
	}

	return result, err
}
