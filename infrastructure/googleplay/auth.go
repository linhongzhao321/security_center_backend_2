package googleplay

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ParseQuery(query string) (url.Values, error) {
	query = strings.ReplaceAll(query, "\n", "&")
	return url.ParseQuery(query)
}

type GoogleAuth struct {
	v url.Values
}

func (g *GoogleAuth) Auth(ctx context.Context, token Token) error {
	data := url.Values{
		"Token":      {token.getToken()},
		"app":        {"com.android.vending"},
		"client_sig": {"38918a453d07199354f8b19af05ec6562ced5788"},
		"service":    {"oauth2:https://www.googleapis.com/auth/googleplay"},
	}
	bodyReader := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", "https://android.googleapis.com/auth", bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		var b strings.Builder
		_ = res.Write(&b)
		return errors.New(b.String())
	}
	text, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	g.v, err = ParseQuery(string(text))
	if err != nil {
		return err
	}
	return nil
}

func (g GoogleAuth) getAuth() string {
	return g.v.Get("Auth")
}

type Token struct {
	Data []byte
	v    url.Values
}

func (g *Token) Auth(oauthToken string) error {
	res, err := http.PostForm(
		"https://android.googleapis.com/auth", url.Values{
			"ACCESS_TOKEN": {"1"},
			"Token":        {oauthToken},
			"service":      {"ac2dm"},
		},
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var b strings.Builder
		_ = res.Write(&b)
		return errors.New(b.String())
	}
	g.Data, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return nil
}

func (g *Token) Unmarshal() error {
	var err error
	g.v, err = ParseQuery(string(g.Data))
	if err != nil {
		return err
	}
	return nil
}

func (g Token) getToken() string {
	return g.v.Get("Token")
}
