package prompter

import (
	"github.com/AlecAivazis/survey/v2"
)

func Select(prompt string, options []string) (string, error) {
	selected := ""
	p := &survey.Select{
		Message: prompt,
		Options: options,
	}
	err := survey.AskOne(p, &selected)
	return selected, err
}

func InputString(prompt, defValue, help string, validator func(string) error) (string, error) {
	var result string
	p := &survey.Input{
		Message: prompt,
		Default: defValue,
		Help:    help,
	}
	err := survey.AskOne(p, &result, survey.WithValidator(func(ans interface{}) error {
		if err := validator(ans.(string)); err != nil {
			return err
		}
		return nil
	}))
	return result, err
}

func Confirm(prompt string) (bool, error) {
	result := false
	p := &survey.Confirm{
		Message: prompt,
	}
	err := survey.AskOne(p, &result)
	return result, err
}
