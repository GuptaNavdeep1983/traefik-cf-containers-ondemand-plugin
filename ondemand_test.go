package traefik_cf_containers_ondemand_plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
			status:   "starting",
			expected: 202,
		},
		{
			desc:     "service is started",
			status:   "started",
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

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, test.status)
			}))

			defer mockServer.Close()

			config := &Config{
				ApiEndpoint: "https://api.run.pivotal.io",
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

			req := httptest.NewRequest(http.MethodGet, "http://mydomain/whoami", nil)

			ondemand.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}