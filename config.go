package git_backup

import (
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

type Config struct {
	Github []*GithubConfig `yaml:"github"`
}

func (c *Config) setDefaults() {
	if c.Github != nil {
		for _, config := range c.Github {
			config.setDefaults()
		}
	}
}


func LoadFile(path string) (out Config, err error) {
	handle, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		err = handle.Close()
	}()
	out, err = LoadReader(handle)
	return
}

func LoadReader(reader io.Reader) (out Config, err error) {
	dec := yaml.NewDecoder(reader)
	dec.SetStrict(true)
	err = dec.Decode(&out)
	out.setDefaults()
	return
}
