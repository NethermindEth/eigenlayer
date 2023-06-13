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

// TypeIndex is an enum used to represent different types of options.
type TypeIndex int

const (
	Int TypeIndex = iota + 1
	Float
	Bool
	String
	PathDir
	PathFile
	URI
	Enum
	Port
	ID
)

type Option interface {
	Name() string
	Type() TypeIndex
	Help() string
	Validate() error
	Set(string) error
	Value() string
	SetDefault()
}

// option is a struct representing a generic profile option. It includes fields for the option's name, target, and help text.
type option struct {
	name string
	help string
}

// OptionInt is a struct representing an integer option. It embeds the Option struct and adds fields for the option's value, default value, minimum value, and maximum value.
type OptionInt struct {
	option
	value    int
	Default  int
	MinValue int
	MaxValue int
}

var _ Option = (*OptionInt)(nil)

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
		Default:  defaultValue,
		MinValue: int(pkgOption.ValidateDef.MinValue),
		MaxValue: int(pkgOption.ValidateDef.MaxValue),
	}, nil
}

func (oi *OptionInt) Name() string {
	return oi.option.name
}

// Type returns the type of the option, in this case Int.
func (oi *OptionInt) Type() TypeIndex {
	return Int
}

func (oi *OptionInt) Help() string {
	return oi.option.help
}

// Validate checks if the option's value is within the allowed range. It returns an error if the value is too low or too high.
func (oi *OptionInt) Validate() error {
	if oi.value < oi.MinValue {
		return InvalidOptionValueError{
			optionName: oi.name,
			value:      oi.Value(),
			msg:        oi.Value() + " is too low",
		}
	}
	if oi.value > oi.MaxValue {
		return InvalidOptionValueError{
			optionName: oi.name,
			value:      oi.Value(),
			msg:        oi.Value() + " is too high",
		}
	}
	return nil
}

// SetDefault sets the default value for the option.
func (oi *OptionInt) SetDefault() {
	oi.value = oi.Default
}

// Set sets the value of the option. It returns an error if the value cannot be converted to an integer.
func (oi *OptionInt) Set(value string) (err error) {
	oi.value, err = strconv.Atoi(value)
	return
}

// Value returns the current value of the option as a string.
func (oi *OptionInt) Value() string {
	return strconv.Itoa(oi.value)
}

// OptionFloat is a struct representing a floating-point option. It embeds the Option struct and adds fields for the option's value, default value, minimum value, and maximum value.
type OptionFloat struct {
	option
	value    float64
	Default  float64
	MinValue float64
	MaxValue float64
}

var _ Option = (*OptionFloat)(nil)

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
		Default:  defaultValue,
		MinValue: pkgOptions.ValidateDef.MinValue,
		MaxValue: pkgOptions.ValidateDef.MaxValue,
	}, nil
}

func (of *OptionFloat) Name() string {
	return of.option.name
}

// Type returns the type of the option, in this case Float.
func (of *OptionFloat) Type() TypeIndex {
	return Float
}

func (of *OptionFloat) Help() string {
	return of.option.help
}

// Validate checks if the option's value is within the allowed range. It returns an error if the value is too low or too high.
func (of *OptionFloat) Validate() error {
	if of.value < of.MinValue {
		return InvalidOptionValueError{
			optionName: of.name,
			value:      of.Value(),
			msg:        of.Value() + " is too low",
		}
	}
	if of.value > of.MaxValue {
		return InvalidOptionValueError{
			optionName: of.name,
			value:      of.Value(),
			msg:        of.Value() + " is too high",
		}
	}
	return nil
}

// SetDefault sets the default value for the option.
func (of *OptionFloat) SetDefault() {
	of.value = of.Default
}

// Set sets the value of the option. It returns an error if the value cannot be converted to a float.
func (of *OptionFloat) Set(value string) (err error) {
	of.value, err = strconv.ParseFloat(value, 64)
	return
}

// Value returns the current value of the option as a string.
func (of *OptionFloat) Value() string {
	return strconv.FormatFloat(of.value, 'f', -1, 64)
}

// OptionBool is a struct representing a boolean option. It embeds the Option struct and adds fields for the option's value and default value.
type OptionBool struct {
	option
	value   bool
	Default bool
}

var _ Option = (*OptionBool)(nil)

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
		Default: defaultValue,
	}, nil
}

func (ob *OptionBool) Name() string {
	return ob.option.name
}

// Type returns the type of the option, in this case Bool.
func (ob *OptionBool) Type() TypeIndex {
	return Bool
}

func (ob *OptionBool) Help() string {
	return ob.option.help
}

// Validate checks if the option's value is valid. The boolean type has no validation rules, so this method always returns nil.
func (ob *OptionBool) Validate() error {
	return nil
}

// SetDefault sets the default value for the option.
func (ob *OptionBool) SetDefault() {
	ob.value = ob.Default
}

// Set sets the value of the option. It returns an error if the value cannot be converted to a boolean.
func (ob *OptionBool) Set(value string) (err error) {
	ob.value, err = strconv.ParseBool(value)
	return
}

// Value returns the current value of the option as a string.
func (ob *OptionBool) Value() string {
	return strconv.FormatBool(ob.value)
}

// OptionString is a struct representing a string option. It embeds the Option struct and adds fields for the option's value, default value, and regular expression.
type OptionString struct {
	option
	value    string
	Default  string
	Re2Regex string
}

var _ Option = (*OptionString)(nil)

func NewOptionString(pkgOption package_handler.Option) *OptionString {
	return &OptionString{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		Default:  pkgOption.Default,
		Re2Regex: pkgOption.ValidateDef.Re2Regex,
	}
}

func (os *OptionString) Name() string {
	return os.option.name
}

// Type returns the type of the option, in this case String.
func (os *OptionString) Type() TypeIndex {
	return String
}

func (os *OptionString) Help() string {
	return os.option.help
}

// Validate checks if the option's value is valid. It returns an error if the value does not match the regular expression, if one is set.
func (os *OptionString) Validate() error {
	if os.Re2Regex == "" {
		return nil
	}

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

	return nil
}

// SetDefault sets the default value for the option.
func (os *OptionString) SetDefault() {
	os.value = os.Default
}

// Set sets the value of the option.
func (os *OptionString) Set(value string) (err error) {
	os.value = value
	return
}

// Value returns the current value of the option as a string.
func (os *OptionString) Value() string {
	return os.value
}

// OptionPathDir is a struct representing a directory path option. It embeds the Option struct and adds fields for the option's value and default value.
type OptionPathDir struct {
	option
	value   string
	Default string
}

var _ Option = (*OptionPathDir)(nil)

func NewOptionPathDir(pkgOption package_handler.Option) *OptionPathDir {
	return &OptionPathDir{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		Default: pkgOption.Default,
	}
}

func (opd *OptionPathDir) Name() string {
	return opd.option.name
}

// Type returns the type of the option, in this case PathDir.
func (opd *OptionPathDir) Type() TypeIndex {
	return PathDir
}

func (opd *OptionPathDir) Help() string {
	return opd.option.help
}

// Validate checks if the option's value is valid. The directory path type has no validation rules, so this method always returns nil.
func (opd *OptionPathDir) Validate() error {
	return nil
}

// SetDefault sets the default value for the option.
func (opd *OptionPathDir) SetDefault() {
	opd.value = opd.Default
}

// Set sets the value of the option. It returns an error if the value is not a valid path.
func (opd *OptionPathDir) Set(value string) (err error) {
	if !pathRe.MatchString(value) {
		return InvalidOptionValueError{
			optionName: opd.name,
			value:      value,
			msg:        value + " is not a valid path",
		}
	}
	opd.value = value
	return
}

// Value returns the current value of the option as a string.
func (opd *OptionPathDir) Value() string {
	return opd.value
}

// OptionPathFile is a struct representing a file path option. It embeds the Option struct and adds fields for the option's value, default value, and file format.
type OptionPathFile struct {
	option
	value   string
	Default string
	Format  string
}

var _ Option = (*OptionPathFile)(nil)

func NewOptionPathFile(pkgOption package_handler.Option) *OptionPathFile {
	return &OptionPathFile{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		Default: pkgOption.Default,
		Format:  pkgOption.ValidateDef.Format,
	}
}

func (opf *OptionPathFile) Name() string {
	return opf.option.name
}

// Type returns the type of the option, in this case PathFile.
func (opf *OptionPathFile) Type() TypeIndex {
	return PathFile
}

func (opf *OptionPathFile) Help() string {
	return opf.option.help
}

// Validate checks if the option's value is valid. It returns an error if the value is not a valid path or if the file format is set and the value does not have the correct extension.
func (opf *OptionPathFile) Validate() error {
	if opf.Format != "" {
		if filepath.Ext(opf.value) != opf.Format {
			return InvalidOptionValueError{
				optionName: opf.name,
				value:      opf.Value(),
				msg:        opf.Value() + " has an invalid format. Required format is " + opf.Format,
			}
		}
	}
	return nil
}

// SetDefault sets the default value for the option.
func (opf *OptionPathFile) SetDefault() {
	opf.value = opf.Default
}

// Set sets the value of the option. It returns an error if the value is not a valid path.
func (opf *OptionPathFile) Set(value string) (err error) {
	if !pathRe.MatchString(value) {
		return InvalidOptionValueError{
			optionName: opf.name,
			value:      value,
			msg:        value + " is not a valid path",
		}
	}
	opf.value = value
	return
}

// Value returns the current value of the option as a string.
func (opf *OptionPathFile) Value() string {
	return opf.value
}

// OptionURI is a struct representing a uri option. It embeds the Option struct and adds fields for the option's value, default value, and uri scheme.
type OptionURI struct {
	option
	value     string
	Default   string
	UriScheme []string
}

var _ Option = (*OptionURI)(nil)

func NewOptionURI(pkgOption package_handler.Option) *OptionURI {
	return &OptionURI{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		Default:   pkgOption.Default,
		UriScheme: pkgOption.ValidateDef.UriScheme,
	}
}

func (ou *OptionURI) Name() string {
	return ou.option.name
}

// Type returns the type of the option, in this case URI.
func (ou *OptionURI) Type() TypeIndex {
	return URI
}

func (ou *OptionURI) Help() string {
	return ou.option.help
}

// Validate checks if the option's value is valid. It returns an error if the value is not a valid uri or if the uri scheme is set and the value does not have the correct scheme.
func (ou *OptionURI) Validate() error {
	if len(ou.UriScheme) == 0 {
		return nil
	}
	for _, scheme := range ou.UriScheme {
		if strings.HasPrefix(ou.value, scheme) {
			return nil
		}
	}
	return InvalidOptionValueError{
		optionName: ou.name,
		value:      ou.Value(),
		msg:        ou.Value() + " is not a valid uri, must be one of " + strings.Join(ou.UriScheme, ", "),
	}
}

// SetDefault sets the default value for the option.
func (ou *OptionURI) SetDefault() {
	ou.value = ou.Default
}

// Set sets the value of the option. It returns an error if the value is not a valid uri.
func (ou *OptionURI) Set(value string) (err error) {
	if _, err := url.Parse(value); err != nil {
		return InvalidOptionValueError{
			optionName: ou.name,
			value:      value,
			msg:        value + " is not a valid uri",
		}
	}

	ou.value = value
	return
}

// Value returns the current value of the option as a string.
func (ou *OptionURI) Value() string {
	return ou.value
}

// OptionEnum is a struct representing an enum option. It embeds the Option struct and adds fields for the option's possible values, and default value.
type OptionEnum struct {
	option
	values  []string
	Default string
}

var _ Option = (*OptionEnum)(nil)

func NewOptionEnum(pkgOption package_handler.Option) *OptionEnum {
	return &OptionEnum{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		Default: pkgOption.Default,
	}
}

func (oe *OptionEnum) Name() string {
	return oe.option.name
}

// Type returns the type of the option, in this case Enum.
func (oe *OptionEnum) Type() TypeIndex {
	return Enum
}

func (oe *OptionEnum) Help() string {
	return oe.option.help
}

// Validate checks if the option's value is valid. The enum type has no validation rules, so this method always returns nil.
func (oe *OptionEnum) Validate() error {
	return nil
}

// SetDefault sets the default value for the option.
func (oe *OptionEnum) SetDefault() {
	oe.values = strings.Split(oe.Default, ",")
}

// Set sets the value of the option. It returns an error if the value is not one of the possible values.
func (oe *OptionEnum) Set(value string) (err error) {
	oe.values = strings.Split(value, ",")
	return
}

// Value returns the current value of the option as a string.
func (oe *OptionEnum) Value() string {
	return strings.Join(oe.values, ",")
}

// OptionPort is a struct representing a port option. It embeds the Option struct and adds fields for the option's value and default value.
type OptionPort struct {
	option
	value   int
	Default int
}

var _ Option = (*OptionPort)(nil)

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
		Default: defaultValue,
	}, nil
}

func (op *OptionPort) Name() string {
	return op.option.name
}

// Type returns the type of the option, in this case Port.
func (op *OptionPort) Type() TypeIndex {
	return Port
}

func (op *OptionPort) Help() string {
	return op.option.help
}

// Validate checks if the option's value is valid. The port type has no validation rules, so this method always returns nil.
func (op *OptionPort) Validate() error {
	return nil
}

// SetDefault sets the default value for the option.
func (op *OptionPort) SetDefault() {
	op.value = op.Default
}

// Set sets the value of the option. It returns an error if the value is not a valid port number.
func (op *OptionPort) Set(value string) (err error) {
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
	return
}

// Value returns the current value of the option as a string.
func (op *OptionPort) Value() string {
	return strconv.Itoa(op.value)
}

// OptionID is a struct representing an id option. It embeds the Option struct and adds fields for the option's value and default value.
type OptionID struct {
	option
	value   string
	Default string
}

var _ Option = (*OptionID)(nil)

func NewOptionID(pkgOption package_handler.Option) *OptionID {
	return &OptionID{
		option: option{
			name: pkgOption.Name,
			help: pkgOption.Help,
		},
		Default: pkgOption.Default,
	}
}

func (oi *OptionID) Name() string {
	return oi.option.name
}

// Type returns the type of the option, in this case ID.
func (oi *OptionID) Type() TypeIndex {
	return ID
}

func (oi *OptionID) Help() string {
	return oi.option.help
}

// Validate checks if the option's value is valid. The id type has no validation rules, so this method always returns nil.
func (oi *OptionID) Validate() error {
	return nil
}

// SetDefault sets the default value for the option.
func (oi *OptionID) SetDefault() {
	oi.value = oi.Default
}

// Set sets the value of the option. It returns an error if the value is not a valid id.
func (oi *OptionID) Set(value string) (err error) {
	oi.value = value
	return
}

// Value returns the current value of the option as a string.
func (oi *OptionID) Value() string {
	return oi.value
}
