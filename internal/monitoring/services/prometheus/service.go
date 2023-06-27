package prometheus

import (
	"embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/NethermindEth/egn/internal/data"
	"github.com/NethermindEth/egn/internal/monitoring"
	"github.com/NethermindEth/egn/internal/monitoring/services/types"
	"gopkg.in/yaml.v3"
)

//go:embed config
var config embed.FS

// Config represents the Prometheus configuration.
type Config struct {
	Global        GlobalConfig   `yaml:"global"`
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

// GlobalConfig represents the global configuration for Prometheus.
type GlobalConfig struct {
	ScrapeInterval string `yaml:"scrape_interval"`
}

// ScrapeConfig represents the configuration for a Prometheus scrape job.
type ScrapeConfig struct {
	JobName       string         `yaml:"job_name"`
	StaticConfigs []StaticConfig `yaml:"static_configs"`
}

// StaticConfig represents the static configuration for a Prometheus scrape job.
type StaticConfig struct {
	Targets []string `yaml:"targets"`
}

// Verify that PrometheusService implements the ServiceAPI interface.
var _ monitoring.ServiceAPI = &PrometheusService{}

// PrometheusService implements the ServiceAPI interface for a Prometheus service.
type PrometheusService struct {
	stack    *data.MonitoringStack
	endpoint string
}

// NewPrometheus creates a new PrometheusService.
func NewPrometheus() *PrometheusService {
	return &PrometheusService{}
}

// Init initializes the Prometheus service with the given options.
func (p *PrometheusService) Init(opts types.ServiceOptions) error {
	// Validate dotEnv
	promPort, ok := opts.Dotenv["PROM_PORT"]
	if !ok {
		return fmt.Errorf("%w: %s missing in options", ErrInvalidOptions, "PROM_PORT")
	} else if promPort == "" {
		return fmt.Errorf("%w: %s can't be empty", ErrInvalidOptions, "PROM_PORT")
	}

	p.stack = opts.Stack
	p.endpoint = fmt.Sprintf("http://%s:%s", monitoring.PrometheusServiceName, opts.Dotenv["PROM_PORT"])
	return nil
}

// AddTarget adds a new target to the Prometheus config and reloads the Prometheus configuration.
// Assumes endpoint is in the form http://<ip/domain>:<port>
func (p *PrometheusService) AddTarget(endpoint string) error {
	path := filepath.Join("prometheus", "prometheus.yml")
	// Read the existing config
	rawConfig, err := p.stack.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	if err = yaml.Unmarshal(rawConfig, &config); err != nil {
		return err
	}

	// Add the new endpoint to the targets of the first scrape config
	endpoint = strings.TrimPrefix(endpoint, "http://")
	// Check if the endpoint is already in the targets
	for _, target := range config.ScrapeConfigs[0].StaticConfigs[0].Targets {
		if target == endpoint {
			// There is no need to add the endpoint if it is already in the targets
			return nil
		}
	}
	config.ScrapeConfigs[0].StaticConfigs[0].Targets = append(config.ScrapeConfigs[0].StaticConfigs[0].Targets, endpoint)

	// Marshal the updated config back to YAML
	newConfig, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// Write the updated YAML data back to the file
	if err = p.stack.WriteFile(path, newConfig); err != nil {
		return err
	}

	// Reload the config
	if err = p.reloadConfig(); err != nil {
		return err
	}

	return nil
}

// RemoveTarget removes a target from the Prometheus config and reloads the Prometheus configuration.
// Assumes endpoint is in the form http://<ip/domain>:<port>
func (p *PrometheusService) RemoveTarget(endpoint string) error {
	path := filepath.Join("prometheus", "prometheus.yml")
	// Read the existing config
	rawConfig, err := p.stack.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	if err = yaml.Unmarshal(rawConfig, &config); err != nil {
		return err
	}

	// Remove the endpoint from the targets of the first scrape config
	prevLen := len(config.ScrapeConfigs[0].StaticConfigs[0].Targets)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	for i, target := range config.ScrapeConfigs[0].StaticConfigs[0].Targets {
		if target == endpoint {
			config.ScrapeConfigs[0].StaticConfigs[0].Targets = append(config.ScrapeConfigs[0].StaticConfigs[0].Targets[:i], config.ScrapeConfigs[0].StaticConfigs[0].Targets[i+1:]...)
			break
		}
	}

	// Check if the endpoint was removed
	if len(config.ScrapeConfigs[0].StaticConfigs[0].Targets) == prevLen {
		// The endpoint was not removed because it was not in the targets
		return fmt.Errorf("%w: %s", ErrNonexistingEndpoint, endpoint)
	}

	// Marshal the updated config back to YAML
	newConfig, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// Write the updated YAML data back to the file
	if err = p.stack.WriteFile(path, newConfig); err != nil {
		return err
	}

	// Reload the config
	if err = p.reloadConfig(); err != nil {
		return err
	}

	return nil
}

// DotEnv returns the dotenv variables and default values for the Prometheus service.
func (p *PrometheusService) DotEnv() map[string]string {
	return dotEnv
}

// Setup sets up the Prometheus service configuration files with the given dotenv values.
func (p *PrometheusService) Setup(options map[string]string) error {
	// Validate options
	nodeExporterPort, ok := options["NODE_EXPORTER_PORT"]
	if !ok {
		return fmt.Errorf("%w: %s missing in options", ErrInvalidOptions, "NODE_EXPORTER_PORT")
	} else if nodeExporterPort == "" {
		return fmt.Errorf("%w: %s can't be empty", ErrInvalidOptions, "NODE_EXPORTER_PORT")
	}

	// Read config from the embedded FS
	rawConfig, err := config.ReadFile("config/prometheus.yml")
	if err != nil {
		return err
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	if err = yaml.Unmarshal(rawConfig, &config); err != nil {
		return err
	}

	// Add node exporter target
	config.ScrapeConfigs[0].StaticConfigs[0].Targets = append(config.ScrapeConfigs[0].StaticConfigs[0].Targets, fmt.Sprintf("%s:%s", monitoring.NodeExporterServiceName, options["NODE_EXPORTER_PORT"]))

	// Marshal the updated config back to YAML
	newConfig, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// Create config directory
	if err = p.stack.CreateDir("prometheus"); err != nil {
		return err
	}

	// Write the updated YAML data to datadir
	if err = p.stack.WriteFile("prometheus/prometheus.yml", newConfig); err != nil {
		return err
	}

	return nil
}

// reloadConfig reloads the Prometheus config by making a POST request to the /-/reload endpoint
func (p *PrometheusService) reloadConfig() error {
	resp, err := http.Post(fmt.Sprintf("%s/-/reload", p.endpoint), "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrReloadFailed, resp.Status)
	}

	return nil
}
