package file

import (
	"errors"
	"fmt"
	"log/slog"
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
		slog.Error("unable to read resource file", "err", err)
		if errors.Is(err, os.ErrNotExist) {
			return nil, driver.ResourceNotFound{ResourceName: name}
		} else {
			return nil, fmt.Errorf("resource file not found: %w", err)
		}
	}

	var resource resource.Resource
	err = yaml.Unmarshal(resourceFile, &resource)
	if err != nil {
		slog.Error("unable to unmarshal resource file contents", "err", err)
		return nil, fmt.Errorf("could not unmarshal file to JRD: %w", err)
	}

	resource.Subject = name

	return &resource, nil
}
