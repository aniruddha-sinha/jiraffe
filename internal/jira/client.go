package jira

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	APIVersion     = "3"
	EndpointMyself = "/rest/api/%s/myself"
)

func BuildURL(org, endpointTemplate string) string {
	baseURL := viper.GetString("jira.base_url")

	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.atlassian.net", org)
	}

	baseURL = strings.TrimRight(baseURL, "/")
	endpoint := fmt.Sprintf(endpointTemplate, APIVersion)

	return baseURL + endpoint
}
