package driver

import (
	"github.com/peeley/carpal/internal/resource"
)

type Driver interface {
	GetResource(string) (*resource.Resource, error)
}
