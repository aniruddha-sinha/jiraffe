package jira

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
	"golang.org/x/term"
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

func (c *Client) HandleAuthentication(ctx context.Context, jc *JiraCredentials) error {
	return c.getLocalOrWebBasedToken(ctx, jc)
}

func (c *Client) getLocalOrWebBasedToken(ctx context.Context, jc *JiraCredentials) error {
	if err := c.getLocalToken(ctx, jc); err != nil {
		// fmt.Println("stored token is invalid or expired, attempting to request new token . . . ")
		fmt.Printf("%v", err)
		return c.getTokenFromWeb(ctx, jc)
	}

	return nil
}

func (c *Client) getTokenFromWeb(ctx context.Context, jc *JiraCredentials) error {
	encodedToken, err := c.obtainEncodedTokenFromUser(jc)
	if err != nil {
		return err
	}

	fmt.Println("validating token")
	fullURL, err := c.getTokenValidatorAPIURL(jc.Org())
	if err != nil {
		return err
	}

	if err := c.validateToken(ctx, fullURL, encodedToken); err != nil {
		return fmt.Errorf("%w:%w", ErrAPITokenValidityVerification, err)
	}

	jc.apiToken = encodedToken
	// token valid, save it
	if err := c.commitSession(jc); err != nil {
		return err
	}

	return nil
}

func (c *Client) commitSession(jc *JiraCredentials) error {
	if err := config.Cfg.Upsert(JiraConfigEmailKey, jc.Email()); err != nil {
		return err
	}

	if err := config.Cfg.Upsert(JiraConfigOrgKey, jc.Org()); err != nil {
		return err
	}

	if err := config.Cfg.Upsert(JiraConfigEncodedTokenKey, jc.ApiToken()); err != nil {
		return err
	}

	return nil
}

func (c *Client) obtainEncodedTokenFromUser(jc *JiraCredentials) (string, error) {
	fmt.Printf("\n starting authentication %s for %s.atlassian.net\n", jc.Email(), jc.Org())
	fmt.Println("click this link to generate the API token")
	fmt.Println("https://id.atlassian.com/manage-profile/security/api-tokens")
	fmt.Print("enter API token -> ")
	bytePass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("%w, %w", ErrTokenReadFailure, err)
	}

	token := strings.TrimSpace(string(bytePass))
	fmt.Println("obtained token")
	authStr := fmt.Sprintf("%s:%s", jc.Email(), token)
	encodedStr := base64.StdEncoding.EncodeToString([]byte(authStr))
	return encodedStr, nil
}

func (c *Client) getLocalToken(ctx context.Context, jc *JiraCredentials) error {
	email, err := GetStoredEmail()
	if err != nil {
		return err
	}

	org, err := GetStoredOrg()
	if err != nil {
		return err
	}

	if jc.Email() != email || jc.Org() != org {
		return fmt.Errorf("account switch detected: stored token does not match requested email/org")
	}

	token, err := GetStoredEncodedToken()
	if err != nil {
		return err
	}

	fullURL, err := c.getTokenValidatorAPIURL(jc.Org())
	if err != nil {
		return err
	}

	return c.validateToken(ctx, fullURL, token)
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
