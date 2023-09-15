package daemon

import (
	"errors"

	"github.com/NethermindEth/eigenlayer/internal/profile"
)

func optionsFromProfile(profile *profile.Profile) ([]Option, error) {
	options := make([]Option, len(profile.Options))
	for i, o := range profile.Options {
		option, err := optionFromProfileOption(o)
		if err != nil {
			return nil, err
		}
		options[i] = option
	}
	return options, nil
}

func optionFromProfileOption(profileOption profile.Option) (Option, error) {
	switch profileOption.Type {
	case "str":
		return NewOptionString(profileOption), nil
	case "int":
		return NewOptionInt(profileOption)
	case "float":
		return NewOptionFloat(profileOption)
	case "bool":
		return NewOptionBool(profileOption)
	case "path_dir":
		return NewOptionPathDir(profileOption), nil
	case "path_file":
		return NewOptionPathFile(profileOption), nil
	case "uri":
		return NewOptionURI(profileOption), nil
	case "select":
		return NewOptionSelect(profileOption), nil
	case "port":
		return NewOptionPort(profileOption)
	default:
		return nil, errors.New("unknown option type: " + profileOption.Type)
	}
}
