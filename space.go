package traefik_cf_containers_ondemand_plugin

import (
	"encoding/json"
	"net/http"
)

type listV3SpacesResponse struct {
	Pagination Pagination `json:"pagination,omitempty"`
	Resources  []V3Space  `json:"resources,omitempty"`
}
type Pagination struct {
	TotalResults int  `json:"total_results"`
	TotalPages   int  `json:"total_pages"`
	First        Link `json:"first"`
	Last         Link `json:"last"`
	Next         Link `json:"next"`
	Previous     Link `json:"previous"`
}
type V3Space struct {
	Name          string                         `json:"name,omitempty"`
	GUID          string                         `json:"guid,omitempty"`
	CreatedAt     string                         `json:"created_at,omitempty"`
	UpdatedAt     string                         `json:"updated_at,omitempty"`
	Relationships map[string]V3ToOneRelationship `json:"relationships,omitempty"`
	Links         map[string]Link                `json:"links,omitempty"`
	Metadata      V3Metadata                     `json:"metadata,omitempty"`
}

type Link struct {
	Href   string `json:"href"`
	Method string `json:"method,omitempty"`
}

type V3ToOneRelationship struct {
	Data V3Relationship `json:"data,omitempty"`
}

type V3Relationship struct {
	GUID string `json:"guid,omitempty"`
}

type V3Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func GetSpace(config Config) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", config.ApiEndpoint+"/v3/spaces?names="+config.SpaceName, nil)
	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"Authorization": []string{"Bearer " + config.Token},
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	var spaces []V3Space
	var data listV3SpacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	spaces = append(spaces, data.Resources...)
	return spaces[0].GUID, nil
}
