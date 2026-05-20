package isdayoff

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"payment_bot/internal/entity"
)

const baseURL = "https://isdayoff.ru"

type Client struct {
	http *http.Client
}

func New(httpClient *http.Client) *Client {
	return &Client{http: httpClient}
}

func (c *Client) FetchWorkdayCalendar(ctx context.Context, year int, month time.Month) (*entity.WorkdayCalendar, error) {
	url := fmt.Sprintf("%s/api/getdata?year=%d&month=%02d", baseURL, year, int(month))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("isdayoff request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("isdayoff returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	data := string(body)
	days := make([]bool, len(data))
	for i, ch := range data {
		days[i] = ch == '0'
	}

	return &entity.WorkdayCalendar{Year: year, Month: month, Days: days}, nil
}
