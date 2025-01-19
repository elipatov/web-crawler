package parser

import (
	"fmt"
	"regexp"
)

type (
	Parser struct {
		linkRegexp *regexp.Regexp
	}

	Result struct {
		Links []string
		Text  string
	}
)

func New() *Parser {
	const expr = "(?s)(https?:\\/\\/[\\w+\\-&@#\\/%?=~_|!:, .;]*[\\w+\\-&@#\\/%=~_|])"

	linkRegexp, err := regexp.Compile("")
	if err != nil {
		panic(fmt.Sprintf("Failed to compile regular expression: %s", expr))
	}

	return &Parser{
		linkRegexp: linkRegexp,
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

	return res
}
