package main

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Defaults Receiver `yaml:"defaults"`
	Templates []string `yaml:"templates"`
	Receivers []*Receiver `yaml:"receivers"`
}

type Receiver struct {
  Name string `yaml:"name"`
	Template string `yaml:"template"`
	Sender string `yaml:"sender"`
	Subscribe *bool `yaml:"subscribe"`
	To []string `yaml:"to"`
}

func Load(s []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.UnmarshalStrict(s, cfg)
	return cfg, err
}

func LoadFile(file string) (*Config, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(content)
	if err != nil {
		return nil, err
	}
  resolveFilepaths(filepath.Dir(file), cfg)
	expandDefaults(cfg)
  return cfg, nil
}

// resolveFilepaths joins all relative paths in a configuration
// with a given base directory.
func resolveFilepaths(baseDir string, cfg *Config) {
  join := func(fp string) string {
    if len(fp) > 0 && !filepath.IsAbs(fp) {
      fp = filepath.Join(baseDir, fp)
    }
    return fp
  }

  for i, tf := range cfg.Templates {
    cfg.Templates[i] = join(tf)
  }
}

func expandDefaults(cfg *Config) {
	for _, recv := range cfg.Receivers {
		if len(recv.Template) == 0 {
			recv.Template = cfg.Defaults.Template
		}
		if len(recv.Sender) == 0 {
			recv.Sender = cfg.Defaults.Sender
		}
		if len(recv.To) == 0 {
			recv.To = cfg.Defaults.To
		}
		if recv.Subscribe == nil {
			recv.Subscribe = cfg.Defaults.Subscribe
		}
	}
}
