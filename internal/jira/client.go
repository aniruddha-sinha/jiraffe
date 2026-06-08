package jira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	apiVersion      = "3"
	baseURLTemplate = "https://%s.atlassian.net"

	urlTemplateValidateMyselfAPI          = "/rest/api/%s/myself"
	urlTemplateIssueSearchAPIJQL          = "/rest/api/%s/search/jql"
	urlTemplateListProjects               = "/rest/api/%s/project"
	urlTemplateProjectSearch              = "/rest/api/%s/project/%s"
	urlTemplateIssueGet                   = "/rest/api/%s/issue/%s"
	urlTemplateIssueCreate                = "/rest/api/%s/issue"
	urlTemplateGetFields                  = "/rest/api/%s/field"
	urlTemplateSearchAtlassianUserByEmail = "/rest/api/%s/user/search"
)

var (
	ErrUnauthorizedRequest          = errors.New("unauthorised request; might be a faulty token")
	ErrUnexpectedStatusCode         = errors.New("unexpected status code")
	ErrTokenReadFailure             = errors.New("failed to read token")
	ErrAPITokenValidityVerification = errors.New("token validation failed")
)

type JiraUser struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

func NewJiraUser(acctID, displayName string) *JiraUser {
	return &JiraUser{
		AccountID:   acctID,
		DisplayName: displayName,
	}
}

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

func (c *Client) buildRawURL(pathTemplate string, pathArgs ...any) (string, error) {
	rawBaseURL := fmt.Sprintf(baseURLTemplate, c.creds.Org())
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	formattedPath := fmt.Sprintf(pathTemplate, pathArgs...)
	finalURL := baseURL.JoinPath(formattedPath)
	return finalURL.String(), nil
}

func (c *Client) buildURLForQueryParams(pathTemplate string, pathArgs ...any) (*url.URL, error) {
	fullURL, err := c.buildRawURL(pathTemplate, pathArgs...)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}

	return parsedURL, nil
}

func (c *Client) NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+c.creds.EncodedAPIToken())
	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) Do(ctx context.Context, method, url string, payload io.Reader) (*http.Response, error) {
	request, err := c.NewRequest(ctx, method, url, payload)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		return nil, fmt.Errorf("jira API error (status %d): %s", response.StatusCode, string(bodyBytes))
	}

	return response, nil
}

func (c *Client) validateToken(ctx context.Context, validateTokenApiURL string) error {
	request, err := c.NewRequest(ctx, http.MethodGet, validateTokenApiURL, nil)
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

func (c *Client) ResolveEmailToAtlassianUserID(ctx context.Context, email string) (string, error) {
	reqURL, err := c.buildURLForQueryParams(urlTemplateSearchAtlassianUserByEmail, apiVersion)
	if err != nil {
		return "", err
	}

	query := reqURL.Query()
	query.Add("query", email)
	reqURL.RawQuery = query.Encode()
	response, err := c.Do(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", err
	}

	var jiraUser []JiraUser
	if err := json.NewDecoder(response.Body).Decode(&jiraUser); err != nil {
		return "", fmt.Errorf("failed to decode response : %w", err)
	}

	return jiraUser[0].AccountID, nil
}
