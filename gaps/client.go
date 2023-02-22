package gaps

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"lutonite.dev/gaps-cli/_internal/version"
	"lutonite.dev/gaps-cli/util"
	"net/http"
	"net/url"
	"strings"
)

type ClientConfiguration struct {
	client  *http.Client
	baseUrl string
}

type TokenClientConfiguration struct {
	ClientConfiguration
	token     string
	studentId uint
}

func (c *ClientConfiguration) Init(baseUrl string) {
	c.baseUrl = baseUrl
	c.client = &http.Client{
		Jar: nil,
	}
}

func (c *ClientConfiguration) SetToken(token string) *TokenClientConfiguration {
	return &TokenClientConfiguration{
		ClientConfiguration: *c,
		token:               token,
	}
}

func (tc *TokenClientConfiguration) InitToken(baseUrl string, token string, studentId uint) {
	if token == "" {
		log.Panic("you must be logged in to use this command, please run 'gaps-cli login'")
	}

	tc.Init(baseUrl)
	tc.token = token
	tc.studentId = studentId
}

func (tc *TokenClientConfiguration) buildRequest(method string, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, tc.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	tc.addAuthCookie(req)
	tc.setUserAgent(req)
	return req, nil
}

func (tc *TokenClientConfiguration) doForm(req *http.Request, data url.Values) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	return tc.client.Do(req)
}

func (tc *TokenClientConfiguration) parseUrl() *url.URL {
	baseUrl, err := url.Parse(tc.baseUrl)
	util.CheckErr(err)
	return baseUrl
}

func (tc *TokenClientConfiguration) addAuthCookie(req *http.Request) {
	req.AddCookie(&http.Cookie{
		Name:     "GAPSSESSID",
		Value:    tc.token,
		Domain:   tc.parseUrl().Host,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}

func (tc *TokenClientConfiguration) setUserAgent(req *http.Request) {
	buildInfo := version.Get()
	req.Header.Set(
		"User-Agent",
		fmt.Sprintf("Mozilla/5.0 (%s) gaps-cli/%s", buildInfo.Arch, buildInfo.Version),
	)
}
