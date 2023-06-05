package package_handler

import (
	"fmt"
	"regexp"
)

type Manifest struct {
	Version              string               `yaml:"version"`
	NodeVersion          string               `yaml:"node_version"`
	Name                 string               `yaml:"name"`
	Upgrade              string               `yaml:"upgrade"`
	HardwareRequirements HardwareRequirements `yaml:"hardware_requirements"`
	Plugin               Plugin               `yaml:"plugin"`
	Profiles             []Profile            `yaml:"profiles"`
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
	pluErr := m.Plugin.validate()

	var proErr InvalidManifestError
	if len(m.Profiles) > 0 {
		for i, profile := range m.Profiles {
			profileErr := profile.validate()
			if profileErr.message != "" {
				proErr = InvalidManifestError{
					message:       fmt.Sprintf("Invalid profile %d", i),
					missingFields: profileErr.missingFields,
					invalidFields: profileErr.invalidFields,
				}
				break
			}
		}
	}

	var mErr InvalidManifestError
	if len(missingFields) > 0 {
		mErr = InvalidManifestError{
			message:       "Invalid manifest file",
			missingFields: missingFields,
		}
	}

	var wrapped error
	if hrErr.message != "" || pluErr.message != "" || proErr.message != "" {
		wrapped = fmt.Errorf("%w %w %w", hrErr, pluErr, proErr)
		if mErr.message != "" {
			wrapped = fmt.Errorf("%w: %w", mErr, wrapped)
		}
	} else if mErr.message != "" {
		wrapped = mErr
	}
	return wrapped
}

type HardwareRequirements struct {
	MinCPUCores                 int  `yaml:"min_cpu_cores"`
	MinRAM                      int  `yaml:"min_ram"`
	MinFreeSpace                int  `yaml:"min_free_space"`
	StopIfRequirementsAreNotMet bool `yaml:"stop_if_requirements_are_not_met"`
}

func (h *HardwareRequirements) validate() InvalidManifestError {
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
		return InvalidManifestError{
			message:       "Invalid hardware requirements",
			invalidFields: invalidFields,
		}
	}
	return InvalidManifestError{}
}

type Plugin struct {
	Image string `yaml:"image"`
	Git   string `yaml:"git"`
}

func (p *Plugin) validate() InvalidManifestError {
	var invalidFields []string
	// Validate plugin git field is a valid git url
	if p.Git != "" {
		re := regexp.MustCompile(`^(https:\/\/github\.com\/|https:\/\/gitlab\.com\/|git@github\.com:|git@gitlab\.com:)[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+(\.git)?$`)
		if !re.MatchString(p.Git) {
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
		return InvalidManifestError{
			message:       "Invalid plugin",
			invalidFields: invalidFields,
		}
	}
	return InvalidManifestError{}
}

type Profile struct {
	Name        string      `yaml:"name"`
	FromProfile FromProfile `yaml:"from_profile"`
}

func (p *Profile) validate() InvalidManifestError {
	var missingFields []string
	if p.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if len(missingFields) > 0 {
		return InvalidManifestError{
			message:       "Invalid profile",
			missingFields: missingFields,
		}
	}
	return InvalidManifestError{}
}

type FromProfile struct {
	Compose    string `yaml:"compose"`
	Env        string `yaml:"env"`
	Dashboards string `yaml:"dashboards"`
	Panels     string `yaml:"panels"`
	Alerts     string `yaml:"alerts"`
}
