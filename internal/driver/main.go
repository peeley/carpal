package driver

import (
	"fmt"

	"github.com/peeley/carpal/internal/resource"
)

type Driver interface {
	GetResource(string) (*resource.Resource, error)
}

type ResourceNotFound struct {
	ResourceName string
}

func (err ResourceNotFound) Error() string {
	return fmt.Sprintf("resource not found: %s", err.ResourceName)
}
