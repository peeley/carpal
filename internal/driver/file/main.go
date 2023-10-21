package file

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
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

func (d fileDriver) GetResource(name string) (*resource.Resource, error) {
	baseDirectory := path.Clean(d.Configuration.FileConfiguration.Directory)

	resourceFile, err := os.ReadFile(path.Join(baseDirectory, name))
	if err != nil {
		log.Printf("unable to read resource file: %v", err)
		if errors.Is(err, os.ErrNotExist) {
			return nil, driver.ResourceNotFound{ResourceName: name}
		} else {
			return nil, fmt.Errorf("resource file not found: %w", err)
		}
	}

	var resource resource.Resource
	err = yaml.Unmarshal(resourceFile, &resource)
	if err != nil {
		log.Printf("unable to unmarshal resource file contents: %v", err)
		return nil, fmt.Errorf("could not unmarshal file to JRD: %w", err)
	}

	resource.Subject = name

	return &resource, nil
}
