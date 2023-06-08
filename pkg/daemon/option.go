package daemon

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

// Option is a struct representing a generic profile option. It includes fields for the option's name, target, and help text.
type Option struct {
	Name   string
	Target string
	Help   string
}

// Validate checks if the option is valid according to the validation rules of the EigenLayer specification. It returns an error if the option is invalid.
func (o *Option) Validate() error {
	return nil
}

// SetDefault sets the default value for the option.
func (o *Option) SetDefault() {}

// Set sets the value of the option. It returns an error if the value cannot be set.
func (o *Option) Set(value string) error {
	return nil
}

// Value returns the current value of the option as a string.
func (o *Option) Value() string {
	return ""
}

// OptionInt is a struct representing an integer option. It embeds the Option struct and adds fields for the option's value, default value, minimum value, and maximum value.
type OptionInt struct {
	Option
	value    int
	Default  int
	MinValue int
	MaxValue int
}

// Type returns the type of the option, in this case Int.
func (oi *OptionInt) Type() TypeIndex {
	return Int
}

// Validate checks if the option's value is within the allowed range. It returns an error if the value is too low or too high.
func (oi *OptionInt) Validate() error {
	if oi.value < oi.MinValue {
		return InvalidOptionValueError{
			optionName: oi.Name,
			value:      oi.Value(),
			msg:        oi.Value() + " is too low",
		}
	}
	if oi.value > oi.MaxValue {
		return InvalidOptionValueError{
			optionName: oi.Name,
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
	Option
	value    float64
	Default  float64
	MinValue float64
	MaxValue float64
}

// Type returns the type of the option, in this case Float.
func (of *OptionFloat) Type() TypeIndex {
	return Float
}

// Validate checks if the option's value is within the allowed range. It returns an error if the value is too low or too high.
func (of *OptionFloat) Validate() error {
	if of.value < of.MinValue {
		return InvalidOptionValueError{
			optionName: of.Name,
			value:      of.Value(),
			msg:        of.Value() + " is too low",
		}
	}
	if of.value > of.MaxValue {
		return InvalidOptionValueError{
			optionName: of.Name,
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
	Option
	value   bool
	Default bool
}

// Type returns the type of the option, in this case Bool.
func (ob *OptionBool) Type() TypeIndex {
	return Bool
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
	Option
	value    string
	Default  string
	Re2Regex string
}

// Type returns the type of the option, in this case String.
func (os *OptionString) Type() TypeIndex {
	return String
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
			optionName: os.Name,
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
	Option
	value   string
	Default string
}

// Type returns the type of the option, in this case PathDir.
func (opd *OptionPathDir) Type() TypeIndex {
	return PathDir
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
			optionName: opd.Name,
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
	Option
	value   string
	Default string
	Format  string
}

// Type returns the type of the option, in this case PathFile.
func (opf *OptionPathFile) Type() TypeIndex {
	return PathFile
}

// Validate checks if the option's value is valid. It returns an error if the value is not a valid path or if the file format is set and the value does not have the correct extension.
func (opf *OptionPathFile) Validate() error {
	if opf.Format != "" {
		if filepath.Ext(opf.value) != opf.Format {
			return InvalidOptionValueError{
				optionName: opf.Name,
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
			optionName: opf.Name,
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
	Option
	value     string
	Default   string
	UriScheme string
}

// Type returns the type of the option, in this case URI.
func (ou *OptionURI) Type() TypeIndex {
	return URI
}

// Validate checks if the option's value is valid. It returns an error if the value is not a valid uri or if the uri scheme is set and the value does not have the correct scheme.
func (ou *OptionURI) Validate() error {
	if ou.UriScheme != "" {
		if !strings.HasPrefix(ou.value, ou.UriScheme) {
			return InvalidOptionValueError{
				optionName: ou.Name,
				value:      ou.Value(),
				msg:        ou.Value() + " is not a valid " + ou.UriScheme + " uri",
			}
		}
	}
	return nil
}

// SetDefault sets the default value for the option.
func (ou *OptionURI) SetDefault() {
	ou.value = ou.Default
}

// Set sets the value of the option. It returns an error if the value is not a valid uri.
func (ou *OptionURI) Set(value string) (err error) {
	if _, err := url.Parse(value); err != nil {
		return InvalidOptionValueError{
			optionName: ou.Name,
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
	Option
	values  []string
	Default string
}

// Type returns the type of the option, in this case Enum.
func (oe *OptionEnum) Type() TypeIndex {
	return Enum
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
	Option
	value   int
	Default int
}

// Type returns the type of the option, in this case Port.
func (op *OptionPort) Type() TypeIndex {
	return Port
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
			optionName: op.Name,
			value:      value,
			msg:        value + " is not a valid port number",
		}
	}

	if op.value < 0 || op.value > 65535 {
		return InvalidOptionValueError{
			optionName: op.Name,
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
	Option
	value   string
	Default string
}

// Type returns the type of the option, in this case ID.
func (oi *OptionID) Type() TypeIndex {
	return ID
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
