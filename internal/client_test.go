package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spf13/viper"
)

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name             string
		org              string
		endpointTemplate string
		viperBaseURL     string
		expectedURL      string
	}{
		{
			name:             "Default Atlassian Cloud URL (no config)",
			org:              "asinha0493",
			endpointTemplate: EndpointMyself, // "/rest/api/%s/myself"
			viperBaseURL:     "",
			expectedURL:      "https://asinha0493.atlassian.net/rest/api/3/myself",
		},
		{
			name:             "Custom Base URL from config",
			org:              "asinha0493", // Should be ignored because of custom URL
			endpointTemplate: EndpointMyself,
			viperBaseURL:     "https://jira.internal-network.com",
			expectedURL:      "https://jira.internal-network.com/rest/api/3/myself",
		},
		{
			name:             "Custom Base URL with trailing slash (should be trimmed)",
			org:              "asinha0493",
			endpointTemplate: EndpointMyself,
			viperBaseURL:     "https://jira.internal-network.com/", // Note the trailing slash
			expectedURL:      "https://jira.internal-network.com/rest/api/3/myself",
		},
		{
			name:             "Different endpoint template",
			org:              "mycompany",
			endpointTemplate: "/rest/api/%s/search", // Pretending we have a search endpoint
			viperBaseURL:     "",
			expectedURL:      "https://mycompany.atlassian.net/rest/api/3/search",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			if tc.viperBaseURL != "" {
				viper.Set("jira.base_url", tc.viperBaseURL)
			}

			actualURL := BuildURL(tc.org, tc.endpointTemplate)
			if actualURL != tc.expectedURL {
				t.Errorf("\nExpected: %s\nGot:      %s", tc.expectedURL, actualURL)
			}
		})
	}

	viper.Reset()
}

type mockTestClient struct {
	client  *http.Client
	testURL string
}

func (m *mockTestClient) Do(req *http.Request) (*http.Response, error) {
	// Intercept the request and redirect it to the httptest server
	parsedTestURL, _ := url.Parse(m.testURL)
	req.URL.Scheme = parsedTestURL.Scheme
	req.URL.Host = parsedTestURL.Host
	req.Host = parsedTestURL.Host

	//nolint:gosec // Safe: This is a test interceptor routing to a local httptest server
	return m.client.Do(req)
}

func TestIsTokenValid(t *testing.T) {
	// Define table-driven test cases
	tests := []struct {
		name           string
		statusCode     int
		token          string
		expectedResult bool
		expectErr      bool
	}{
		{
			name:           "Valid Token Returns 200 OK",
			statusCode:     http.StatusOK,
			token:          "validToken123",
			expectedResult: true,
			expectErr:      false,
		},
		{
			name:           "Invalid Token Returns 401 Unauthorized",
			statusCode:     http.StatusUnauthorized,
			token:          "dingDong",
			expectedResult: false,
			expectErr:      true,
		},
		{
			name:           "Unexpected Server Error Returns 500",
			statusCode:     http.StatusInternalServerError,
			token:          "someToken",
			expectedResult: false,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Spin up the httptest Server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Assert: Check if the required headers are being set correctly
				expectedAuth := "Basic " + tt.token
				if auth := r.Header.Get("Authorization"); auth != expectedAuth {
					t.Errorf("Expected Authorization header '%s', got '%s'", expectedAuth, auth)
				}

				if accept := r.Header.Get("Accept"); accept != "application/json" {
					t.Errorf("Expected Accept header 'application/json', got '%s'", accept)
				}

				// Respond with the mock status code defined in our test table
				w.WriteHeader(tt.statusCode)
			}))
			// Ensure the server is closed when the test finishes
			defer server.Close()

			// 2. Setup mock data
			profile := UserProfile{
				Org: "test-org", // BuildURL will use this, but our mock client will intercept it
			}

			// Wrap the httptest client in our interceptor so requests are routed locally
			interceptClient := &mockTestClient{
				client:  server.Client(),
				testURL: server.URL,
			}

			// 3. Execute the target function
			isValid, err := IsTokenValid(profile, interceptClient, tt.token)

			// 4. Assert the outcomes
			if isValid != tt.expectedResult {
				t.Errorf("Expected isValid to be %v, got %v", tt.expectedResult, isValid)
			}

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error presence to be %v, got err: %v", tt.expectErr, err)
			}
		})
	}
}
