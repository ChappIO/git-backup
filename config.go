package git_backup

import (
	"bytes"
	"io"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Github []*GithubConfig `yaml:"github"`
	GitLab []*GitLabConfig `yaml:"gitlab"`
}

func (c *Config) GetSources() []RepositorySource {
	sources := make([]RepositorySource, len(c.Github)+len(c.GitLab))

	offset := 0
	for i := 0; i < len(c.Github); i++ {
		sources[offset] = c.Github[i]
		offset++
	}
	for i := 0; i < len(c.GitLab); i++ {
		sources[offset] = c.GitLab[i]
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
	if c.GitLab != nil {
		for _, config := range c.GitLab {
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
	data, err := io.ReadAll(reader)
	if err != nil {
		return
	}
	rendered, err := parse(string(data))
	if err != nil {
		return
	}

	dec := yaml.NewDecoder(rendered)
	dec.KnownFields(true)
	err = dec.Decode(&out)
	out.setDefaults()
	return
}

func parse(rawTemplate string) (rendered io.Reader, err error) {
	fmap := template.FuncMap{
		"env": os.Getenv,
	}

	tmpl := template.New("").Funcs(fmap).Option("missingkey=error")

	tmpl, err = tmpl.Parse(rawTemplate)
	if err != nil {
		return
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, nil)
	rendered = bytes.NewReader(buf.Bytes())

	return
}
