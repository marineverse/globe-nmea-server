package nmea

import (
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Client) FetchNMEASentence(boatUUID string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/globe/boats/%s/nmea0183", c.baseURL, boatUUID)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch data: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return string(body), nil
}