package html2text

import (
	"context"
	"strings"

	"github.com/dance/plego/core"
	"golang.org/x/net/html"
)

type Filter struct{}

func New() *Filter {
	return &Filter{}
}

func (f *Filter) Name() string { return "Filter::HTML2Text" }

func (f *Filter) Filter(_ context.Context, entry *core.Entry) (*core.Entry, error) {
	if entry == nil {
		return nil, nil
	}

	text := convert(entry.Body)
	result := *entry
	result.Body = text
	return &result, nil
}

func convert(input string) string {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return stripTagsSimple(input)
	}
	var b strings.Builder
	extractText(doc, &b)
	return strings.TrimSpace(b.String())
}

func extractText(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(text)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, b)
	}
	if n.Type == html.ElementNode {
		switch n.Data {
		case "p", "br", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "tr":
			b.WriteByte('\n')
		}
	}
}

func stripTagsSimple(input string) string {
	var b strings.Builder
	inTag := false
	for _, r := range input {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			b.WriteByte(' ')
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

var _ core.Filter = (*Filter)(nil)
