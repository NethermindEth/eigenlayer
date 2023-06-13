package prompter

import (
	"github.com/AlecAivazis/survey/v2"
)

func SelectProfile(options []string) (string, error) {
	return selectString("Select a profile:", options)
}

func SelectVersion(options []string) (string, error) {
	return selectString("Select a version:", options)
}

func InputString(prompt string, validator func(string) error) (string, error) {
	var result string
	p := &survey.Input{
		Message: prompt,
	}
	err := survey.AskOne(p, &result, survey.WithValidator(func(ans interface{}) error {
		if err := validator(ans.(string)); err != nil {
			return err
		}
		return nil
	}))
	return result, err
}

func selectString(prompt string, options []string) (string, error) {
	selected := ""
	p := &survey.Select{
		Message: prompt,
		Options: options,
	}
	err := survey.AskOne(p, &selected)
	return selected, err
}
