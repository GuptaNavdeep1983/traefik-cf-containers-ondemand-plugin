package traefik_cf_containers_ondemand_plugin

import (
	"encoding/json"
	"net/http"
	"time"
)

type AppStartActionResponse struct {
	GUID      string    `json:"guid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
}

type listV3AppsResponse struct {
	Pagination Pagination `json:"pagination,omitempty"`
	Resources  []V3App    `json:"resources,omitempty"`
}

type V3App struct {
	Name string `json:"name,omitempty"`
	GUID string `json:"guid,omitempty"`
}

func GetAppIds(config Config, spaceGuid string) ([]V3App, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", config.ApiEndpoint+"/v3/apps?names="+config.Apps+"&space_guids="+spaceGuid, nil)
	if err != nil {
		return []V3App{}, err
	}

	req.Header = http.Header{
		"Authorization": []string{"Bearer " + config.Token},
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	var apps []V3App
	var data listV3AppsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return []V3App{}, err
	}
	apps = append(apps, data.Resources...)
	return apps, nil
}

func StartApps(config Config, apps []V3App) ([]AppStartActionResponse, error) {
	client := http.Client{}
	var responses []AppStartActionResponse
	for _, app := range apps {
		req, err := http.NewRequest("POST", config.ApiEndpoint+"/v3/apps/"+app.GUID+"/actions/start", nil)
		if err != nil {
			return []AppStartActionResponse{}, err
		}
		req.Header = http.Header{
			"Authorization": []string{"Bearer " + config.Token},
		}

		resp, err := client.Do(req)
		defer resp.Body.Close()

		var data AppStartActionResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return []AppStartActionResponse{}, err
		}
		responses = append(responses, data)
	}
	return responses, nil
}
