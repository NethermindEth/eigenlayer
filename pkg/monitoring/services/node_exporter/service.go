package node_exporter

import (
	"fmt"

	"github.com/NethermindEth/eigenlayer/pkg/monitoring"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/types"
)

var _ monitoring.ServiceAPI = &NodeExporterService{}

type NodeExporterService struct {
	containerIP string
	port        string
}

func NewNodeExporter() *NodeExporterService {
	return &NodeExporterService{}
}

func (n *NodeExporterService) Init(opts types.ServiceOptions) error {
	// Validate dotEnv
	nodeExporterPort, ok := opts.Dotenv["NODE_EXPORTER_PORT"]
	if !ok {
		return fmt.Errorf("%w: %s missing in options", ErrInvalidOptions, "NODE_EXPORTER_PORT")
	} else if nodeExporterPort == "" {
		return fmt.Errorf("%w: %s can't be empty", ErrInvalidOptions, "NODE_EXPORTER_PORT")
	}

	n.port = opts.Dotenv["NODE_EXPORTER_PORT"]
	return nil
}

func (n *NodeExporterService) AddTarget(endpoint, instanceID string) error {
	return nil
}

func (n *NodeExporterService) RemoveTarget(endpoint string) error {
	return nil
}

func (n *NodeExporterService) DotEnv() map[string]string {
	return dotEnv
}

func (n *NodeExporterService) Setup(options map[string]string) error {
	return nil
}

func (n *NodeExporterService) SetContainerIP(ip string) {
	n.containerIP = ip
}

func (n *NodeExporterService) ContainerName() string {
	return monitoring.NodeExporterContainerName
}

func (n *NodeExporterService) Endpoint() string {
	return fmt.Sprintf("http://%s:%s", n.containerIP, n.port)
}
