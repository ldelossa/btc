package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// base url to contact binance api
	baseURL = "api.binance.com"
	// path to ticker prices
	tickerPath = "/api/v3/ticker/price"
)

// TickerPrice is the returned json schema from a call to tickerPath
type TickerPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"` // string tag tells unmarshaller this is a string encoded float
}

// Client is an http client for contacting binance api
type Client struct {
	c *http.Client
}

// NewClient is a constructor for our client
func NewClient() *Client {
	c := &Client{
		c: &http.Client{},
	}

	return c
}

// GetPrice retrieves ticker prices. It takes a symbol string and returns the TickerPrice response.
// In event of error we return the status code and status
func (c *Client) GetPrice(symbol string) (*TickerPrice, error) {
	values := url.Values{"symbol": []string{symbol}}

	req := &http.Request{
		URL: &url.URL{
			Scheme:   "https",
			Host:     baseURL,
			Path:     tickerPath,
			RawQuery: values.Encode(),
		},
		Method: "GET",
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GetPrice: did not receive 200. Code: %d Status: %s", resp.StatusCode, resp.Status)
	}

	var tp TickerPrice
	err = json.NewDecoder(resp.Body).Decode(&tp)
	if err != nil {
		return nil, fmt.Errorf("GetPrice: failed to deserialize response body: %s", err.Error())
	}

	return &tp, nil
}
