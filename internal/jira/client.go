package jira

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	apiVersion      = "3"
	baseURLTemplate = "https://%s.atlassian.net"

	urlTemplateValidateMyselfAPI = "/rest/api/%s/myself"
	urlTemplateSearchAPI         = "/rest/api/%s/search/jql"
	urlTemplateListProjects      = "/rest/api/%s/project"
	urlTemplateProjectSearch     = "/rest/api/%s/project/%s"
	urlTemplateIssueSearch       = "/rest/api/%s/issue/%s"
)

var (
	ErrUnauthorizedRequest          = errors.New("unauthorised request; might be a faulty token")
	ErrUnexpectedStatusCode         = errors.New("unexpected status code")
	ErrTokenReadFailure             = errors.New("failed to read token")
	ErrAPITokenValidityVerification = errors.New("token validation failed")
)

type Client struct {
	httpClient *http.Client
	creds      *JiraCreds
}

func NewClient(creds *JiraCreds) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		creds:      creds,
	}

	return c
}

func (c *Client) BuildURL(pathTemplate string, pathArgs ...any) (string, error) {
	rawBaseURL := fmt.Sprintf(baseURLTemplate, c.creds.Org())
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	formattedPath := fmt.Sprintf(pathTemplate, pathArgs...)
	finalURL := baseURL.JoinPath(formattedPath)
	return finalURL.String(), nil
}

func (c *Client) NewRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+c.creds.EncodedAPIToken())
	req.Header.Add("Accept", "application/json")
	return req, nil
}

func (c *Client) validateToken(ctx context.Context, validateTokenApiURL string) error {
	request, err := c.NewRequest(ctx, http.MethodGet, validateTokenApiURL)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	return mapStatusToError(response.StatusCode)
}

func mapStatusToError(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return ErrUnauthorizedRequest
	default:
		return ErrUnexpectedStatusCode
	}
}
