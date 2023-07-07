package package_handler

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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
		case "str":
			if o.ValidateDef != nil {
				invalidDefault = !regexp.MustCompile(o.ValidateDef.Re2Regex).MatchString(o.Default)
			}
		case "int":
			val, err := strconv.Atoi(o.Default)
			invalidDefault = err != nil
			if o.ValidateDef != nil {
				if o.ValidateDef.MinValue != nil && val < int(*o.ValidateDef.MinValue) {
					invalidDefault = true
				}
				if o.ValidateDef.MaxValue != nil && val > int(*o.ValidateDef.MaxValue) {
					invalidDefault = true
				}
			}
		case "port":
			port, err := strconv.Atoi(o.Default)
			invalidDefault = err != nil || port <= 0 || port > 65535
		case "float":
			val, err := strconv.ParseFloat(o.Default, 64)
			invalidDefault = err != nil
			if o.ValidateDef != nil {
				if o.ValidateDef.MinValue != nil && val < *o.ValidateDef.MinValue {
					invalidDefault = true
				}
				if o.ValidateDef.MaxValue != nil && val > *o.ValidateDef.MaxValue {
					invalidDefault = true
				}
			}
		case "bool":
			_, err := strconv.ParseBool(o.Default)
			invalidDefault = err != nil
		case "path_dir":
			invalidDefault = !pathRe.MatchString(o.Default)
		case "path_file":
			invalidDefault = !pathRe.MatchString(o.Default)
			if o.ValidateDef != nil {
				invalidDefault = filepath.Ext(o.Default) != o.ValidateDef.Format
			}
		case "uri":
			gotUrl, err := url.Parse(o.Default)
			if err != nil {
				invalidDefault = true
			} else if o.ValidateDef != nil {
				invalidDefault = !utils.Contains(o.ValidateDef.UriScheme, gotUrl.Scheme)
			}
		case "select":
			if o.ValidateDef == nil {
				missingFields = append(missingFields, "options.validate")
			} else {
				invalidDefault = !utils.Contains(o.ValidateDef.Options, o.Default)
			}
		case "id":
			invalidDefault = len(strings.Split(o.Default, " ")) != 1
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
	MinValue  *float64 `yaml:"min_value,omitempty"`
	MaxValue  *float64 `yaml:"max_value,omitempty"`
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
