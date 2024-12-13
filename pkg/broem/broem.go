package broem

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Browser struct {
	origin          string
	referer         string
	cookies         map[string]string
	csrfToken       string
	csrfTokenRexp   *regexp.Regexp
	onCookiesUpdate func(map[string]string)
}

func New(origin, csrfTokenRexp string, onCookiesUpdate func(map[string]string)) *Browser {
	return &Browser{
		origin:          origin,
		onCookiesUpdate: onCookiesUpdate,
		csrfTokenRexp:   newRegexp(csrfTokenRexp),
		cookies:         make(map[string]string),
	}
}

func (b *Browser) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if b.referer == "" {
		b.referer = b.origin
	}

	AddHeaders(req, b.origin, b.referer)
	b.addCookies(req)

	return req, nil
}

func (b *Browser) CsrfToken() string {
	return b.csrfToken
}

func (b *Browser) SendRequest(req *http.Request) (*http.Response, []byte, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	b.parseResponse(respBody, res.Header)

	b.referer = req.URL.String()

	return res, respBody, nil
}

func (b *Browser) SetCookies(key, value string) {
	b.cookies[key] = value
}

func (b *Browser) addCookies(req *http.Request) {
	if len(b.cookies) > 0 {
		cookies := make([]string, 0, len(b.cookies))

		for k, v := range b.cookies {
			cookies = append(cookies, fmt.Sprintf("%s=%s", k, v))
		}

		cookieHeader := strings.Join(cookies, "; ")
		req.Header.Add("Cookie", cookieHeader)
	}
}

func (b *Browser) parseResponse(body []byte, headers http.Header) {
	setCookies := headers.Values("Set-Cookie")
	newCookies := make(map[string]string, len(setCookies))

	for _, val := range setCookies {
		parts := strings.Split(val, ";")
		if len(parts) > 0 {
			keyValue := strings.Split(parts[0], "=")
			b.cookies[keyValue[0]] = keyValue[1]
			newCookies[keyValue[0]] = keyValue[1]
		}
	}

	if len(newCookies) > 0 {
		b.onCookiesUpdate(newCookies)
	}

	bodyStr := string(body)

	if b.csrfTokenRexp != nil {
		matchToken := b.csrfTokenRexp.FindStringSubmatch(bodyStr)

		if len(matchToken) > 3 {
			b.csrfToken = matchToken[2]
		}
	}
}

func newRegexp(expr string) *regexp.Regexp {
	if expr == "" {
		return nil
	}

	res, err := regexp.Compile(expr)
	if err != nil {
		fmt.Printf("Failed to compile regular expression: %s\n", expr)
		return nil
	}

	return res
}

func AddHeaders(req *http.Request, origin, referer string) {
	req.Header.Add("Accept", "text/html,spain/xhtml+xml,spain/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,spain/signed-exchange;v=b3;q=0.7")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36'")
	req.Header.Add("sec-ch-ua", "Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"115\", \"Chromium\";v=\"115\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "macOS")
}
