package backup

import (
	"os"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/backup/testdata"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	afs := afero.NewMemMapFs()
	outFilePath := "/config.yml"

	outFile, err := afs.OpenFile(outFilePath, os.O_CREATE|os.O_RDWR, 0o644)
	require.NoError(t, err)
	defer outFile.Close()

	config := backupConfig{
		Prefix: "volumes/instance-1",
		Volumes: []string{
			"/path/to/volume1",
			"/path/to/volume/2.txt",
		},
	}

	// Save the config to the file
	err = config.Save(outFile)
	require.NoError(t, err)

	// Read the file
	actual, err := afero.ReadFile(afs, outFilePath)
	require.NoError(t, err)

	// Assert
	expected, err := testdata.TestData.ReadFile("data/config.yml")
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
