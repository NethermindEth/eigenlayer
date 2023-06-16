package daemon

import (
	"fmt"
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
	// Target returns the target docker-compose environment variable for the option.
	Target() string
}

// option is a struct representing a generic profile option. It includes fields for the option's name, target, and help text.
type option struct {
	name   string
	target string
	help   string
}

// OptionInt is a struct representing an integer option. It implements the Option interface.
type OptionInt struct {
	option
	value        int
	defaultValue int
	validate     bool
	MinValue     int
	MaxValue     int
}

// NewOptionInt creates a new OptionInt from a package_handler.Option.
func NewOptionInt(pkgOption package_handler.Option) (*OptionInt, error) {
	defaultValue, err := strconv.Atoi(pkgOption.Default)
	if err != nil {
		return nil, err
	}
	o := &OptionInt{
		option: option{
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defaultValue: defaultValue,
		validate:     pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.MinValue = int(pkgOption.ValidateDef.MinValue)
		o.MaxValue = int(pkgOption.ValidateDef.MaxValue)
	}
	return o, nil
}

var _ Option = (*OptionInt)(nil)

func (oi *OptionInt) Name() string {
	return oi.option.name
}

func (oi *OptionInt) Help() string {
	if oi.validate {
		return fmt.Sprintf("%s (min: %d, max: %d)", oi.option.help, oi.MinValue, oi.MaxValue)
	}
	return oi.option.help
}

func (oi *OptionInt) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return InvalidOptionValueError{
			optionName: oi.name,
			value:      value,
			msg:        "it is not an integer",
		}
	}
	if oi.validate {
		if v < oi.MinValue {
			return InvalidOptionValueError{
				optionName: oi.name,
				value:      value,
				msg:        fmt.Sprintf("should be greater than %d", oi.MinValue),
			}
		}
		if v > oi.MaxValue {
			return InvalidOptionValueError{
				optionName: oi.name,
				value:      value,
				msg:        fmt.Sprintf("should be less than %d", oi.MaxValue),
			}
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

func (oi *OptionInt) Target() string {
	return oi.target
}

// OptionFloat is a struct representing a floating-point option. It implements the Option interface.
type OptionFloat struct {
	option
	value    float64
	defValue float64
	validate bool
	MinValue float64
	MaxValue float64
}

func NewOptionFloat(pkgOption package_handler.Option) (*OptionFloat, error) {
	defaultValue, err := strconv.ParseFloat(pkgOption.Default, 64)
	if err != nil {
		return nil, err
	}
	o := &OptionFloat{
		option: option{
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defValue: defaultValue,
		validate: pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.MinValue = pkgOption.ValidateDef.MinValue
		o.MaxValue = pkgOption.ValidateDef.MaxValue
	}
	return o, nil
}

var _ Option = (*OptionFloat)(nil)

func (of *OptionFloat) Name() string {
	return of.option.name
}

func (of *OptionFloat) Help() string {
	if of.validate {
		return fmt.Sprintf("%s (min: %f, max: %f)", of.option.help, of.MinValue, of.MaxValue)
	}
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
			value:      value,
			msg:        value + " is too low",
		}
	}
	if v > of.MaxValue {
		return InvalidOptionValueError{
			optionName: of.name,
			value:      value,
			msg:        value + " is too high",
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

func (of *OptionFloat) Target() string {
	return of.target
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
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
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

func (ob *OptionBool) Target() string {
	return ob.target
}

// OptionString is a struct representing a string option. It implements the Option interface.
type OptionString struct {
	option
	value    string
	defValue string
	validate bool
	Re2Regex string
}

func NewOptionString(pkgOption package_handler.Option) *OptionString {
	o := &OptionString{
		option: option{
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defValue: pkgOption.Default,
		validate: pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.Re2Regex = pkgOption.ValidateDef.Re2Regex
	}
	return o
}

var _ Option = (*OptionString)(nil)

func (os *OptionString) Name() string {
	return os.option.name
}

func (os *OptionString) Help() string {
	if os.validate {
		return fmt.Sprintf("%s (regex: %s)", os.option.help, os.Re2Regex)
	}
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
		if !regex.MatchString(value) {
			return InvalidOptionValueError{
				optionName: os.name,
				value:      value,
				msg:        "does not match with regex: " + os.Re2Regex,
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

func (os *OptionString) Target() string {
	return os.target
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
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
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
			msg:        "it is not a valid path",
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

func (opd *OptionPathDir) Target() string {
	return opd.target
}

// OptionPathFile is a struct representing a file path option. It implements the Option interface.
type OptionPathFile struct {
	option
	value    string
	validate bool
	defValue string
	Format   string
}

func NewOptionPathFile(pkgOption package_handler.Option) *OptionPathFile {
	o := &OptionPathFile{
		option: option{
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defValue: pkgOption.Default,
		validate: pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.Format = pkgOption.ValidateDef.Format
	}
	return o
}

var _ Option = (*OptionPathFile)(nil)

func (opf *OptionPathFile) Name() string {
	return opf.option.name
}

func (opf *OptionPathFile) Help() string {
	return opf.option.help
}

func (opf *OptionPathFile) Set(value string) error {
	if opf.validate {
		if filepath.Ext(value) != "."+opf.Format {
			return InvalidOptionValueError{
				optionName: opf.name,
				value:      value,
				msg:        "it has an invalid format. Required format is " + opf.Format,
			}
		}
	}
	if !pathRe.MatchString(value) {
		return InvalidOptionValueError{
			optionName: opf.name,
			value:      value,
			msg:        "is not a valid path",
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

func (opf *OptionPathFile) Target() string {
	return opf.target
}

// OptionURI is a struct representing a uri option. It implements the Option interface.
type OptionURI struct {
	option
	value     string
	validate  bool
	defValue  string
	UriScheme []string
}

func NewOptionURI(pkgOption package_handler.Option) *OptionURI {
	o := OptionURI{
		option: option{
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defValue: pkgOption.Default,
		validate: pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.UriScheme = pkgOption.ValidateDef.UriScheme
	}
	return &o
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
			msg:        "it is not a valid uri",
		}
	}
	if ou.validate && len(ou.UriScheme) != 0 {
		for _, scheme := range ou.UriScheme {
			if strings.HasPrefix(value, scheme) {
				ou.value = value
				return nil
			}
		}
		return InvalidOptionValueError{
			optionName: ou.name,
			value:      value,
			msg:        "it is not a valid uri, must be one of [" + strings.Join(ou.UriScheme, ", ") + "]",
		}
	}
	ou.value = value
	return nil
}

func (ou *OptionURI) Value() string {
	return ou.value
}

func (ou *OptionURI) Default() string {
	return ou.defValue
}

func (ou *OptionURI) Target() string {
	return ou.target
}

// OptionSelect is a struct representing an enum option. It implements the Option interface.
type OptionSelect struct {
	option
	value    string
	defValue string
	validate bool
	Options  []string
}

func NewOptionSelect(pkgOption package_handler.Option) *OptionSelect {
	o := OptionSelect{
		option: option{
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defValue: pkgOption.Default,
		validate: pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.Options = pkgOption.ValidateDef.Options
	}
	return &o
}

var _ Option = (*OptionSelect)(nil)

func (oe *OptionSelect) Name() string {
	return oe.option.name
}

func (oe *OptionSelect) Help() string {
	return fmt.Sprintf("%s (options: %s)", oe.option.help, strings.Join(oe.Options, ", "))
}

func (oe *OptionSelect) Set(value string) error {
	for _, option := range oe.Options {
		if option == value {
			oe.value = value
			return nil
		}
	}
	return InvalidOptionValueError{
		optionName: oe.name,
		value:      value,
		msg:        "must be one of " + strings.Join(oe.Options, ", "),
	}
}

func (oe *OptionSelect) Value() string {
	return oe.value
}

func (oe *OptionSelect) Default() string {
	return oe.defValue
}

func (oe *OptionSelect) Target() string {
	return oe.target
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
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
		},
		defValue: defaultValue,
	}, nil
}

var _ Option = (*OptionPort)(nil)

func (op *OptionPort) Name() string {
	return op.option.name
}

func (op *OptionPort) Help() string {
	return fmt.Sprintf("%s (min: 0, max: 65535)", op.option.help)
}

func (op *OptionPort) Set(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return InvalidOptionValueError{
			optionName: op.name,
			value:      value,
			msg:        "it is not a valid port number",
		}
	}

	if op.value < 0 || op.value > 65535 {
		return InvalidOptionValueError{
			optionName: op.name,
			value:      value,
			msg:        "it is not a valid port. Port must be between 0 and 65535",
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

func (op *OptionPort) Target() string {
	return op.target
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
			name:   pkgOption.Name,
			target: pkgOption.Target,
			help:   pkgOption.Help,
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

func (oi *OptionID) Target() string {
	return oi.target
}
