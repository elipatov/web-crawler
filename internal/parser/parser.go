package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/logger"
	"golang.org/x/net/html"
)

type (
	Parser struct {
		linkRegexp *regexp.Regexp
		logger     *logger.Logger
	}

	Result struct {
		Links []string
		Text  string
	}
)

func New(logger *logger.Logger) *Parser {
	const expr = "(?s)(https?:\\/\\/[\\w+\\-&@#\\/%?=~_|!:, .;]*[\\w+\\-&@#\\/%=~_|])"

	linkRegexp, err := regexp.Compile("")
	if err != nil {
		panic(fmt.Sprintf("Failed to compile regular expression: %s", expr))
	}

	return &Parser{
		linkRegexp: linkRegexp,
		logger:     logger,
	}
}

func (p *Parser) ParseBody(body []byte) Result {
	matches := p.linkRegexp.FindAll(body, -1)
	res := Result{
		Links: make([]string, len(matches)),
	}

	for i, match := range matches {
		res.Links[i] = string(match)
	}

	text, err := htmlToText(string(body))
	if err != nil {
		p.logger.WithError(err).Warn("failed to extract text from HTML")
		res.Text = string(body)
	} else {
		res.Text = text
	}

	return res
}

func extractText(n *html.Node, builder *strings.Builder) {
	if n.Type == html.ElementNode && n.Data != "script" && n.Data != "style" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c, builder)
		}

		return
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			builder.WriteString(text)
			builder.WriteString("\r\n")
		}
	}

}

func htmlToText(htmlInput string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", errs.WrapError(err, "failed to parse HTML")
	}

	builder := new(strings.Builder)

	extractText(doc, builder)

	return builder.String(), nil
}
