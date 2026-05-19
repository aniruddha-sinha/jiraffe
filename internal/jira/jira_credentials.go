package jira

import "github.com/aniruddha-sinha/jiraffe/internal/config"

var (
	JiraConfigEmailKey        = "auth.jira.email"
	JiraConfigOrgKey          = "auth.jira.org"
	JiraConfigEncodedTokenKey = "auth.jira.encoded_token" // nolint:gosec // this is a config key and not an actual token
)

type JiraCredentials struct {
	email    string
	org      string
	apiToken string
}

func NewJiraCredentials(email, org, apiToken string) *JiraCredentials {
	return &JiraCredentials{
		email:    email,
		org:      org,
		apiToken: apiToken,
	}
}

func GetStoredEmail() (string, error) {
	return config.Cfg.Get(JiraConfigEmailKey)
}

func GetStoredOrg() (string, error) {
	return config.Cfg.Get(JiraConfigOrgKey)
}

func GetStoredEncodedToken() (string, error) {
	return config.Cfg.Get(JiraConfigEncodedTokenKey)
}
