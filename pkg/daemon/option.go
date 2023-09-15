package daemon

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/NethermindEth/eigenlayer/internal/profile"
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
	Value() (string, error)
	// Default returns the default value of the option.
	Default() string
	// Target returns the target docker-compose environment variable for the option.
	Target() string
	IsSet() bool
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
	value    *int
	defValue int
	validate bool
	MinValue int
	MaxValue int
}

// NewOptionInt creates a new OptionInt from a profile.Option.
func NewOptionInt(pkgOption profile.Option) (*OptionInt, error) {
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
		defValue: defaultValue,
		validate: pkgOption.ValidateDef != nil,
	}
	if o.validate {
		o.MinValue = int(*pkgOption.ValidateDef.MinValue)
		o.MaxValue = int(*pkgOption.ValidateDef.MaxValue)
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
	oi.value = &v
	return nil
}

func (oi *OptionInt) Value() (string, error) {
	if oi.IsSet() {
		return strconv.Itoa(*oi.value), nil
	}
	return "", ErrOptionNotSet
}

func (oi *OptionInt) Default() string {
	return strconv.Itoa(oi.defValue)
}

func (oi *OptionInt) Target() string {
	return oi.target
}

func (oi *OptionInt) IsSet() bool {
	return oi.value != nil
}

// OptionFloat is a struct representing a floating-point option. It implements the Option interface.
type OptionFloat struct {
	option
	value    *float64
	defValue float64
	validate bool
	MinValue float64
	MaxValue float64
}

func NewOptionFloat(pkgOption profile.Option) (*OptionFloat, error) {
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
		o.MinValue = *pkgOption.ValidateDef.MinValue
		o.MaxValue = *pkgOption.ValidateDef.MaxValue
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
	of.value = &v
	return nil
}

func (of *OptionFloat) Value() (string, error) {
	if of.IsSet() {
		return strconv.FormatFloat(*of.value, 'f', -1, 64), nil
	}
	return "", ErrOptionNotSet
}

func (of *OptionFloat) Default() string {
	return strconv.FormatFloat(of.defValue, 'f', -1, 64)
}

func (of *OptionFloat) Target() string {
	return of.target
}

func (of *OptionFloat) IsSet() bool {
	return of.value != nil
}

// OptionBool is a struct representing a boolean option. It implements the Option interface.
type OptionBool struct {
	option
	value    *bool
	defValue bool
}

func NewOptionBool(pkgOption profile.Option) (*OptionBool, error) {
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
	ob.value = &v
	return nil
}

func (ob *OptionBool) Value() (string, error) {
	if ob.IsSet() {
		return strconv.FormatBool(*ob.value), nil
	}
	return "", ErrOptionNotSet
}

func (ob *OptionBool) Default() string {
	return strconv.FormatBool(ob.defValue)
}

func (ob *OptionBool) Target() string {
	return ob.target
}

func (ob *OptionBool) IsSet() bool {
	return ob.value != nil
}

// OptionString is a struct representing a string option. It implements the Option interface.
type OptionString struct {
	option
	value    *string
	defValue string
	validate bool
	Re2Regex string
}

func NewOptionString(pkgOption profile.Option) *OptionString {
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
	os.value = &value
	return nil
}

func (os *OptionString) Value() (string, error) {
	if os.IsSet() {
		return *os.value, nil
	}
	return "", ErrOptionNotSet
}

func (os *OptionString) Default() string {
	return os.defValue
}

func (os *OptionString) Target() string {
	return os.target
}

func (os *OptionString) IsSet() bool {
	return os.value != nil
}

// OptionPathDir is a struct representing a directory path option. It implements the Option interface.
type OptionPathDir struct {
	option
	value    *string
	defValue string
}

func NewOptionPathDir(pkgOption profile.Option) *OptionPathDir {
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
	opd.value = &value
	return nil
}

func (opd *OptionPathDir) Value() (string, error) {
	if opd.IsSet() {
		return *opd.value, nil
	}
	return "", ErrOptionNotSet
}

func (opd *OptionPathDir) Default() string {
	return opd.defValue
}

func (opd *OptionPathDir) Target() string {
	return opd.target
}

func (opd *OptionPathDir) IsSet() bool {
	return opd.value != nil
}

// OptionPathFile is a struct representing a file path option. It implements the Option interface.
type OptionPathFile struct {
	option
	value    *string
	validate bool
	defValue string
	Format   string
}

func NewOptionPathFile(pkgOption profile.Option) *OptionPathFile {
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
	if opf.validate {
		return fmt.Sprintf("%s (format: %s)", opf.option.help, opf.Format)
	}
	return opf.option.help
}

func (opf *OptionPathFile) Set(value string) error {
	if opf.validate {
		if filepath.Ext(value) != opf.Format {
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
	opf.value = &value
	return nil
}

func (opf *OptionPathFile) Value() (string, error) {
	if opf.IsSet() {
		return *opf.value, nil
	}
	return "", ErrOptionNotSet
}

func (opf *OptionPathFile) Default() string {
	return opf.defValue
}

func (opf *OptionPathFile) Target() string {
	return opf.target
}

func (opf *OptionPathFile) IsSet() bool {
	return opf.value != nil
}

// OptionURI is a struct representing a uri option. It implements the Option interface.
type OptionURI struct {
	option
	value     *string
	validate  bool
	defValue  string
	UriScheme []string
}

func NewOptionURI(pkgOption profile.Option) *OptionURI {
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
	if ou.validate {
		return fmt.Sprintf("%s (uri scheme: %s)", ou.option.help, strings.Join(ou.UriScheme, ", "))
	}
	return ou.option.help
}

func (ou *OptionURI) Set(value string) error {
	url, err := url.Parse(value)
	if err != nil {
		return InvalidOptionValueError{
			optionName: ou.name,
			value:      value,
			msg:        "it is not a valid uri",
		}
	}
	if ou.validate && len(ou.UriScheme) != 0 {
		for _, scheme := range ou.UriScheme {
			if url.Scheme == scheme {
				ou.value = &value
				return nil
			}
		}
		return InvalidOptionValueError{
			optionName: ou.name,
			value:      value,
			msg:        "it is not a valid uri, must be one of [" + strings.Join(ou.UriScheme, ", ") + "]",
		}
	}
	ou.value = &value
	return nil
}

func (ou *OptionURI) Value() (string, error) {
	if ou.IsSet() {
		return *ou.value, nil
	}
	return "", ErrOptionNotSet
}

func (ou *OptionURI) Default() string {
	return ou.defValue
}

func (ou *OptionURI) Target() string {
	return ou.target
}

func (ou *OptionURI) IsSet() bool {
	return ou.value != nil
}

// OptionSelect is a struct representing an enum option. It implements the Option interface.
type OptionSelect struct {
	option
	value    *string
	defValue string
	validate bool
	Options  []string
}

func NewOptionSelect(pkgOption profile.Option) *OptionSelect {
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
			oe.value = &value
			return nil
		}
	}
	return InvalidOptionValueError{
		optionName: oe.name,
		value:      value,
		msg:        "must be one of " + strings.Join(oe.Options, ", "),
	}
}

func (oe *OptionSelect) Value() (string, error) {
	if oe.IsSet() {
		return *oe.value, nil
	}
	return "", ErrOptionNotSet
}

func (oe *OptionSelect) Default() string {
	return oe.defValue
}

func (oe *OptionSelect) Target() string {
	return oe.target
}

func (oe *OptionSelect) IsSet() bool {
	return oe.value != nil
}

// OptionPort is a struct representing a port option. It implements the Option interface.
type OptionPort struct {
	option
	value    *int
	defValue int
}

func NewOptionPort(pkgOption profile.Option) (*OptionPort, error) {
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

	if port < 0 || port > 65535 {
		return InvalidOptionValueError{
			optionName: op.name,
			value:      value,
			msg:        "it is not a valid port. Port must be between 0 and 65535",
		}
	}

	op.value = &port
	return nil
}

func (op *OptionPort) Value() (string, error) {
	if op.IsSet() {
		return strconv.Itoa(*op.value), nil
	}
	return "", ErrOptionNotSet
}

func (op *OptionPort) Default() string {
	return strconv.Itoa(op.defValue)
}

func (op *OptionPort) Target() string {
	return op.target
}

func (op *OptionPort) IsSet() bool {
	return op.value != nil
}
