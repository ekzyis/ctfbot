package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) ensureLoggedIn() error {
	if c.Nsec == "" || c.loggedIn || c.authenticating {
		return nil
	}

	// guard against recursion: login() itself calls the GraphQL API (createAuth)
	c.authenticating = true
	defer func() { c.authenticating = false }()

	if err := c.login(); err != nil {
		return fmt.Errorf("nostr login failed: %w", err)
	}
	c.loggedIn = true
	return nil
}

func (c *Client) login() error {
	csrfToken, err := c.fetchCsrfToken()
	if err != nil {
		return err
	}

	k1, err := c.createAuth()
	if err != nil {
		return err
	}

	event, err := signAuthEvent(c.Nsec, k1, c.BaseUrl)
	if err != nil {
		return err
	}

	return c.postNostrCallback(csrfToken, event)
}

func (c *Client) fetchCsrfToken() (string, error) {
	endpoint := fmt.Sprintf("%s/api/auth/csrf", c.BaseUrl)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		CsrfToken string `json:"csrfToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("error decoding csrf token: %w", err)
	}
	if body.CsrfToken == "" {
		return "", errors.New("no csrf token in response")
	}
	return body.CsrfToken, nil
}

func (c *Client) createAuth() (string, error) {
	body := t.GqlBody{
		Query: `
		mutation createAuth {
			createAuth {
				k1
			}
		}`,
	}
	resp, err := c.callApi(body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respBody struct {
		Data struct {
			CreateAuth struct {
				K1 string `json:"k1"`
			} `json:"createAuth"`
		} `json:"data"`
		Errors []t.GqlError `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("error decoding createAuth: %w", err)
	}
	if err := c.checkForErrors(respBody.Errors); err != nil {
		return "", err
	}
	if respBody.Data.CreateAuth.K1 == "" {
		return "", errors.New("no k1 challenge in response")
	}
	return respBody.Data.CreateAuth.K1, nil
}

func (c *Client) postNostrCallback(csrfToken, event string) error {
	form := url.Values{}
	form.Set("csrfToken", csrfToken)
	form.Set("event", event)
	form.Set("callbackUrl", c.BaseUrl)
	form.Set("json", "true")

	endpoint := fmt.Sprintf("%s/api/auth/callback/nostr", c.BaseUrl)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body struct {
		Url string `json:"url"`
	}
	// body is best-effort; the session cookie is the source of truth
	_ = json.NewDecoder(resp.Body).Decode(&body)

	if resp.StatusCode != http.StatusOK || !c.hasSessionCookie() {
		if body.Url != "" {
			return fmt.Errorf("login rejected (status %d): %s", resp.StatusCode, body.Url)
		}
		return fmt.Errorf("login rejected (status %d): no session cookie set", resp.StatusCode)
	}
	return nil
}

func (c *Client) hasSessionCookie() bool {
	u, err := url.Parse(c.BaseUrl)
	if err != nil {
		return false
	}
	for _, cookie := range c.httpClient.Jar.Cookies(u) {
		if strings.Contains(cookie.Name, "session-token") {
			return true
		}
	}
	return false
}
