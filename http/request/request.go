package request

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	cli *http.Client
}

func NewClient(cli *http.Client) *Client {
	return &Client{cli: cli}
}

func (c *Client) Do(ctx context.Context, to string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", to, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return b, nil
}
