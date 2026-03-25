package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/metadata/pkg/model"
	"github.com/rahulbalajee/Movie/movie/internal/gateway"
	"github.com/rahulbalajee/Movie/pkg/discovery"
)

// Gateway defines a movie metadata HTTP gateway
type Gateway struct {
	registry discovery.Registry
	client   *http.Client
}

// New creates a new HTTP gateway for movie metadata service
func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry: registry,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Get gets movie metadata by a movie id
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	addrs, err := g.registry.ServiceAddresses(ctx, "metadata")
	if err != nil {
		return nil, err
	}
	url := "http://" + addrs[rand.IntN(len(addrs))] + "/metadata"
	log.Printf("calling metadata service. Request: GET %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

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
