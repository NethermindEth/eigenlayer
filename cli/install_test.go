package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstall_ValidateArguments(t *testing.T) {
	ts := []struct {
		name string
		args []string
		err  error
	}{
		{
			name: "no arguments",
			args: []string{},
			err:  errors.New("accepts 1 arg(s), received 0"),
		},
		{
			name: "more than one argument",
			args: []string{"arg1", "arg2"},
			err:  errors.New("accepts 1 arg(s), received 2"),
		},
		{
			name: "invalid URL",
			args: []string{"invalid-url"},
			err:  fmt.Errorf("%w: parse \"invalid-url\": invalid URI for request", ErrInvalidURL),
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			installCmd := InstallCmd(nil)

			installCmd.SetArgs(tc.args)
			err := installCmd.Execute()

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
