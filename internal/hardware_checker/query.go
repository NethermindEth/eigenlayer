package hardwarechecker

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// QueryNodeExporter queries the Prometheus server at the specified address with the given query.
func QueryNodeExporter(address, query string) (float64, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return 0, fmt.Errorf("error creating client: %v", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, _, err := v1api.Query(ctx, query, time.Now(), v1.WithTimeout(5*time.Second))
	if err != nil {
		return 0, fmt.Errorf("error querying Prometheus: %v", err)
	}

	vectorResult, ok := result.(model.Vector)
	if !ok || len(vectorResult) == 0 {
		return 0, fmt.Errorf("no data found for query: %s", query)
	}

	// Return the first value
	return float64(vectorResult[0].Value), nil
}
