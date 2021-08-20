package traefik_cf_containers_ondemand_plugin

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	Jti          string `json:"jti"`
}
type Endpoint struct {
	AuthorizationEndpoint    string `json:"authorization_endpoint"`
	TokenEndpoint            string `json:"token_endpoint"`
	AppSSHEndpoint           string `json:"app_ssh_endpoint"`
	AppSSHHostKeyFingerprint string `json:"app_ssh_host_key_fingerprint"`
	AppSSHOauthClient        string `json:"app_ssh_oauth_client"`
	DopplerLoggingEndpoint   string `json:"doppler_logging_endpoint"`
	APIVersion               string `json:"api_version"`
	OsbapiVersion            string `json:"osbapi_version"`
	RoutingEndpoint          string `json:"routing_endpoint"`
}

func GetInfo(config Config) (Endpoint, error) {
	client := http.Client{}
	var data Endpoint
	req, err := http.NewRequest("GET", config.ApiEndpoint+"/v2/info", nil)
	if err != nil {
		return data, err
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return data, err
	}
	return data, nil
}
func GetToken(config Config, authorizationEndpoint string) (LoginResponse, error) {
	client := http.Client{}

	var loginResponse LoginResponse
	var jsonStr = []byte(fmt.Sprintf("username=%s&password=%s&client_id=cf&grant_type=password&response_type=token", config.Username, config.Password))
	req, err := http.NewRequest("POST", authorizationEndpoint + "/oauth/token", bytes.NewBuffer(jsonStr))
	
	if err != nil {
		return loginResponse, err
	}
	
	encodedHeader := b64.StdEncoding.EncodeToString([]byte("cf:"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+encodedHeader)

	resp, err := client.Do(req)
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return loginResponse, err
	}
	return loginResponse, nil
}
