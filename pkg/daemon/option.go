package daemon

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/NethermindEth/eigen-wiz/internal/package_handler"
)

// This pattern matches Unix-like paths, both absolute and relative
var pathRe = regexp.MustCompile(`^(/|./|../|[^/ ]([^/ ]*/)*[^/ ]*$)`)

// Option is an interface representing a generic profile option.
type Option interface {
	// Name returns the name of the option.
	Name() string
	// Help returns the help text for the option.
	Help() string
	// Set sets the value of the option. If the value is invalid, it returns an error.
	Set(string) error
	// Value returns the current value of the option.
	Value() string
	// Default returns the default value of the option.
	Default() string
}

// option is a struct representing a generic profile option. It includes fields for the option's name, target, and help text.
type option struct {
	name string
	help string
}

// OptionInt is a struct representing an integer option. It implements the Option interface.
type OptionInt struct {
	option
	value        int
	defaultValue int
	MinValue     int
	MaxValue     int
}

// NewOptionInt creates a new OptionInt from a package_handler.Option.
func NewOptionInt(pkgOption package_handler.Option) (*OptionInt, error) {
	defaultValue, err := strconv.Atoi(pkgOption.Default)
	if err != nil {
		return nil, err
	}
	return &OptionInt{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defaultValue: defaultValue,
		MinValue:     int(pkgOption.ValidateDef.MinValue),
		MaxValue:     int(pkgOption.ValidateDef.MaxValue),
	}, nil
}

var _ Option = (*OptionInt)(nil)

func (oi *OptionInt) Name() string {
	return oi.option.name
}

func (oi *OptionInt) Help() string {
	return oi.option.help
}

func (oi *OptionInt) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if oi.MinValue != 0 && v < oi.MinValue {
		return InvalidOptionValueError{
			optionName: oi.name,
			value:      oi.Value(),
			msg:        oi.Value() + " is too low",
		}
	}
	if oi.MaxValue != 0 && v > oi.MaxValue {
		return InvalidOptionValueError{
			optionName: oi.name,
			value:      oi.Value(),
			msg:        oi.Value() + " is too high",
		}
	}
	oi.value = v
	return nil
}

func (oi *OptionInt) Value() string {
	return strconv.Itoa(oi.value)
}

func (oi *OptionInt) Default() string {
	return strconv.Itoa(oi.defaultValue)
}

// OptionFloat is a struct representing a floating-point option. It implements the Option interface.
type OptionFloat struct {
	option
	value    float64
	defValue float64
	MinValue float64
	MaxValue float64
}

func NewOptionFloat(pkgOptions package_handler.Option) (*OptionFloat, error) {
	defaultValue, err := strconv.ParseFloat(pkgOptions.Default, 64)
	if err != nil {
		return nil, err
	}
	return &OptionFloat{
		option: option{
			name: pkgOptions.Name,
			help: pkgOptions.Help,
		},
		defValue: defaultValue,
		MinValue: pkgOptions.ValidateDef.MinValue,
		MaxValue: pkgOptions.ValidateDef.MaxValue,
	}, nil
}

var _ Option = (*OptionFloat)(nil)

func (of *OptionFloat) Name() string {
	return of.option.name
}

func (of *OptionFloat) Help() string {
	return of.option.help
}

func (of *OptionFloat) Set(value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	if v < of.MinValue {
		return InvalidOptionValueError{
			optionName: of.name,
			value:      of.Value(),
			msg:        of.Value() + " is too low",
		}
	}
	if v > of.MaxValue {
		return InvalidOptionValueError{
			optionName: of.name,
			value:      of.Value(),
			msg:        of.Value() + " is too high",
		}
	}
	of.value = v
	return nil
}

func (of *OptionFloat) Value() string {
	return strconv.FormatFloat(of.value, 'f', -1, 64)
}

func (of *OptionFloat) Default() string {
	return strconv.FormatFloat(of.defValue, 'f', -1, 64)
}

// OptionBool is a struct representing a boolean option. It implements the Option interface.
type OptionBool struct {
	option
	value    bool
	defValue bool
}

func NewOptionBool(pkgOption package_handler.Option) (*OptionBool, error) {
	defaultValue, err := strconv.ParseBool(pkgOption.Default)
	if err != nil {
		return nil, err
	}
	return &OptionBool{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: defaultValue,
	}, nil
}

var _ Option = (*OptionBool)(nil)

func (ob *OptionBool) Name() string {
	return ob.option.name
}

func (ob *OptionBool) Help() string {
	return ob.option.help
}

func (ob *OptionBool) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	ob.value = v
	return nil
}

func (ob *OptionBool) Value() string {
	return strconv.FormatBool(ob.value)
}

func (ob *OptionBool) Default() string {
	return strconv.FormatBool(ob.defValue)
}

// OptionString is a struct representing a string option. It implements the Option interface.
type OptionString struct {
	option
	value    string
	defValue string
	Re2Regex string
}

func NewOptionString(pkgOption package_handler.Option) *OptionString {
	return &OptionString{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: pkgOption.Default,
		Re2Regex: pkgOption.ValidateDef.Re2Regex,
	}
}

var _ Option = (*OptionString)(nil)

func (os *OptionString) Name() string {
	return os.option.name
}

func (os *OptionString) Help() string {
	return os.option.help
}

func (os *OptionString) Set(value string) error {
	if os.Re2Regex != "" {
		regex, err := regexp.Compile(os.Re2Regex)
		if err != nil {
			return InvalidRegexError{
				regex: os.Re2Regex,
			}
		}
		if !regex.MatchString(os.value) {
			return InvalidOptionValueError{
				optionName: os.name,
				value:      os.Value(),
				msg:        os.Value() + " does not match regex",
			}
		}
	}
	os.value = value
	return nil
}

func (os *OptionString) Value() string {
	return os.value
}

func (os *OptionString) Default() string {
	return os.defValue
}

// OptionPathDir is a struct representing a directory path option. It implements the Option interface.
type OptionPathDir struct {
	option
	value    string
	defValue string
}

func NewOptionPathDir(pkgOption package_handler.Option) *OptionPathDir {
	return &OptionPathDir{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: pkgOption.Default,
	}
}

var _ Option = (*OptionPathDir)(nil)

func (opd *OptionPathDir) Name() string {
	return opd.option.name
}

func (opd *OptionPathDir) Help() string {
	return opd.option.help
}

func (opd *OptionPathDir) Set(value string) error {
	if !pathRe.MatchString(value) {
		return InvalidOptionValueError{
			optionName: opd.name,
			value:      value,
			msg:        value + " is not a valid path",
		}
	}
	opd.value = value
	return nil
}

func (opd *OptionPathDir) Value() string {
	return opd.value
}

func (opd *OptionPathDir) Default() string {
	return opd.defValue
}

// OptionPathFile is a struct representing a file path option. It implements the Option interface.
type OptionPathFile struct {
	option
	value    string
	defValue string
	Format   string
}

func NewOptionPathFile(pkgOption package_handler.Option) *OptionPathFile {
	return &OptionPathFile{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: pkgOption.Default,
		Format:   pkgOption.ValidateDef.Format,
	}
}

var _ Option = (*OptionPathFile)(nil)

func (opf *OptionPathFile) Name() string {
	return opf.option.name
}

func (opf *OptionPathFile) Help() string {
	return opf.option.help
}

func (opf *OptionPathFile) Set(value string) error {
	if opf.Format != "" {
		if filepath.Ext(opf.value) != opf.Format {
			return InvalidOptionValueError{
				optionName: opf.name,
				value:      opf.Value(),
				msg:        opf.Value() + " has an invalid format. Required format is " + opf.Format,
			}
		}
	}
	if !pathRe.MatchString(value) {
		return InvalidOptionValueError{
			optionName: opf.name,
			value:      value,
			msg:        value + " is not a valid path",
		}
	}
	opf.value = value
	return nil
}

func (opf *OptionPathFile) Value() string {
	return opf.value
}

func (opf *OptionPathFile) Default() string {
	return opf.defValue
}

// OptionURI is a struct representing a uri option. It implements the Option interface.
type OptionURI struct {
	option
	value     string
	defValue  string
	UriScheme []string
}

func NewOptionURI(pkgOption package_handler.Option) *OptionURI {
	return &OptionURI{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue:  pkgOption.Default,
		UriScheme: pkgOption.ValidateDef.UriScheme,
	}
}

var _ Option = (*OptionURI)(nil)

func (ou *OptionURI) Name() string {
	return ou.option.name
}

func (ou *OptionURI) Help() string {
	return ou.option.help
}

func (ou *OptionURI) Set(value string) error {
	if _, err := url.Parse(value); err != nil {
		return InvalidOptionValueError{
			optionName: ou.name,
			value:      value,
			msg:        value + " is not a valid uri",
		}
	}
	if len(ou.UriScheme) != 0 {
		for _, scheme := range ou.UriScheme {
			if strings.HasPrefix(ou.value, scheme) {
				ou.value = value
				return nil
			}
		}
		return InvalidOptionValueError{
			optionName: ou.name,
			value:      ou.Value(),
			msg:        ou.Value() + " is not a valid uri, must be one of " + strings.Join(ou.UriScheme, ", "),
		}
	} else {
		ou.value = value
		return nil
	}
}

func (ou *OptionURI) Value() string {
	return ou.value
}

func (ou *OptionURI) Default() string {
	return ou.defValue
}

// OptionEnum is a struct representing an enum option. It implements the Option interface.
type OptionEnum struct {
	option
	value    string
	defValue string
	Options  []string
}

func NewOptionEnum(pkgOption package_handler.Option) *OptionEnum {
	return &OptionEnum{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: pkgOption.Default,
		// TODO: add Options
	}
}

var _ Option = (*OptionEnum)(nil)

func (oe *OptionEnum) Name() string {
	return oe.option.name
}

func (oe *OptionEnum) Help() string {
	return oe.option.help
}

func (oe *OptionEnum) Set(value string) error {
	// TODO: add validation. Check if value is one of the possible values
	oe.value = value
	return nil
}

func (oe *OptionEnum) Value() string {
	return oe.value
}

func (oe *OptionEnum) Default() string {
	return oe.defValue
}

// OptionPort is a struct representing a port option. It implements the Option interface.
type OptionPort struct {
	option
	value    int
	defValue int
}

func NewOptionPort(pkgOption package_handler.Option) (*OptionPort, error) {
	defaultValue, err := strconv.Atoi(pkgOption.Default)
	if err != nil {
		return nil, err
	}
	return &OptionPort{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: defaultValue,
	}, nil
}

var _ Option = (*OptionPort)(nil)

func (op *OptionPort) Name() string {
	return op.option.name
}

func (op *OptionPort) Help() string {
	return op.option.help
}

func (op *OptionPort) Set(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return InvalidOptionValueError{
			optionName: op.name,
			value:      value,
			msg:        value + " is not a valid port number",
		}
	}

	if op.value < 0 || op.value > 65535 {
		return InvalidOptionValueError{
			optionName: op.name,
			value:      value,
			msg:        value + " is not a valid port. Port must be between 0 and 65535",
		}
	}

	op.value = port
	return nil
}

func (op *OptionPort) Value() string {
	return strconv.Itoa(op.value)
}

func (op *OptionPort) Default() string {
	return strconv.Itoa(op.defValue)
}

// OptionID is a struct representing an id option. It implements the Option interface.
type OptionID struct {
	option
	value    string
	defValue string
}

func NewOptionID(pkgOption package_handler.Option) *OptionID {
	return &OptionID{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		defValue: pkgOption.Default,
	}
}

var _ Option = (*OptionID)(nil)

func (oi *OptionID) Name() string {
	return oi.option.name
}

func (oi *OptionID) Help() string {
	return oi.option.help
}

func (oi *OptionID) Set(value string) error {
	oi.value = value
	return nil
}

func (oi *OptionID) Value() string {
	return oi.value
}

func (oi *OptionID) Default() string {
	return oi.defValue
}
