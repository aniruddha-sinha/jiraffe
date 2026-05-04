package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

func IsTokenValid(p UserProfile, client HTTPClient, encodedAPItoken string) (bool, error) {
	// isTokenValid() ==then move on] otherwise ask user to scratch his head
	ctx := context.Background()

	url := BuildURL(p.Org, EndpointMyself)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	r.Header.Add("Authorization", "Basic "+encodedAPItoken) //! there should be a SPACE post Basic if token is dingDong then the Authorization
	//! is BasicdingDong which cannot be parsed by JIRA servers
	r.Header.Add("Accept", "application/json")

	resp, err := client.Do(r)
	if err != nil {
		return false, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return false, errors.New("unauthorized; might be faulty token")
	}

	return false, fmt.Errorf("unexpected status code %d", resp.StatusCode)
}
