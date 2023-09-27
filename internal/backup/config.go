package backup

import (
	"io"

	"gopkg.in/yaml.v3"
)

type backupConfig struct {
	Prefix  string   `yaml:"prefix"`
	Out     string   `yaml:"out"`
	Volumes []string `yaml:"volumes"`
}

func (b *backupConfig) Save(f io.Writer) error {
	data, err := yaml.Marshal(b)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
