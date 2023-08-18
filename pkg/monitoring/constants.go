package monitoring

const (
	PrometheusServiceName     = "prometheus"
	PrometheusContainerName   = "egn_prometheus"
	GrafanaServiceName        = "grafana"
	GrafanaContainerName      = "egn_grafana"
	NodeExporterServiceName   = "node_exporter"
	NodeExporterContainerName = "egn_node_exporter"
	monitoringPath            = "monitoring"
	InstanceIDLabel           = "instance_id"
	CommitHashLabel           = "instance_commit_hash"
	AVSNameLabel              = "avs_name"
	AVSVersionLabel           = "avs_version"
	SpecVersionLabel          = "spec_version"
)
