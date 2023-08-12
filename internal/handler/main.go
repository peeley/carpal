package handler

import (
	"errors"
	"net/http"

	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
)

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type resourceHandler struct {
	Driver driver.Driver
}

func NewResourceHandler(driver driver.Driver) Handler {
	return resourceHandler{driver}
}

func (handler resourceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}

	resourceParam := r.URL.Query().Get("resource")

	if resourceParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		return
	}

	resourceStruct, err := handler.Driver.GetResource(resourceParam)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}

	JRD, err := resource.MarshalResource(*resourceStruct)
	if err != nil {
		if errors.Is(err, driver.ResourceNotFound{}) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/jrd+json")
	w.Write(JRD)
}
