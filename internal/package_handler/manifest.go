package package_handler

import (
	"fmt"
	"net/url"
	"regexp"
)

// Manifest represents the manifest file of a package
type Manifest struct {
	Version              string               `yaml:"version"`
	NodeVersion          string               `yaml:"node_version"`
	Name                 string               `yaml:"name"`
	Upgrade              string               `yaml:"upgrade"`
	HardwareRequirements hardwareRequirements `yaml:"hardware_requirements"`
	Plugin               *Plugin              `yaml:"plugin"`
	Profiles             []profileDefinition  `yaml:"profiles"`
}

func (m *Manifest) validate() error {
	var missingFields []string
	if m.Version == "" {
		missingFields = append(missingFields, "version")
	}
	if m.NodeVersion == "" {
		missingFields = append(missingFields, "node_version")
	}
	if m.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if m.Upgrade == "" {
		missingFields = append(missingFields, "upgrade")
	}
	if len(m.Profiles) == 0 {
		missingFields = append(missingFields, "profiles")
	}

	hrErr := m.HardwareRequirements.validate()
	var pluErr InvalidConfError
	if m.Plugin != nil {
		pluErr = m.Plugin.validate()
	}

	var proErr InvalidConfError
	for i, profile := range m.Profiles {
		profileErr := profile.validate()
		if profileErr.message != "" {
			proErr = InvalidConfError{
				message:       fmt.Sprintf("Invalid profile %d", i+1),
				missingFields: profileErr.missingFields,
				invalidFields: profileErr.invalidFields,
			}
			break
		}
	}

	var mErr InvalidConfError
	if len(missingFields) > 0 {
		mErr = InvalidConfError{
			message:       "Invalid manifest file",
			missingFields: missingFields,
		}
	}

	var wrapped error
	if hrErr.message != "" || pluErr.message != "" || proErr.message != "" {
		wrapped = fmt.Errorf("%w %w %w", hrErr, pluErr, proErr)
		if mErr.message != "" {
			wrapped = fmt.Errorf("%w %w", mErr, wrapped)
		}
	} else if mErr.message != "" {
		wrapped = mErr
	}
	return wrapped
}

type hardwareRequirements struct {
	MinCPUCores                 int  `yaml:"min_cpu_cores"`
	MinRAM                      int  `yaml:"min_ram"`
	MinFreeSpace                int  `yaml:"min_free_space"`
	StopIfRequirementsAreNotMet bool `yaml:"stop_if_requirements_are_not_met"`
}

func (h *hardwareRequirements) validate() InvalidConfError {
	var invalidFields []string
	if h.MinCPUCores < 0 {
		invalidFields = append(invalidFields, "hardware_requirements.min_cpu_cores -> (negative value)")
	}
	if h.MinRAM < 0 {
		invalidFields = append(invalidFields, "hardware_requirements.min_ram -> (negative value)")
	}
	if h.MinFreeSpace < 0 {
		invalidFields = append(invalidFields, "hardware_requirements.min_free_space -> (negative value)")
	}
	if len(invalidFields) > 0 {
		return InvalidConfError{
			message:       "Invalid hardware requirements",
			invalidFields: invalidFields,
		}
	}
	return InvalidConfError{}
}

type Plugin struct {
	Image string `yaml:"image"`
	Git   string `yaml:"git"`
}

func (p *Plugin) validate() InvalidConfError {
	var invalidFields []string
	// Validate plugin git and image fields are not both empty
	if p.Git == "" && p.Image == "" {
		invalidFields = append(invalidFields, "plugin.git, plugin.image -> (both empty)")
	}
	// Validate plugin git field is a valid git url
	if p.Git != "" {
		uri, err := url.ParseRequestURI(p.Git)
		if err != nil || uri.Scheme == "" || uri.Host == "" || (uri.Scheme != "https" && uri.Scheme != "http") {
			invalidFields = append(invalidFields, "plugin.git -> (invalid git url)")
		}
	}
	// Validate plugin image field is a valid docker image
	if p.Image != "" {
		re := regexp.MustCompile(`^([\w-]+\/)?([\w-]+)(:[\w-\.]+)?$`)
		if !re.MatchString(p.Image) {
			invalidFields = append(invalidFields, "plugin.image -> (invalid docker image)")
		}
	}
	if len(invalidFields) > 0 {
		return InvalidConfError{
			message:       "Invalid plugin",
			invalidFields: invalidFields,
		}
	}
	return InvalidConfError{}
}

type profileDefinition struct {
	Name        string      `yaml:"name"`
	FromProfile fromProfile `yaml:"from_profile"`
}

func (p *profileDefinition) validate() InvalidConfError {
	var missingFields []string
	if p.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if len(missingFields) > 0 {
		return InvalidConfError{
			message:       "Invalid profile",
			missingFields: missingFields,
		}
	}
	return InvalidConfError{}
}

type fromProfile struct {
	Compose    string `yaml:"compose"`
	Env        string `yaml:"env"`
	Dashboards string `yaml:"dashboards"`
	Panels     string `yaml:"panels"`
	Alerts     string `yaml:"alerts"`
}
