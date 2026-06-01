package jira

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to generate mock credentials for tests
func mockCreds() *JiraCreds {
	return NewJiraCreds("test@example.com", "testorg", "mock-token")
}

func TestNewClient(t *testing.T) {
	creds := mockCreds()
	client := NewClient(creds)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, creds, client.creds)
	assert.Equal(t, 10, int(client.httpClient.Timeout.Seconds()))
}

func TestClient_buildRawURL(t *testing.T) {
	client := NewClient(mockCreds())

	t.Run("valid path format", func(t *testing.T) {
		url, err := client.buildRawURL(urlTemplateIssueGet, apiVersion, "PROJ-123")
		require.NoError(t, err)
		assert.Contains(t, url, "/rest/api/3/issue/PROJ-123")
	})
}

func TestClient_NewRequest(t *testing.T) {
	client := NewClient(mockCreds())
	ctx := context.Background()

	req, err := client.NewRequest(ctx, http.MethodGet, "https://testorg.atlassian.net/rest/api/3/myself")

	require.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Contains(t, req.Header.Get("Authorization"), "Basic ")
}

func TestClient_Do(t *testing.T) {
	client := NewClient(mockCreds())
	testURL := "https://testorg.atlassian.net/api/test"

	tests := []struct {
		name           string
		method         string
		mockStatusCode int
		mockRespBody   string
		setupMock      bool
		expectedErr    error
	}{
		{
			name:           "success 200 OK",
			method:         http.MethodGet,
			mockStatusCode: http.StatusOK,
			mockRespBody:   `{"success": true}`,
			setupMock:      true,
			expectedErr:    nil,
		},
		{
			name:           "unauthorized 401",
			method:         http.MethodGet,
			mockStatusCode: http.StatusUnauthorized,
			mockRespBody:   `{"error": "unauthorized"}`,
			setupMock:      true,
			expectedErr:    ErrUnauthorizedRequest,
		},
		{
			name:           "unexpected status code 500",
			method:         http.MethodGet,
			mockStatusCode: http.StatusInternalServerError,
			mockRespBody:   `{"error": "server error"}`,
			setupMock:      true,
			expectedErr:    ErrUnexpectedStatusCode,
		},
		{
			name:        "network failure",
			method:      http.MethodGet,
			setupMock:   false,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(client.httpClient)
			defer httpmock.DeactivateAndReset()

			if tt.setupMock {
				httpmock.RegisterResponder(tt.method, testURL,
					httpmock.NewStringResponder(tt.mockStatusCode, tt.mockRespBody))
			}

			resp, err := client.Do(context.Background(), tt.method, testURL)

			if tt.expectedErr != nil {
				if errors.Is(tt.expectedErr, assert.AnError) {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.mockStatusCode, resp.StatusCode)
			}
		})
	}
}

func TestClient_validateToken(t *testing.T) {
	client := NewClient(mockCreds())
	validateURL := "https://testorg.atlassian.net/rest/api/3/myself"

	tests := []struct {
		name           string
		mockStatusCode int
		setupMock      bool
		expectedErr    error
	}{
		{
			name:           "valid token",
			mockStatusCode: http.StatusOK,
			setupMock:      true,
			expectedErr:    nil,
		},
		{
			name:           "invalid token 401",
			mockStatusCode: http.StatusUnauthorized,
			setupMock:      true,
			expectedErr:    ErrUnauthorizedRequest,
		},
		{
			name:           "unexpected error 404",
			mockStatusCode: http.StatusNotFound,
			setupMock:      true,
			expectedErr:    ErrUnexpectedStatusCode,
		},
		{
			name:        "http client error",
			setupMock:   false,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(client.httpClient)
			defer httpmock.DeactivateAndReset()

			if tt.setupMock {
				httpmock.RegisterResponder(http.MethodGet, validateURL,
					httpmock.NewStringResponder(tt.mockStatusCode, ""))
			}

			err := client.validateToken(context.Background(), validateURL)

			if tt.expectedErr != nil {
				if errors.Is(tt.expectedErr, assert.AnError) {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
