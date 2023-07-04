package node_exporter

import (
	"github.com/NethermindEth/egn/internal/monitoring"
	"github.com/NethermindEth/egn/internal/monitoring/services/types"
)

var _ monitoring.ServiceAPI = &NodeExporterService{}

type NodeExporterService struct{}

func NewNodeExporter() *NodeExporterService {
	return &NodeExporterService{}
}

func (n *NodeExporterService) Init(opts types.ServiceOptions) error {
	return nil
}

func (n *NodeExporterService) AddTarget(endpoint string) error {
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

func (n *NodeExporterService) SetContainerIP(ip, containerName string) {
}

func (n *NodeExporterService) ContainerName() string {
	return monitoring.NodeExporterContainerName
}
