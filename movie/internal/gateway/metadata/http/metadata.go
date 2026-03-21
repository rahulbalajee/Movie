package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/metadata/pkg/model"
	"github.com/rahulbalajee/Movie/movie/internal/gateway"
)

type Gateway struct {
	client *http.Client
	addr   string
}

func New(addr string) *Gateway {
	return &Gateway{
		addr:   addr,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	req, err := http.NewRequest(http.MethodGet, g.addr+"/metadata", nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	values := req.URL.Query()
	values.Add("id", id)
	req.URL.RawQuery = values.Encode()

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-2xx response: %v", resp)
	}

	var v *model.Metadata
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
