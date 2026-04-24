package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"omctl/config"
)

type Client struct {
	Host  string
	Token string
	HTTP  *http.Client
}

func New(cfg *config.Config) *Client {
	if cfg.Token == "" {
		panic("OM_TOKEN not set. RUN: export OM_TOKEN=your_token")
	}
	return &Client{
		Host:  cfg.Host,
		Token: cfg.Token,
		HTTP:  &http.Client{},
	}
}

func (c *Client) Get(path string, params map[string]string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/v1/%s", c.Host, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	// query params
	query := req.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
