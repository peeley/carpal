package handler

import (
	"errors"
	"log/slog"
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
	slog.Info("received request for resource", "resource_name", resourceParam)

	if resourceParam == "" {
		slog.Warn("received blank resource request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		return
	}

	resourceStruct, err := handler.Driver.GetResource(resourceParam)
	if err != nil {
		if errors.As(err, &driver.ResourceNotFound{}) {
			slog.Warn("resource not found", "resource_name", resourceParam, "err", err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		} else {
			slog.Error("error retrieving resource", "resource_name", resourceParam, "err", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("bad gateway"))
			return
		}
	}

	relParams := r.URL.Query()["rel"]
	if len(relParams) != 0 {
		relParamsSet := make(map[string]bool)
		for _, rel := range(relParams) {
			relParamsSet[rel] = true
		}

		filteredResourceLinks := []resource.Link{}
		for _, link := range(resourceStruct.Links) {

			_, ok := relParamsSet[link.Rel]
			if ok {
				filteredResourceLinks = append(filteredResourceLinks, link)
			}
		}

		resourceStruct.Links = filteredResourceLinks
	}

	JRD, err := resource.MarshalResource(*resourceStruct)
	if err != nil {
		slog.Error("unable to marshal resource", "resource_name", resourceParam, "err", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/jrd+json")
	w.WriteHeader(http.StatusOK)
	w.Write(JRD)
}
