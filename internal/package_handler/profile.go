package package_handler

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/docker/distribution/uuid"

	"github.com/NethermindEth/egn/internal/utils"
)

var pathRe = regexp.MustCompile(`^(/|./|../|[^/ ]([^/ ]*/)*[^/ ]*$)`)

// Profile represents a profile file of a package
type Profile struct {
	Name                          string                        `yaml:"-"`
	HardwareRequirementsOverrides HardwareRequirementsOverrides `yaml:"hardware_requirements_overrides"`
	PluginOverrides               PluginOverrides               `yaml:"plugin_overrides"`
	Options                       []Option                      `yaml:"options"`
	Monitoring                    Monitoring                    `yaml:"monitoring"`
}

// Validate validates the profile file
func (p *Profile) Validate() error {
	var missingFields []string
	if len(p.Options) == 0 {
		missingFields = append(missingFields, "options")
	}

	var invalidOptionsErr error
	for i, option := range p.Options {
		if err := option.validate(); len(err.missingFields) > 0 || len(err.invalidFields) > 0 {
			err.message = fmt.Sprintf("Invalid option %d", i+1)
			invalidOptionsErr = fmt.Errorf("%w %w", invalidOptionsErr, err)
		}
	}

	var invalidProfileErr InvalidConfError
	if len(missingFields) > 0 {
		invalidProfileErr = InvalidConfError{
			message:       "Invalid profile.yml",
			missingFields: missingFields,
		}
	}

	if invalidProfileErr.message != "" || invalidOptionsErr != nil {
		return fmt.Errorf("%w %w", invalidProfileErr, invalidOptionsErr)
	}

	return nil
}

// HardwareRequirementsOverrides represents the hardware requirements overrides field of a profile
type HardwareRequirementsOverrides struct {
	MinCPUCores  int `yaml:"min_cpu_cores"`
	MinRAM       int `yaml:"min_ram"`
	MinFreeSpace int `yaml:"min_free_space"`
}

// TODO: add validation for hardware requirements overrides

// PluginOverrides represents the plugin overrides field of a profile
type PluginOverrides struct {
	Image string `yaml:"image"`
	Git   string `yaml:"git"`
}

// TODO: add validation for plugin overrides

// Option represents an option within the options field of a profile
type Option struct {
	Name        string    `yaml:"name"`
	Target      string    `yaml:"target"`
	Type        string    `yaml:"type"`
	Default     string    `yaml:"default"`
	Help        string    `yaml:"help"`
	ValidateDef *Validate `yaml:"validate,omitempty"`
}

// Validate validates the option
func (o *Option) validate() InvalidConfError {
	var missingFields, invalidFields []string
	if o.Name == "" {
		missingFields = append(missingFields, "options.name")
	}
	if o.Target == "" {
		missingFields = append(missingFields, "options.target")
	}
	if o.Type == "" {
		missingFields = append(missingFields, "options.type")
	}
	if o.Help == "" {
		missingFields = append(missingFields, "options.help")
	}

	var invalidDefault bool
	if o.Default != "" {
		switch o.Type {
		case "string":
			invalidDefault = true
		case "int", "port":
			_, err := strconv.Atoi(o.Default)
			invalidDefault = err != nil
		case "float":
			_, err := strconv.ParseFloat(o.Default, 64)
			invalidDefault = err != nil
		case "bool":
			_, err := strconv.ParseBool(o.Default)
			invalidDefault = err != nil
		case "path_dir", "path_file":
			invalidDefault = !pathRe.MatchString(o.Default)
		case "uri":
			_, err := url.Parse(o.Default)
			invalidDefault = err != nil
		case "select":
			if o.ValidateDef == nil {
				missingFields = append(missingFields, "options.validate")
			} else {
				invalidDefault = !utils.Contains(o.ValidateDef.Options, o.Default)
			}
		case "id":
			_, err := uuid.Parse(o.Default)
			invalidDefault = err != nil
		default:
			invalidDefault = true
		}
	}
	if invalidDefault {
		invalidFields = append(invalidFields, "options.default")
	}

	if len(missingFields) > 0 || len(invalidFields) > 0 {
		return InvalidConfError{
			missingFields: missingFields,
			invalidFields: invalidFields,
		}
	}

	return InvalidConfError{}
}

// Validate represents the validate field of an option
type Validate struct {
	Re2Regex  string   `yaml:"re2_regex"`
	Format    string   `yaml:"format"`
	UriScheme []string `yaml:"uri_scheme"`
	MinValue  float64  `yaml:"min_value"`
	MaxValue  float64  `yaml:"max_value"`
	Options   []string `yaml:"options"`
}

// Monitoring represents the monitoring field of a profile
type Monitoring struct {
	Tag     string             `yaml:"tag"`
	Targets []MonitoringTarget `yaml:"targets"`
}

// MonitoringTarget represents a monitoring target within the targets field of a monitoring
type MonitoringTarget struct {
	Service string `yaml:"service"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// TODO: add validation for monitoring
