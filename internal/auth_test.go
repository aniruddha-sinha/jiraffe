package internal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestGetWebToken_Success(t *testing.T) {
	// 1. MOCK TERMINAL INPUT
	// Swap out the real terminal reader for a fake one that returns a dummy token instantly
	originalPasswordReader := readPasswordFunc
	readPasswordFunc = func(fd int) ([]byte, error) {
		return []byte("dummy_secret_token"), nil
	}
	t.Cleanup(func() {
		// Restore the real terminal reader after the test
		readPasswordFunc = originalPasswordReader
	})

	// 2. MOCK FILE SYSTEM (for Save)
	tempDir := t.TempDir()
	mockUserConfigDir(t, tempDir)

	// 3. MOCK NETWORK (for IsTokenValid)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Simulate successful Jira auth
	}))
	defer mockServer.Close()

	viper.Set("jira.base_url", mockServer.URL)
	t.Cleanup(func() {
		viper.Reset()
	})

	// 4. EXECUTE
	p := UserProfile{Email: "user@test.com", Org: "testorg"}
	valid, err := getWebToken(p)
	// 5. ASSERTIONS
	if err != nil {
		t.Fatalf("getWebToken() returned unexpected error: %v", err)
	}
	if !valid {
		t.Errorf("Expected token to be valid, got false")
	}

	// Verify that Save() actually wrote the file to our temporary directory
	expectedConfigPath := tempDir + "/jiraffe/config.yaml" // Note: adjust if mockUserConfigDir behaves differently
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be saved, but it wasn't found at %s", expectedConfigPath)
	}
}

func TestGetLocalToken_Success(t *testing.T) {
	// 1. MOCK FILE SYSTEM (for loadProfileConfig)
	tempDir := t.TempDir()
	mockUserConfigDir(t, tempDir)

	// Create a fake config file so loadProfileConfig() succeeds
	paths, err := getPaths()
	if err != nil {
		t.Fatalf("Failed to get paths during test setup: %v", err)
	}

	if err := os.MkdirAll(paths.DirPath, 0o700); err != nil {
		t.Fatalf("Failed to create mock directory: %v", err)
	}

	dummyYAML := []byte("auth:\n  encoded_token: fake_encoded_token\n")
	if err := os.WriteFile(paths.FilePath, dummyYAML, 0o600); err != nil {
		t.Fatalf("Failed to write mock config file: %v", err)
	}

	// 2. MOCK NETWORK (for IsTokenValid)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	viper.Set("jira.base_url", mockServer.URL)
	t.Cleanup(func() {
		viper.Reset()
	})

	// 3. EXECUTE
	p := UserProfile{Email: "user@test.com", Org: "testorg"}
	valid, err := getLocalToken(p)
	// 4. ASSERTIONS
	if err != nil {
		t.Fatalf("getLocalToken() returned unexpected error: %v", err)
	}
	if !valid {
		t.Errorf("Expected token to be valid, got false")
	}
}
