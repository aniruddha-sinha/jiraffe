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
	apiVersion                = "3"
	baseURLTemplate           = "https://%s.atlassian.net"
	endpointMyselfValidateAPI = "/rest/api/%s/myself"
)

var (
	ErrUnauthorizedRequest          = errors.New("unauthorised request; might be a faulty token")
	ErrUnexpectedStatusCode         = errors.New("unexpected status code")
	ErrTokenReadFailure             = errors.New("failed to read token")
	ErrAPITokenValidityVerification = errors.New("token validation failed")
)

type Client struct {
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) BuildBaseURL(org, path string) (string, error) {
	base, err := url.Parse(fmt.Sprintf(baseURLTemplate, org))
	if err != nil {
		return "", fmt.Errorf("error building base url: %w", err)
	}

	finalURL := base.JoinPath(path)
	return finalURL.String(), nil
}

func (c *Client) getTokenValidatorAPIURL(org string) (string, error) {
	apiPath := fmt.Sprintf(endpointMyselfValidateAPI, apiVersion)
	fullURL, err := c.BuildBaseURL(org, apiPath)
	if err != nil {
		return "", fmt.Errorf("failed to construct API URL: %w", err)
	}

	return fullURL, nil
}

func (c *Client) validateToken(ctx context.Context, validateTokenApiURL, encodedAPIToken string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, validateTokenApiURL, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Basic "+encodedAPIToken)
	request.Header.Add("Accept", "application/json")

	response, err := c.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close the response body: %w", closeErr))
		}
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
