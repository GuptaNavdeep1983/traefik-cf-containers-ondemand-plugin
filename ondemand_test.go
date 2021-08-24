package traefik_cf_containers_ondemand_plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/json"
	"strings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOndemand(t *testing.T) {
	testCases := []struct {
		desc          string
		config        *Config
		expectedError bool
	}{
		{
			desc: "invalid Config",
			config: &Config{
				ApiEndpoint: "",
				OrgName: "",
				SpaceName: "",
				Apps: "",
				Username: "",
				Password: "",
				Name: "TRAEFIK_CF_CONTAINERS_ONDEMAND",
			},
			expectedError: true,
		},
		{
			desc: "valid Config",
			config: &Config{
				ApiEndpoint: "https://api.run.pivotal.io",
				OrgName: "DEFAULT_ORG",
				SpaceName: "DEFAULT_SPACE",
				Apps: "TEST_APP",
				Username: "TEST_USER",
				Password: "TEST_PASS",
				Name: "TRAEFIK_CF_CONTAINERS_ONDEMAND",
			},
			expectedError: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			ondemand, err := New(context.Background(), next, test.config, "traefikTest")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, ondemand)
			}
		})
	}
}

func TestOndemand_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc     string
		status   string
		expected int
	}{
		{
			desc:     "service is starting",
			status:   "STARTING",
			expected: 202,
		},
		{
			desc:     "service is started",
			status:   "STARTED",
			expected: 200,
		},
		{
			desc:     "ondemand service is in error",
			status:   "error",
			expected: 500,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// Mock Server for Auth endpoint to generate token
			mockServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/oauth/token") {
					loginResponse := LoginResponse{}
					bytes, _ := json.Marshal(loginResponse)
					fmt.Fprint(w, string(bytes[:]))
				}
			}))
			// Mock server for Resource end point to generate spaces, apps and info responses
			defer mockServer1.Close()
			mockServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/v3/spaces") {
					V3SpacesResponse := listV3SpacesResponse{}
					V3SpacesResponse.Resources = 
					[]V3Space{ 
						{
							GUID: "test-space-guid",
						},
					}
					bytes, _ := json.Marshal(V3SpacesResponse)
					fmt.Fprint(w, string(bytes[:]))
				} else if strings.Contains(r.URL.Path, "/v3/apps/") {
					appStartResponse := AppStartActionResponse{
							State: test.status,
					}
					bytes, _ := json.Marshal(appStartResponse)
					fmt.Fprint(w, string(bytes[:]))
				} else if strings.Contains(r.URL.Path, "/v3/apps") {
					V3AppsResponse := listV3AppsResponse{}
					V3AppsResponse.Resources = []V3App{
						{
							GUID: "test-app-guid",
						},
					}
					bytes, _ := json.Marshal(V3AppsResponse)
					fmt.Fprint(w, string(bytes[:]))
				} else {
					endpoint := Endpoint{}
					endpoint.AuthorizationEndpoint = mockServer1.URL
					bytes, _ := json.Marshal(endpoint)
					fmt.Fprint(w, string(bytes[:]))
				}
				
			}))
			defer mockServer2.Close()
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			config := &Config{
				ApiEndpoint: mockServer2.URL,
				OrgName: "DEFAULT_ORG",
				SpaceName: "DEFAULT_SPACE",
				Apps: "TEST_APP",
				Username: "TEST_USER",
				Password: "TEST_PASS",
				Name: "TRAEFIK_CF_CONTAINERS_ONDEMAND",
			}
			ondemand, err := New(context.Background(), next, config, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://mydomain/", nil)

			ondemand.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}