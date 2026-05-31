package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// the structs are complex; there are too many fields; many of which are not needed for a CLI application
// I have obtained these structs from gemini ; omitted many of the fields; it works!!
type Issue struct {
	Expand string       `json:"expand,omitempty"`
	ID     string       `json:"id,omitempty"`
	Self   string       `json:"self,omitempty"`
	Key    string       `json:"key,omitempty"`
	Fields *IssueFields `json:"fields,omitempty"`
}

type IssueFields struct {
	Type        IssueType       `json:"issuetype,omitempty"`
	Project     Project         `json:"project,omitempty"`
	Summary     string          `json:"summary,omitempty"`
	Description json.RawMessage `json:"description,omitempty"`

	Assignee *User `json:"assignee,omitempty"`
	Reporter *User `json:"reporter,omitempty"`
	Creator  *User `json:"creator,omitempty"`

	Status     *Status     `json:"status,omitempty"`
	Priority   *Priority   `json:"priority,omitempty"`
	Resolution *Resolution `json:"resolution,omitempty"`

	Labels       []string `json:"labels,omitempty"`
	TimeSpent    int      `json:"timespent,omitempty"`
	TimeEstimate int      `json:"timeestimate,omitempty"`

	Comments   *Comments    `json:"comment,omitempty"`
	IssueLinks []*IssueLink `json:"issuelinks,omitempty"`
	Worklog    *Worklog     `json:"worklog,omitempty"`
	Subtasks   []*Subtasks  `json:"subtasks,omitempty"`

	Unknowns map[string]interface{} `json:"-"`
}

type ADFNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text,omitempty"`
	Content []ADFNode `json:"content,omitempty"`
}

type IssueType struct {
	Self           string `json:"self,omitempty"`
	ID             string `json:"id,omitempty"`
	Description    string `json:"description,omitempty"`
	IconURL        string `json:"iconUrl,omitempty"`
	Name           string `json:"name,omitempty"`
	Subtask        bool   `json:"subtask,omitempty"`
	AvatarID       int    `json:"avatarId,omitempty"`
	HierarchyLevel int    `json:"hierarchyLevel,omitempty"`
}

type User struct {
	Self         string `json:"self,omitempty"`
	Name         string `json:"name,omitempty"`
	Key          string `json:"key,omitempty"`
	AccountID    string `json:"accountId,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	TimeZone     string `json:"timeZone,omitempty"`
	AccountType  string `json:"accountType,omitempty"`
}

type Status struct {
	Self           string         `json:"self,omitempty"`
	Description    string         `json:"description,omitempty"`
	IconURL        string         `json:"iconUrl,omitempty"`
	Name           string         `json:"name,omitempty"`
	ID             string         `json:"id,omitempty"`
	StatusCategory StatusCategory `json:"statusCategory,omitempty"`
}

type StatusCategory struct {
	Self      string `json:"self,omitempty"`
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Key       string `json:"key,omitempty"`
	ColorName string `json:"colorName,omitempty"`
}

type Priority struct {
	Self        string `json:"self,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Name        string `json:"name,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}

type Resolution struct {
	Self        string `json:"self,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
}

type Comments struct {
	Comments   []*Comment `json:"comments,omitempty"`
	MaxResults int        `json:"maxResults,omitempty"`
	Total      int        `json:"total,omitempty"`
	StartAt    int        `json:"startAt,omitempty"`
}

type Comment struct {
	ID           string             `json:"id,omitempty"`
	Self         string             `json:"self,omitempty"`
	Author       *User              `json:"author,omitempty"`
	Body         json.RawMessage    `json:"body,omitempty"`
	UpdateAuthor *User              `json:"updateAuthor,omitempty"`
	Created      string             `json:"created,omitempty"`
	Updated      string             `json:"updated,omitempty"`
	Visibility   *CommentVisibility `json:"visibility,omitempty"`
}

type CommentVisibility struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type IssueLink struct {
	ID           string        `json:"id,omitempty"`
	Self         string        `json:"self,omitempty"`
	Type         IssueLinkType `json:"type,omitempty"`
	OutwardIssue *Issue        `json:"outwardIssue,omitempty"`
	InwardIssue  *Issue        `json:"inwardIssue,omitempty"`
	Comment      *Comment      `json:"comment,omitempty"`
}

type IssueLinkType struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Inward  string `json:"inward,omitempty"`
	Outward string `json:"outward,omitempty"`
	Self    string `json:"self,omitempty"`
}

type Worklog struct {
	StartAt    int             `json:"startAt,omitempty"`
	MaxResults int             `json:"maxResults,omitempty"`
	Total      int             `json:"total,omitempty"`
	Worklogs   []WorklogRecord `json:"worklogs,omitempty"`
}

type WorklogRecord struct {
	ID               string `json:"id,omitempty"`
	Self             string `json:"self,omitempty"`
	IssueID          string `json:"issueId,omitempty"`
	Author           *User  `json:"author,omitempty"`
	UpdateAuthor     *User  `json:"updateAuthor,omitempty"`
	Comment          string `json:"comment,omitempty"`
	TimeSpent        string `json:"timeSpent,omitempty"`
	TimeSpentSeconds int    `json:"timeSpentSeconds,omitempty"`
}

type Subtasks struct {
	ID     string       `json:"id,omitempty"`
	Key    string       `json:"key,omitempty"`
	Self   string       `json:"self,omitempty"`
	Fields *IssueFields `json:"fields,omitempty"`
}

type SearchResult struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

type IssueService struct {
	issueClient *Client
}

func NewIssueService(client *Client) *IssueService {
	return &IssueService{issueClient: client}
}

func (is *IssueService) List(ctx context.Context, projectKey string) ([]Issue, error) {
	fullURL, err := is.issueClient.buildURLForQueryParams(urlTemplateSearchAPI, apiVersion)
	if err != nil {
		return nil, err
	}

	query := fullURL.Query()
	query.Add("jql", fmt.Sprintf(`project="%s"`, projectKey))
	query.Add("fields", "summary,key")
	fullURL.RawQuery = query.Encode()

	request, err := is.issueClient.NewRequest(ctx, http.MethodGet, fullURL.String())
	if err != nil {
		return nil, err
	}

	response, err := is.issueClient.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if err := mapStatusToError(response.StatusCode); err != nil {
		return nil, err
	}

	var searchResult SearchResult
	if err := json.NewDecoder(response.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResult.Issues, nil
}

func (is *IssueService) Get(ctx context.Context, issueKey string) (*Issue, error) {
	fullURL, err := is.issueClient.buildRawURL(urlTemplateIssueGet, apiVersion, issueKey)
	if err != nil {
		return &Issue{}, err
	}

	request, err := is.issueClient.NewRequest(ctx, http.MethodGet, fullURL)
	if err != nil {
		return &Issue{}, err
	}

	response, err := is.issueClient.httpClient.Do(request)
	if err != nil {
		return &Issue{}, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if err := mapStatusToError(response.StatusCode); err != nil {
		return nil, err
	}

	var issue Issue
	if err := json.NewDecoder(response.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue %s: %w", issueKey, err)
	}

	return &issue, nil
}

// from this point on I have taken the help of Gemini to get my printing of so many of the Issue fields sorted; too much boiler plate
// might consider a library of some kind
func (issue *Issue) PrintIssueDetail() {
	if issue == nil {
		fmt.Println("Error: Issue is nil")
		return
	}
	if issue.Fields == nil {
		fmt.Printf("Issue %s: No fields populated.\n", issue.Key)
		return
	}

	f := issue.Fields // Shortcut for readability

	// --- Header ---
	fmt.Printf("\n=== %s: %s ===\n", issue.Key, f.Summary)

	// --- Core Metadata (with inline nil-checks) ---
	status := "Unknown"
	if f.Status != nil {
		status = f.Status.Name
	}

	priority := "None"
	if f.Priority != nil {
		priority = f.Priority.Name
	}

	assignee := "Unassigned"
	if f.Assignee != nil {
		assignee = f.Assignee.DisplayName
	}

	reporter := "Unknown"
	if f.Reporter != nil {
		reporter = f.Reporter.DisplayName
	}

	// Print metadata block, aligning columns with spaces for readability
	fmt.Printf("Type:     %-15s Status:   %s\n", f.Type.Name, status)
	fmt.Printf("Priority: %-15s Assignee: %s\n", priority, assignee)
	fmt.Printf("Reporter: %-15s Labels:   %s\n", reporter, formatLabels(f.Labels))
	fmt.Println(strings.Repeat("-", 50))

	if len(f.Subtasks) > 0 {
		fmt.Printf("Subtasks (%d): ", len(f.Subtasks))
		var subKeys []string
		for _, st := range f.Subtasks {
			subKeys = append(subKeys, st.Key)
		}
		fmt.Println(strings.Join(subKeys, ", "))
	}

	if f.Comments != nil && len(f.Comments.Comments) > 0 {
		fmt.Printf("Comments: %d\n", f.Comments.Total) // Relying on the pagination total
	}

	if f.Worklog != nil && f.Worklog.Total > 0 {
		// Convert total seconds to hours
		fmt.Printf("Time Logged: %.1f hours\n", float64(f.TimeSpent)/3600.0)
	}

	fmt.Println(strings.Repeat("-", 50))

	// --- Comments ---
	if f.Comments != nil && len(f.Comments.Comments) > 0 {
		fmt.Printf("\nComments (%d):\n", f.Comments.Total)

		for i, c := range f.Comments.Comments {
			author := "Unknown"
			if c.Author != nil {
				author = c.Author.DisplayName
			}

			// Re-use your extractor!
			commentText := ExtractPlaintextFromADF(c.Body)

			fmt.Printf("--- %s (%s) ---\n", author, c.Created)
			fmt.Println(commentText)

			// Cap it at showing 3 comments so it doesn't flood the terminal
			if i >= 2 {
				fmt.Printf("\n... and %d more comments not shown.\n", f.Comments.Total-3)
				break
			}
		}
	}

	// --- Description ---
	fmt.Println("\nDescription:")
	if len(f.Description) > 0 && string(f.Description) != "null" {

		// Parse the ADF into a plain string!
		plainText := ExtractPlaintextFromADF([]byte(f.Description))

		// Truncate massively long descriptions for CLI sanity
		if len(plainText) > 800 {
			fmt.Println(plainText[:800] + "\n... (truncated)")
		} else {
			fmt.Println(plainText)
		}
	} else {
		fmt.Println("*No description provided*")
	}
	fmt.Println()
}

// Helper for formatting string slices
func formatLabels(labels []string) string {
	if len(labels) == 0 {
		return "None"
	}
	return strings.Join(labels, ", ")
}

func ExtractPlaintextFromADF(raw []byte) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var root ADFNode
	if err := json.Unmarshal(raw, &root); err != nil {
		return "[Failed to parse ADF description]"
	}

	var builder strings.Builder
	extractTextRec(root, &builder)

	return strings.TrimSpace(builder.String())
}

func extractTextRec(node ADFNode, builder *strings.Builder) {
	if node.Type == "text" && node.Text != "" {
		builder.WriteString(node.Text)
	}

	if node.Type == "paragraph" || node.Type == "listItem" || strings.HasPrefix(node.Type, "heading") {
		builder.WriteString("\n")
	}

	if node.Type == "listItem" {
		builder.WriteString("• ")
	}

	for _, child := range node.Content {
		extractTextRec(child, builder)
	}
}
