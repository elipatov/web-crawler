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

func New() (, error) {
	const expr = "(?s)(https?:\\/\\/[\\w+\\-&@#\\/%?=~_|!:, .;]*[\\w+\\-&@#\\/%=~_|])"

	linkRegexp, err := regexp.Compile("")
	if err != nil {
		panic(fmt.Sprintf("Failed to compile regular expression: %s", expr))
		return nil
	}
}

func (c *Crawler) ParseBody(body []byte) Res {
	matches := c.linkRegexp.FindAll(body, -1)
	res := make([]string, len(matches))

	for i, match := range matches {
		res[i] = string(match)
	}

	return res
}
