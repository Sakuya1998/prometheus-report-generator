package client

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// CreatePrometheusClient creates a new Prometheus client.
func CreatePrometheusClient(url string) (v1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(client), nil
}

// QueryRange queries Prometheus within a specified time range with a given step.
func QueryRange(client v1.API, query string, start, end time.Time, step time.Duration) (model.Value, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}
	result, warnings, err := client.QueryRange(context.Background(), query, r)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		// Handle warnings if necessary
		return result, nil
	}
	return result, nil
}
