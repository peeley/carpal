package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigWizard interface {
	readConfigFile() ([]byte, error)
	deserializeConfigYaml([]byte) (*Configuration, error)
	GetConfiguration() (*Configuration, error)
}

type configWizard struct {
	ConfigFileLocation string
}

func NewConfigWizard(configFileLocation string) ConfigWizard {
	return configWizard{configFileLocation}
}

type FileConfiguration struct {
	Directory string `yaml:"directory"`
}

type LDAPConfiguration struct {
	URL        string   `yaml:"url"`
	BindUser   string   `yaml:"bind_user"`
	BindPass   string   `yaml:"bind_pass"`
	BaseDN     string   `yaml:"basedn"`
	Filter     string   `yaml:"filter"`
	UserAttr   string   `yaml:"user_attr"`
	Attributes []string `yaml:"attributes"`
	Template   string   `yaml:"template"`
}

type DatabaseConfiguration struct {
	// TODO
}

type Configuration struct {
	Driver                string                 `yaml:"driver"`
	FileConfiguration     *FileConfiguration     `yaml:"file"`
	LDAPConfiguration     *LDAPConfiguration     `yaml:"ldap"`
	DatabaseConfiguration *DatabaseConfiguration `yaml:"database"`
}

func (wiz configWizard) readConfigFile() ([]byte, error) {
	contents, err := os.ReadFile(wiz.ConfigFileLocation)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	return contents, nil
}

func (wiz configWizard) deserializeConfigYaml(
	configYaml []byte,
) (*Configuration, error) {
	var config Configuration
	err := yaml.Unmarshal(configYaml, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (wiz configWizard) GetConfiguration() (*Configuration, error) {
	configYaml, err := wiz.readConfigFile()
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	config, err := wiz.deserializeConfigYaml(configYaml)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config YAML: %w", err)
	}

	return config, nil
}
