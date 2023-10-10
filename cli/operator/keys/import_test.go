package keys

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"

	prompterMock "github.com/NethermindEth/eigenlayer/cli/prompter/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestImportCmd(t *testing.T) {
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
			err:  fmt.Errorf("%w: accepts 2 arg, received 0", ErrInvalidNumberOfArgs),
		},
		{
			name: "one argument",
			args: []string{"arg1"},
			err:  fmt.Errorf("%w: accepts 2 arg, received 1", ErrInvalidNumberOfArgs),
		},

		{
			name: "more than two argument",
			args: []string{"arg1", "arg2", "arg3"},
			err:  fmt.Errorf("%w: accepts 2 arg, received 3", ErrInvalidNumberOfArgs),
		},
		{
			name: "empty key name argument",
			args: []string{"", ""},
			err:  ErrEmptyKeyName,
		},
		{
			name: "keyname with whitespaces",
			args: []string{"hello world", ""},
			err:  ErrKeyContainsWhitespaces,
		},
		{
			name: "empty private key argument",
			args: []string{"hello", ""},
			err:  ErrEmptyPrivateKey,
		},
		{
			name: "keyname with whitespaces",
			args: []string{"hello", "hello world"},
			err:  ErrPrivateKeyContainsWhitespaces,
		},
		{
			name: "invalid keytype",
			args: []string{"--key-type", "invalid", "hello", "privkey"},
			err:  ErrInvalidKeyType,
		},
		{
			name: "invalid password based on validation function - ecdsa",
			args: []string{"--key-type", "ecdsa", "test", "6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "invalid password based on validation function - bls",
			args: []string{"--key-type", "bls", "test", "123"},
			err:  ErrInvalidPassword,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrInvalidPassword)
			},
		},
		{
			name: "valid ecdsa key import",
			args: []string{"--key-type", "ecdsa", "test", "6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: OperatorKeyFolder + "/test.ecdsa.key.json",
		},
		{
			name: "valid ecdsa key import with 0x prefix",
			args: []string{"--key-type", "ecdsa", "test", "0x6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: OperatorKeyFolder + "/test.ecdsa.key.json",
		},
		{
			name: "valid bls key import",
			args: []string{"--key-type", "bls", "test", "20030410000080487431431153104351076122223465926814327806350179952713280726583"},
			err:  nil,
			promptMock: func(p *prompterMock.MockPrompter) {
				p.EXPECT().InputHiddenString(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
			keyPath: OperatorKeyFolder + "/test.bls.key.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				_ = os.RemoveAll(OperatorKeyFolder)
			})
			controller := gomock.NewController(t)
			p := prompterMock.NewMockPrompter(controller)
			if tt.promptMock != nil {
				tt.promptMock(p)
			}

			importCmd := ImportCmd(p)
			importCmd.SetArgs(tt.args)
			err := importCmd.Execute()

			if tt.err == nil {
				assert.NoError(t, err)
				_, err := os.Stat(tt.keyPath)

				// Check if the error indicates that the file does not exist
				if os.IsNotExist(err) {
					assert.Failf(t, "file does not exist", "file %s does not exist", tt.keyPath)
				}

				if tt.args[1] == KeyTypeECDSA {
					key, err := GetECDSAPrivateKey(tt.keyPath, "")
					assert.NoError(t, err)
					assert.Equal(t, strings.Trim(tt.args[3], "0x"), hex.EncodeToString(key.D.Bytes()))
				} else if tt.args[1] == KeyTypeBLS {
					key, err := bls.ReadPrivateKeyFromFile(tt.keyPath, "")
					assert.NoError(t, err)
					assert.Equal(t, tt.args[3], key.PrivKey.String())
				}
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
