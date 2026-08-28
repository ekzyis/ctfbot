package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"

	t "github.com/ekzyis/snappy/types"
)

const userAgent = "Mozilla/5.0 (X11; Linux x86_64) Chrome/120.0"

type Client struct {
	BaseUrl  string
	ApiUrl   string
	ApiKey   string
	Nsec     string
	MediaUrl string

	httpClient     *http.Client
	loggedIn       bool
	authenticating bool
}

func NewClient(options ...func(*Client)) *Client {
	c := &Client{}
	for _, o := range options {
		o(c)
	}

	// set defaults
	var ok bool
	if c.BaseUrl == "" {
		c.BaseUrl, ok = os.LookupEnv("SN_BASE_URL")
		if !ok {
			c.BaseUrl = "https://stacker.news"
		}
	}
	if c.ApiKey == "" {
		c.ApiKey = os.Getenv("SN_API_KEY")
	}
	if c.Nsec == "" {
		c.Nsec = os.Getenv("SN_NSEC")
	}
	if c.MediaUrl == "" {
		c.MediaUrl, ok = os.LookupEnv("SN_MEDIA_URL")
		if !ok {
			c.MediaUrl = "https://m.stacker.news"
		}
	}
	c.ApiUrl = fmt.Sprintf("%s/api/graphql", c.BaseUrl)

	jar, _ := cookiejar.New(nil)
	c.httpClient = &http.Client{Jar: jar}

	return c
}

func WithApiKey(apiKey string) func(*Client) {
	return func(c *Client) {
		c.ApiKey = apiKey
	}
}

func WithNsec(nsec string) func(*Client) {
	return func(c *Client) {
		c.Nsec = nsec
	}
}

func WithBaseUrl(baseUrl string) func(*Client) {
	return func(c *Client) {
		c.BaseUrl = baseUrl
	}
}

func WithMediaUrl(mediaUrl string) func(*Client) {
	return func(c *Client) {
		c.MediaUrl = mediaUrl
	}
}

func (c *Client) callApi(body t.GqlBody) (*http.Response, error) {
	if err := c.ensureLoggedIn(); err != nil {
		return nil, err
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		err = fmt.Errorf("error encoding SN payload: %w", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ApiUrl, bytes.NewBuffer(bodyJSON))
	if err != nil {
		err = fmt.Errorf("error preparing SN request: %w", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if c.ApiKey != "" {
		req.Header.Set("X-Api-Key", c.ApiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) checkForErrors(err []t.GqlError) error {
	if len(err) > 0 {
		errMsg, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			return marshalErr
		}
		return errors.New(string(errMsg))
	}
	return nil
}

func (c *Client) checkForPayInErrors(payIn t.PayIn) error {
	privates := payIn.PayerPrivates
	if privates.PayInFailureReason != "" {
		return fmt.Errorf("mutation failed: %s", privates.PayInFailureReason)
	}

	bolt11 := privates.PayInBolt11
	if bolt11.Id != 0 {
		return fmt.Errorf("mutation failed: bolt11 payment required")
	}

	result := privates.Result
	if result.Id == 0 {
		return fmt.Errorf("mutation failed: no result id")
	}

	return nil
}
