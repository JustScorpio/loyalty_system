package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JustScorpio/loyalty_system/internal/customerrors"
)

// Client оборачивает работу с API начислений
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New создает клиент с настройками
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// OrderResponse ответ API начислений
type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

// GetOrderInfo получает данные о начислении
func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, customerrors.NewHTTPError(errors.New("unexpected status"), resp.StatusCode)
	}

	var data OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("json decode failed: %w", err)
	}

	return &data, nil
}
