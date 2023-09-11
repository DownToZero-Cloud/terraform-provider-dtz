package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client -
type Client struct {
	ApiUrl     string
	HTTPClient *http.Client
	ApiKey     string `json:"apiKey"`
}

func NewClient(section *string, version *string, apikey *string) (*Client, error) {

	if *apikey == "" {
		return nil, fmt.Errorf("define apikey")
	}
	if *version == "" {
		return nil, fmt.Errorf("define version")
	}
	if *section == "" {
		return nil, fmt.Errorf("define section")
	}

	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		ApiUrl:     "https://" + *section + ".dtz.rocks/api/" + *version,
		ApiKey:     *apikey,
	}

	if apikey == nil {
		return &c, nil
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) (int, []byte, error) {

	// set auth headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", c.ApiKey)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	return res.StatusCode, body, err
}
