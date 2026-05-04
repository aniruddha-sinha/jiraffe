package internal

import (
	"os"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

func mockUserConfigDir(t *testing.T, tempDir string) {
	t.Helper()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", tempDir)
	case "darwin":
		t.Setenv("HOME", tempDir) // macOS appends "/Library/Application Support" to HOME
	default: // Linux and others
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir) // Fallback if XDG isn't used
	}
}

func TestSave(t *testing.T) {
	// 1. Create a temporary sandbox for the file system
	tempDir := t.TempDir()

	// 2. Trick the application into using our sandbox
	mockUserConfigDir(t, tempDir)

	// 3. Ensure viper is completely reset after this test finishes
	t.Cleanup(func() {
		viper.Reset()
	})

	// 4. Set up test data
	// (Assuming UserProfile looks something like this based on your code)
	profile := UserProfile{
		Email: "test@jiraffe.com",
		Org:   "my-test-org",
	}
	token := "dummy_encoded_token_123"

	// 5. Execute the function
	filePath, err := Save(profile, token)
	// 6. Assertions
	if err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	// Verify the directory and file were actually created on disk
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected config file to exist at %s, but it does not", filePath)
	}

	// Verify the contents of the file were written correctly.
	// We create a fresh viper instance just to read the file independently.
	verifyViper := viper.New()
	verifyViper.SetConfigFile(filePath)
	if err := verifyViper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read the newly created config file: %v", err)
	}

	if got := verifyViper.GetString("auth.email"); got != profile.Email {
		t.Errorf("Expected email %s, got %s", profile.Email, got)
	}
	if got := verifyViper.GetString("auth.org"); got != profile.Org {
		t.Errorf("Expected org %s, got %s", profile.Org, got)
	}
	if got := verifyViper.GetString("auth.encoded_token"); got != token {
		t.Errorf("Expected token %s, got %s", token, got)
	}
}

func TestLoadProfileConfig(t *testing.T) {
	tempDir := t.TempDir()
	mockUserConfigDir(t, tempDir)

	t.Cleanup(func() {
		viper.Reset()
	})

	// 1. Manually create a dummy configuration file in our sandbox
	paths, err := getPaths()
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(paths.DirPath, 0o700); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Write some dummy yaml data
	dummyYAML := []byte("auth:\n  email: user@test.com\n  org: testorg\n")
	if err := os.WriteFile(paths.FilePath, dummyYAML, 0o600); err != nil {
		t.Fatalf("Failed to write dummy config: %v", err)
	}

	// 2. Execute the function being tested
	err = loadProfileConfig()
	// 3. Assertions
	if err != nil {
		t.Fatalf("loadProfileConfig() returned unexpected error: %v", err)
	}

	// Verify viper correctly loaded the data into memory
	if got := viper.GetString("auth.email"); got != "user@test.com" {
		t.Errorf("Expected viper to load email 'user@test.com', got '%s'", got)
	}
}
