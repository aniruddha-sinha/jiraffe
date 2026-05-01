package jira

import (
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
