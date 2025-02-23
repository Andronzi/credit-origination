package client

import (
	"context"
	"time"
)

type ScoringClient struct {
	baseURL string
	timeout time.Duration
}

func NewScoringClient(baseURL string) *ScoringClient {
	return &ScoringClient{
		baseURL: baseURL,
		timeout: 5 * time.Second,
	}
}

func (c *ScoringClient) GetScore(ctx context.Context, userID string) (int, error) {
	return 750, nil
}
