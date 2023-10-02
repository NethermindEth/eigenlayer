package backup

import (
	"io"

	"gopkg.in/yaml.v3"
)

type backupConfig struct {
	Prefix  string   `yaml:"prefix"`
	Volumes []string `yaml:"volumes"`
}

func (b *backupConfig) Save(f io.Writer) error {
	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(b)
}
