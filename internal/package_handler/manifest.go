package package_handler

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
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

	hardReqErr := m.HardwareRequirements.validate()

	var pluginErr error
	if m.Plugin != nil {
		pluginErr = m.Plugin.validate()
	}

	profileErr := errors.New("invalid profiles")
	invalidProfiles := false
	for i, profile := range m.Profiles {
		if err := profile.validate(i); err != nil {
			invalidProfiles = true
			profileErr = fmt.Errorf("%w: %w", profileErr, err)
		}
	}

	if hardReqErr != nil || pluginErr != nil || invalidProfiles || len(missingFields) > 0 {
		var err error = InvalidConfError{
			message:       "Invalid manifest file",
			missingFields: missingFields,
		}
		if hardReqErr != nil {
			err = fmt.Errorf("%w: %w", err, hardReqErr)
		}
		if pluginErr != nil {
			err = fmt.Errorf("%w: %w", err, pluginErr)
		}
		if invalidProfiles {
			err = fmt.Errorf("%w: %w", err, profileErr)
		}
		return err
	}

	return nil
}

type hardwareRequirements struct {
	MinCPUCores                 int  `yaml:"min_cpu_cores"`
	MinRAM                      int  `yaml:"min_ram"`
	MinFreeSpace                int  `yaml:"min_free_space"`
	StopIfRequirementsAreNotMet bool `yaml:"stop_if_requirements_are_not_met"`
}

func (h *hardwareRequirements) validate() error {
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
	return nil
}

type Plugin struct {
	Image     string `yaml:"image"`
	BuildFrom string `yaml:"build_from"`
}

func (p *Plugin) validate() error {
	var invalidFields []string
	// Validate plugin git field is a valid git url
	if p.BuildFrom != "" {
		_, err := url.ParseRequestURI(p.BuildFrom)
		if err != nil {
			invalidFields = append(invalidFields, "plugin.build_from -> (invalid build from)")
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
	return nil
}

type profileDefinition struct {
	Name        string      `yaml:"name"`
	FromProfile fromProfile `yaml:"from_profile"`
}

func (p *profileDefinition) validate(i int) error {
	var missingFields []string
	if p.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if len(missingFields) > 0 {
		return InvalidConfError{
			message:       "Profile #" + strconv.Itoa(i+1) + " is invalid",
			missingFields: missingFields,
		}
	}
	return nil
}

type fromProfile struct {
	Compose    string `yaml:"compose"`
	Env        string `yaml:"env"`
	Dashboards string `yaml:"dashboards"`
	Panels     string `yaml:"panels"`
	Alerts     string `yaml:"alerts"`
}
