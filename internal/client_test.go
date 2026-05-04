package internal

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aniruddha-sinha/jiraffe/mocks"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"
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

func TestIsTokenValid_WithGomock(t *testing.T) {
	// 1. Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		encodedToken  string
		setupMock     func(mockClient *mocks.MockHTTPClient)
		wantValid     bool
		wantErr       bool
		errorContains string
	}{
		{
			name:         "Valid token returns 200 OK",
			encodedToken: "validDingDong",
			setupMock: func(mockClient *mocks.MockHTTPClient) {
				// Tell gomock: Expect 'Do' to be called once, with any request.
				// Return a mock Response with a 200 status code.
				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					// You must provide a dummy body because the function calls resp.Body.Close()
					Body: io.NopCloser(strings.NewReader("")),
				}
				mockClient.EXPECT().Do(gomock.Any()).Return(mockResponse, nil).Times(1)
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:         "Faulty token returns 401",
			encodedToken: "invalidDingDong",
			setupMock: func(mockClient *mocks.MockHTTPClient) {
				mockResponse := &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader("")),
				}
				mockClient.EXPECT().Do(gomock.Any()).Return(mockResponse, nil).Times(1)
			},
			wantValid:     false,
			wantErr:       true,
			errorContains: "unauthorized",
		},
		{
			name:         "Client network error",
			encodedToken: "anyToken",
			setupMock: func(mockClient *mocks.MockHTTPClient) {
				// Simulate an actual network failure (e.g., DNS resolution failed)
				mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("network timeout")).Times(1)
			},
			wantValid:     false,
			wantErr:       true,
			errorContains: "network timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 2. Create a new instance of the mock client for this specific test
			mockClient := mocks.NewMockHTTPClient(ctrl)

			// 3. Set up the expected behavior for this test case
			tt.setupMock(mockClient)

			p := UserProfile{Org: "test-org"}

			// 4. Call the function, passing in the mock client
			valid, err := IsTokenValid(p, mockClient, tt.encodedToken)

			// 5. Assertions
			if valid != tt.wantValid {
				t.Errorf("IsTokenValid() valid = %v, want %v", valid, tt.wantValid)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("IsTokenValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("IsTokenValid() error = %v, expected it to contain %v", err, tt.errorContains)
				}
			}
		})
	}
}
