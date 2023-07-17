package node_exporter

import "errors"

var (
	ErrInvalidOptions      = errors.New("invalid options for grafana setup")
	ErrNonexistingEndpoint = errors.New("endpoint to remove does not exist")
)
