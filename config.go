package git_backup

import (
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

type Config struct {
	Github []*GithubConfig `yaml:"github"`
}

func (c *Config) GetSources() []RepositorySource {
	sources := make([]RepositorySource, len(c.Github))

	offset := 0
	for i := 0; i < len(c.Github); i++ {
		sources[i+offset] = c.Github[i]
		offset++
	}

	return sources
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
