package parser

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/elipatov/web-crawler/pkg/logger"
	"golang.org/x/net/html"
)

type (
	Parser struct {
		logger *logger.Logger
	}

	Result struct {
		Links []string
		Text  string
	}
)

func New(logger *logger.Logger) *Parser {
	return &Parser{
		logger: logger,
	}
}

func (p *Parser) ParseBody(body []byte, pageURL string) Result {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		p.logger.WithError(err).Warn("failed to parse HTML")
		return Result{Text: string(body)}
	}

	base, err := url.Parse(pageURL)
	if err != nil {
		p.logger.Warn("invalid page URL", "url", pageURL)
	}

	var links []string
	builder := new(strings.Builder)

	walk(doc, &links, builder, base)

	return Result{
		Links: links,
		Text:  builder.String(),
	}
}

// walk traverses the HTML AST, collecting links and text.
func walk(n *html.Node, links *[]string, builder *strings.Builder, base *url.URL) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			builder.WriteString(text)
			builder.WriteString("\r\n")
		}

		return
	}

	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style":
			// Skip script and style elements.
			return
		case "a":
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				resolved := resolveURL(attr.Val, base)
				if resolved != "" {
					*links = append(*links, resolved)
				}

				break
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, links, builder, base)
	}
}

func resolveURL(href string, base *url.URL) string {
	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	if base != nil {
		parsed = base.ResolveReference(parsed)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	return parsed.String()
}
