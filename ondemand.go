package traefik_cf_containers_ondemand_plugin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const defaultApiEndpoint = "https://api.run.pivotal.io"

// Config the plugin configuration
type Config struct {
	Name        string
	ApiEndpoint string
	OrgName     string
	SpaceName   string
	Username    string
	Password    string
	Apps        string
	Token       string
}

// CreateConfig creates a config with its default values
func CreateConfig() *Config {
	return &Config{
		ApiEndpoint: defaultApiEndpoint,
		OrgName: "TEST_ORG",
		SpaceName: "TEST_SPACE",
		Apps: "APPS_TO_BE_SCALED",
		Name: "traefik_cf_containers_ondemand",
		Username: "TEST_USER",
		Password: "TEST_PASSWORD",
	}
}

// Ondemand holds the request for the on demand service
type Ondemand struct {
	name     string
	next     http.Handler
	config   Config
	endpoint Endpoint
}

// New function creates the configuration and end points
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	if len(config.ApiEndpoint) == 0 {
		return nil, fmt.Errorf("ApiEndpoint cannot be null")
	}

	if len(config.OrgName) == 0 {
		return nil, fmt.Errorf("OrgName cannot be null")
	}

	if len(config.SpaceName) == 0 {
		return nil, fmt.Errorf("SpaceName cannot be null")
	}

	if len(config.Apps) == 0 {
		return nil, fmt.Errorf("Apps cannot be null")
	}

	if len(config.Name) == 0 {
		return nil, fmt.Errorf("Name cannot be null")
	}

	endpoint, err := GetInfo(*config)
	if err != nil {
		return nil, fmt.Errorf("Error while getting apiendpoint info")
	}
	log.Printf("%+v\n", endpoint)

	return &Ondemand{
		next:     next,
		name:     name,
		config:   *config,
		endpoint: endpoint,
	}, nil
}

// ServeHTTP retrieve the service status
func (e *Ondemand) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	status, err := getServiceStatus(&e.endpoint, &e.config)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
	}

	if status == "started" {
		// Service started forward request
		e.next.ServeHTTP(rw, req)

	} else if status == "starting" {
		// Service starting, notify client
		rw.WriteHeader(http.StatusAccepted)
		rw.Write([]byte("Service is starting..."))
	} else {
		// Error
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Unexpected status answer from ondemand service"))
	}
}

func getServiceStatus(endpoint *Endpoint, config *Config) (string, error) {

	// Get Access Token
	loginResponse, err := GetToken(*config, endpoint.AuthorizationEndpoint)
	if err != nil {
		return "error_starting", fmt.Errorf("Error in getting token")
	}
	config.Token = loginResponse.AccessToken

	// Get CF Space GUID
	spaceGuid, err := GetSpace(*config)
	if err != nil {
		return "error_starting", fmt.Errorf("Error in getting space guid")
	}
	log.Printf("%s\n", spaceGuid)

	// Get App GUID's
	apps, err := GetAppIds(*config, spaceGuid)
	if err != nil {
		return "error_starting", fmt.Errorf("Error in getting app guid")
	}
	log.Printf("%+v\n", apps)

	// Start Apps
	appResponses, err := StartApps(*config, apps)
	if err != nil {
		return "error_starting", fmt.Errorf("Error in starting app using guids")
	}
	log.Printf("%+v\n", appResponses)

	return appResponses[0].State, nil
}
