package resource

import (
	"encoding/json"
	"fmt"
)

type Properties map[string]any

type Link struct {
	Rel string
	Type *string
	Href *string
	Titles []string
	Properties Properties
}

type Resource struct {
	Subject string
	Aliases []string
	Properties Properties
	Links []Link
}

func MarshalResource(resource Resource) ([]byte, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("could not marshal resource to json: %w", err)
	}

	return jsonBytes, nil
}
