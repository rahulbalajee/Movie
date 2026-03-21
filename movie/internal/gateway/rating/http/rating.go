package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/movie/internal/gateway"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
)

type Gateway struct {
	addr   string
	client *http.Client
}

func New(addr string) *Gateway {
	return &Gateway{
		addr:   addr,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *Gateway) GetAggregatedRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.addr+"/rating", nil)
	if err != nil {
		return 0, err
	}

	values := req.URL.Query()
	values.Add("id", string(recordId))
	values.Add("type", string(recordType))
	req.URL.RawQuery = values.Encode()

	resp, err := g.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return 0, fmt.Errorf("non-2xx response: %v", resp)
	}

	var v float64
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return 0, err
	}
	return v, nil
}

func (g *Gateway) PutRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	body, err := json.Marshal(map[string]any{
		"id":     string(recordId),
		"type":   string(recordType),
		"userId": string(rating.UserID),
		"value":  rating.Value,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, g.addr+"/rating", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-2xx response: %v", resp)
	}
	return nil
}
