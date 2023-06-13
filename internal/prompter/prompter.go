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

func selectString(prompt string, options []string) (string, error) {
	selected := ""
	p := &survey.Select{
		Message: prompt,
		Options: options,
	}
	err := survey.AskOne(p, &selected)
	return selected, err
}
