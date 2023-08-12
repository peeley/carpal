package resource

import (
	"encoding/json"
	"fmt"
)

type Properties map[string]any

type Link struct {
	Rel        string     `json:"rel"`
	Type       *string    `json:"type,omitempty"`
	Href       *string    `json:"href,omitempty"`
	Titles     []string   `json:"titles,omitempty"`
	Properties Properties `json:"properties,omitempty"`
}

type Resource struct {
	Subject    string     `json:"subject"`
	Aliases    []string   `json:"aliases,omitempty"`
	Properties Properties `json:"properties,omitempty"`
	Links      []Link     `json:"links,omitempty"`
}

func MarshalResource(resource Resource) ([]byte, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("could not marshal resource to json: %w", err)
	}

	return jsonBytes, nil
}
