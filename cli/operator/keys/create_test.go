package keys

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	prompterMock "github.com/NethermindEth/eigenlayer/cli/prompter/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateCmd(t *testing.T) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       []string
		err        error
		keyPath    string
		promptMock func(p *prompterMock.MockPrompter)
	}{
		{
			name: "no arguments",
			args: []string{},
			err:  fmt.Errorf("%w: accepts 1 arg, received 0", ErrInvalidNumberOfArgs),
		},
		{
			name: "more than one argument",
			args: []string{"arg1", "arg2"},
			err:  fmt.Errorf("%w: accepts 1 arg, received 2", ErrInvalidNumberOfArgs),
		},
		{
			name: "empty name argument",
			args: []string{""},
			err:  ErrEmptyKeyName,
		},
		{
			name: "keyname with whitespaces",
			args: []string{"hello world"},
			err:  ErrKeyContainsWhitespaces,
		},
		{
			name: "invalid keytype",
			args: []string{"--key-type", "invalid", "hello"},
			err:  ErrInvalidKeyType,
		},
		{
			name: "invalid password based on validation function - ecdsa",
			args: []string{"--key-type", "ecdsa", "test"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "invalid password based on validation function - bls",
			args: []string{"--key-type", "bls", "test"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "valid ecdsa key creation",
			args: []string{"--key-type", "ecdsa", "test"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.ecdsa.key.json"),
		},
		{
			name: "valid bls key creation",
			args: []string{"--key-type", "bls", "test"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: filepath.Join(homePath, OperatorKeystoreSubFolder, "/test.bls.key.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				_ = os.Remove(tt.keyPath)
			})
			controller := gomock.NewController(t)
			p := prompterMock.NewMockPrompter(controller)
			if tt.promptMock != nil {
				tt.promptMock(p)
			}

			createCmd := CreateCmd(p)
			createCmd.SetArgs(tt.args)
			err := createCmd.Execute()

			if tt.err == nil {
				assert.NoError(t, err)
				_, err := os.Stat(tt.keyPath)

				// Check if the error indicates that the file does not exist
				if os.IsNotExist(err) {
					assert.Failf(t, "file does not exist", "file %s does not exist", tt.keyPath)
				}
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
