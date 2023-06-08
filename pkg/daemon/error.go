package daemon

// InvalidOptionValueError is returned when an Option's value is invalid.
type InvalidOptionValueError struct {
	optionName string
	value      string
	msg        string
}

func (e InvalidOptionValueError) Error() string {
	return "invalid value for option " + e.optionName + ": " + e.value + ". " + e.msg
}

// InvalidRegexError is returned when a regex is invalid.
type InvalidRegexError struct {
	regex string
}

func (e InvalidRegexError) Error() string {
	return "invalid regex: " + e.regex
}
