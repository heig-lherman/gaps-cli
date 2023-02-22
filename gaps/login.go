package gaps

import (
	"errors"
	"lutonite.dev/gaps-cli/parser"
	"net/url"
)

type LoginAction struct {
	cfg *ClientConfiguration

	username string
	password string
}

func NewLoginAction(config *ClientConfiguration, username string, password string) *LoginAction {
	return &LoginAction{
		cfg:      config,
		username: username,
		password: password,
	}
}

func (a *LoginAction) FetchToken() (string, error) {
	res, err := a.cfg.client.PostForm(a.cfg.baseUrl+"/consultation/index.php", url.Values{
		"login":    {a.username},
		"password": {a.password},
		"submit":   {"Enter"},
	})

	if err != nil {
		return "", err
	}

	var value = ""
	for _, cookie := range res.Cookies() {
		if cookie.Name == "GAPSSESSID" {
			value = cookie.Value
		}
	}

	if value != "" {
		return value, nil
	}

	return "", errors.New("no token found in response")
}

func (a *LoginAction) FetchStudentId(token string) (uint, error) {
	tc := a.cfg.SetToken(token)

	req, err := tc.buildRequest("GET", "/consultation/etudiant/")
	res, err := tc.client.Do(req)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	pres, err := parser.FromResponseBody(res.Body)
	if err != nil {
		return 0, err
	}

	return pres.StudentId()
}
