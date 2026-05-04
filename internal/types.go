package internal

import "net/http"

type UserProfile struct {
	Email string
	Org   string
}

type ProfilePath struct {
	DirPath  string
	FilePath string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
