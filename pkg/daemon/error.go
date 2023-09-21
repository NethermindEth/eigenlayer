package daemon

import "errors"

var (
	ErrInstanceAlreadyExists      = errors.New("instance already exists")
	ErrProfileDoesNotExist        = errors.New("profile does not exist")
	ErrInstanceNotRunning         = errors.New("instance is not running")
	ErrInstanceNotFound           = errors.New("instance not found")
	ErrOptionWithoutValue         = errors.New("option without value")
	ErrMonitoringTargetPortNotSet = errors.New("monitoring target port is not set")
	ErrInstanceHasNoPlugin        = errors.New("instance has no plugin")
	ErrVersionOrCommitNotSet      = errors.New("version or commit not set")
	ErrPluginPathNotInsidePackage = errors.New("plugin path is not inside package")
	ErrUnknownPluginType          = errors.New("unknown plugin type")
	ErrInvalidUpdateVersion       = errors.New("invalid update version")
	ErrInvalidUpdateCommit        = errors.New("invalid update commit")
	ErrOptionNotSet               = errors.New("option not set")
	ErrVersionAlreadyInstalled    = errors.New("version already installed")
)

// InvalidOptionValueError is returned when an Option's value is invalid.
type InvalidOptionValueError struct {
	optionName string
	value      string
	msg        string
	hidden     bool
}

func (e InvalidOptionValueError) Error() string {
	if e.hidden {
		return "invalid value for option " + e.optionName + ": " + e.msg
	}
	return "invalid value for option " + e.optionName + ": " + e.value + ". " + e.msg
}

// InvalidRegexError is returned when a regex is invalid.
type InvalidRegexError struct {
	regex string
}

func (e InvalidRegexError) Error() string {
	return "invalid regex: " + e.regex
}
