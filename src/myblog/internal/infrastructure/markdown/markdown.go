package markdown

import (
	"bytes"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type Converter struct {
	md        goldmark.Markdown
	sanitizer *bluemonday.Policy
}

func NewConverter() *Converter {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("id", "class").OnElements("h1", "h2", "h3", "h4", "h5", "h6")

	// Создаем кастомный рендерер
	r := renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(&HeadingRenderer{}, 100),
			util.Prioritized(html.NewRenderer(), 200),
		),
	)

	return &Converter{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Typographer,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
				parser.WithAttribute(),
			),
			goldmark.WithRenderer(r), // Используем WithRenderer вместо цепочки вызовов
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
				html.WithUnsafe(),
			),
		),
		sanitizer: policy,
	}
}

type HeadingRenderer struct{}

func (r *HeadingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

func (r *HeadingRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		level := n.Level
		title := string(n.Text(source))
		customID := strings.ToLower(strings.ReplaceAll(title, " ", "-"))

		_, _ = w.WriteString("<h")
		_ = w.WriteByte("123456"[level]) // Упрощенная запись уровня
		_, _ = w.WriteString(` id="` + customID + `"`)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

func (c *Converter) ToHTML(mdContent string) (string, error) {
	var buf bytes.Buffer
	if err := c.md.Convert([]byte(mdContent), &buf); err != nil {
		return "", err
	}
	return c.sanitizer.Sanitize(buf.String()), nil
}
