package jira

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue_Properties(t *testing.T) {
	issue := &Issue{
		Key: "PROJ-123",
		Fields: IssueFields{
			Summary:   "Fix the bug",
			Status:    IssueStatus{Name: "In Progress"},
			Priority:  IssuePriority{Name: "High"},
			IssueType: IssueType{Name: "Task"},
			Created:   "2023-01-01T10:00:00Z",
			Updated:   "2023-01-02T12:00:00Z",
		},
	}

	t.Run("basic properties", func(t *testing.T) {
		assert.Equal(t, "Fix the bug", issue.Summary())
		assert.Equal(t, "In Progress", issue.Status())
		assert.Equal(t, "High", issue.Priority())
		assert.Equal(t, "Task", issue.Type())
		assert.Equal(t, "2023-01-01T10:00:00Z", issue.Created())
		assert.Equal(t, "2023-01-02T12:00:00Z", issue.Updated())
	})

	t.Run("assignee - unassigned", func(t *testing.T) {
		assert.Equal(t, "unassigned", issue.Assignee())
	})

	t.Run("assignee - assigned", func(t *testing.T) {
		issue.Fields.Assignee = &IssueUser{DisplayName: "John Doe"}
		assert.Equal(t, "John Doe", issue.Assignee())
	})

	t.Run("description - empty or null", func(t *testing.T) {
		issue.Fields.Description = json.RawMessage(`null`)
		assert.Equal(t, "", issue.Description())

		issue.Fields.Description = json.RawMessage(``)
		assert.Equal(t, "", issue.Description())
	})

	t.Run("description - plain string", func(t *testing.T) {
		issue.Fields.Description = json.RawMessage(`"This is a plain text description"`)
		assert.Equal(t, "This is a plain text description", issue.Description())
	})

	t.Run("description - ADF document", func(t *testing.T) {
		adfJSON := `{
			"content": [
				{
					"content": [
						{"text": "Hello "},
						{"text": "World!"}
					]
				},
				{
					"content": [
						{"text": "Second paragraph."}
					]
				}
			]
		}`
		issue.Fields.Description = json.RawMessage(adfJSON)

		expectedText := "Hello \nWorld!\nSecond paragraph."
		assert.Equal(t, expectedText, issue.Description())
	})
}

func TestIssueService_Get(t *testing.T) {
	client := NewClient(NewJiraCreds("test@example.com", "testorg", "mock-token"))
	service := NewIssueService(client)

	issueKey := "PROJ-1"
	expectedURL := "https://testorg.atlassian.net/rest/api/3/issue/PROJ-1"

	validJSONResp := `{
		"key": "PROJ-1",
		"fields": {
			"summary": "Sample Issue",
			"status": {"name": "Done"}
		}
	}`

	tests := []struct {
		name           string
		mockStatusCode int
		mockRespBody   string
		setupMock      bool
		expectedErr    error
		validateResult func(*testing.T, *Issue)
	}{
		{
			name:           "success 200",
			mockStatusCode: http.StatusOK,
			mockRespBody:   validJSONResp,
			setupMock:      true,
			expectedErr:    nil,
			validateResult: func(t *testing.T, i *Issue) {
				require.NotNil(t, i)
				assert.Equal(t, "PROJ-1", i.Key)
				assert.Equal(t, "Sample Issue", i.Summary())
				assert.Equal(t, "Done", i.Status())
			},
		},
		{
			name:           "unauthorized 401",
			mockStatusCode: http.StatusUnauthorized,
			mockRespBody:   `{"error": "unauthorized"}`,
			setupMock:      true,
			expectedErr:    ErrUnauthorizedRequest,
		},
		{
			name:           "invalid json response",
			mockStatusCode: http.StatusOK,
			mockRespBody:   `{invalid-json}`,
			setupMock:      true,
			expectedErr:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(client.httpClient)
			defer httpmock.DeactivateAndReset()

			if tt.setupMock {
				httpmock.RegisterResponder(http.MethodGet, expectedURL,
					httpmock.NewStringResponder(tt.mockStatusCode, tt.mockRespBody))
			}

			issue, err := service.Get(context.Background(), issueKey)

			if tt.expectedErr != nil {
				if errors.Is(tt.expectedErr, assert.AnError) {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, issue)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, issue)
				}
			}
		})
	}
}

func TestIssueService_buildSearchURL(t *testing.T) {
	client := NewClient(NewJiraCreds("test@example.com", "testorg", "mock-token"))
	service := NewIssueService(client)

	t.Run("without page token", func(t *testing.T) {
		urlStr, err := service.buildSearchURL("PROJ", "")
		require.NoError(t, err)

		assert.Contains(t, urlStr, "https://testorg.atlassian.net/rest/api/3/search/jql")
		assert.Contains(t, urlStr, "jql=project%3DPROJ")
		assert.Contains(t, urlStr, "fields=summary%2Cstatus%2Cpriority%2Cassignee%2CissueType")
		assert.Contains(t, urlStr, "maxResults=50")
		assert.NotContains(t, urlStr, "nextPageToken")
	})

	t.Run("with page token", func(t *testing.T) {
		urlStr, err := service.buildSearchURL("PROJ", "token123")
		require.NoError(t, err)
		assert.Contains(t, urlStr, "nextPageToken=token123")
	})
}

func TestIssueService_List(t *testing.T) {
	client := NewClient(NewJiraCreds("test@example.com", "testorg", "mock-token"))
	service := NewIssueService(client)

	baseURL := "https://testorg.atlassian.net/rest/api/3/search/jql"

	page1JSON := `{
		"issues": [{"key": "PROJ-1"}],
		"nextPageToken": "token2"
	}`
	page2JSON := `{
		"issues": [{"key": "PROJ-2"}],
		"nextPageToken": ""
	}`

	t.Run("successful multi-page pagination", func(t *testing.T) {
		httpmock.ActivateNonDefault(client.httpClient)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, baseURL,
			func(req *http.Request) (*http.Response, error) {
				token := req.URL.Query().Get("nextPageToken")
				switch token {
				case "":
					return httpmock.NewStringResponse(http.StatusOK, page1JSON), nil
				case "token2":
					return httpmock.NewStringResponse(http.StatusOK, page2JSON), nil
				}
				return httpmock.NewStringResponse(http.StatusBadRequest, "unexpected token"), nil
			},
		)

		issues, err := service.List(context.Background(), "PROJ", 3)

		require.NoError(t, err)
		require.Len(t, issues, 2)
		assert.Equal(t, "PROJ-1", issues[0].Key)
		assert.Equal(t, "PROJ-2", issues[1].Key)

		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 2, info["GET "+baseURL])
	})

	t.Run("stops at max pages", func(t *testing.T) {
		httpmock.ActivateNonDefault(client.httpClient)
		defer httpmock.DeactivateAndReset()

		infinitePageJSON := `{"issues": [{"key": "PROJ-X"}], "nextPageToken": "infinite"}`
		httpmock.RegisterResponder(http.MethodGet, baseURL,
			httpmock.NewStringResponder(http.StatusOK, infinitePageJSON))

		issues, err := service.List(context.Background(), "PROJ", 1)

		require.NoError(t, err)
		require.Len(t, issues, 1)

		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 1, info["GET "+baseURL])
	})

	t.Run("http error on fetch", func(t *testing.T) {
		httpmock.ActivateNonDefault(client.httpClient)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, baseURL,
			httpmock.NewStringResponder(http.StatusUnauthorized, "{}"))

		issues, err := service.List(context.Background(), "PROJ", 1)

		assert.ErrorIs(t, err, ErrUnauthorizedRequest)
		assert.Nil(t, issues)
	})
}
