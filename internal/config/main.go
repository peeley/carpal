package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigWizard interface {
	readConfigFile() ([]byte, error)
	readSecretFromFile(filePath string) (string, error)
	deserializeConfigYaml([]byte) (*Configuration, error)
	processConfigYaml([]byte) (*Configuration, error)
	GetConfiguration() (*Configuration, error)
	processLDAPBindPassword(config *Configuration) error
	processDatabaseURL(config *Configuration) error
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
	URL          string   `yaml:"url"`
	BindUser     string   `yaml:"bind_user"`
	BindPass     string   `yaml:"bind_pass"`
	BindPassFile string   `yaml:"bind_pass_file"`
	BaseDN       string   `yaml:"basedn"`
	Filter       string   `yaml:"filter"`
	UserAttr     string   `yaml:"user_attr"`
	Attributes   []string `yaml:"attributes"`
	Template     string   `yaml:"template"`
}

type DatabaseConfiguration struct {
	Driver      string   `yaml:"driver"`       // e.g., "postgres"
	URL         string   `yaml:"url"`          // Database connection URL
	URLFile     string   `yaml:"url_file"`     // File containing database connection URL
	Table       string   `yaml:"table"`        // Table name
	KeyColumn   string   `yaml:"key_column"`   // Column to search by (e.g., "uid")
	ColumnNames []string `yaml:"column_names"` // Mapping of column names to template variables
	Template    string   `yaml:"template"`     // Path to the template file
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

func (wiz configWizard) processConfigYaml(
	configYaml []byte,
) (*Configuration, error) {
	config, err := wiz.deserializeConfigYaml(configYaml)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config YAML: %w", err)
	}

	if err := wiz.processLDAPBindPassword(config); err != nil {
		return nil, err
	}

	if err := wiz.processDatabaseURL(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (wiz configWizard) readSecretFromFile(filePath string) (string, error) {
	secretBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read secret from %s: %w", filePath, err)
	}

	return string(bytes.TrimSpace(secretBytes)), nil
}

func (wiz configWizard) processLDAPBindPassword(config *Configuration) error {
	if config.LDAPConfiguration == nil {
		return nil
	}

	hasBindPass := config.LDAPConfiguration.BindPass != ""
	hasBindPassFile := config.LDAPConfiguration.BindPassFile != ""

	if hasBindPass == hasBindPassFile {
		return fmt.Errorf("must specify either bind_pass or bind_pass_file")
	}

	if config.LDAPConfiguration.BindPassFile != "" {
		password, err := wiz.readSecretFromFile(config.LDAPConfiguration.BindPassFile)
		if err != nil {
			return fmt.Errorf("cannot read LDAP bind password file: %w", err)
		}

		config.LDAPConfiguration.BindPass = password
	}

	return nil
}

func (wiz configWizard) processDatabaseURL(config *Configuration) error {
	if config.DatabaseConfiguration == nil {
		return nil
	}

	hasURL := config.DatabaseConfiguration.URL != ""
	hasURLFile := config.DatabaseConfiguration.URLFile != ""

	if hasURL == hasURLFile {
		return fmt.Errorf("must specify either url or url_file")
	}

	if config.DatabaseConfiguration.URLFile != "" {
		url, err := wiz.readSecretFromFile(config.DatabaseConfiguration.URLFile)
		if err != nil {
			return fmt.Errorf("cannot read database URL file: %w", err)
		}

		config.DatabaseConfiguration.URL = url
	}

	return nil
}

func (wiz configWizard) GetConfiguration() (*Configuration, error) {
	configYaml, err := wiz.readConfigFile()
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	config, err := wiz.processConfigYaml(configYaml)
	if err != nil {
		return nil, err
	}

	return config, nil
}
