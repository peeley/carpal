package file

import (
	"fmt"
	"os"
	"path"

	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/resource"
	"github.com/peeley/carpal/internal/driver"
	"gopkg.in/yaml.v3"
)

type fileDriver struct {
	Configuration config.Configuration
}

func NewFileDriver(config config.Configuration) driver.Driver {
	return fileDriver{
		config,
	}
}

func (driver fileDriver) GetResource(name string) (*resource.Resource, error) {
	baseDirectory := path.Clean(driver.Configuration.FileConfiguration.Directory)

	resourceFile, err := os.ReadFile(path.Join(baseDirectory, name))
	if err != nil {
		return nil, fmt.Errorf("resource file not found: %w", err)
	}

	var resource resource.Resource
	err = yaml.Unmarshal(resourceFile, &resource)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file to JRD: %w", err)
	}

	resource.Subject = "acct:" + name + "@" + driver.Configuration.Domain

	return &resource, nil
}
