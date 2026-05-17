package jira

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrUnauthorized = errors.New("unauthorized: invalid API token or email")

type client struct {
	hc       *http.Client
	p        profile // Capitalized to match the exported Profile type
	apiToken string
}

func NewClient(hc *http.Client, validProfile profile, apiToken string) *client {
	// Safeguard the HTTP client
	if hc == nil {
		hc = &http.Client{}
	}

	return &client{
		hc:       hc,
		p:        validProfile,
		apiToken: apiToken,
	}
}

// HandleAuthentication verifies the credentials against the Jira API.
func (c *client) HandleAuthentication() error {
	// Construct the URL using the profile's domain/org
	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/3/myself", c.p.Org())

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Basic Auth header using the unexported profile data and token
	req.SetBasicAuth(c.p.Email(), c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("network error during authentication: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // TODO: check error and return

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
