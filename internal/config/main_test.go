package config_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// All LoadFromEnv tests need a non-empty database password since there
	// is no default. Tests that specifically test password validation clear
	// this via t.Setenv within the test itself.
	if os.Getenv("MOTUS_DATABASE_PASSWORD") == "" {
		_ = os.Setenv("MOTUS_DATABASE_PASSWORD", "testpassword")
		defer func() { _ = os.Unsetenv("MOTUS_DATABASE_PASSWORD") }()
	}
	os.Exit(m.Run())
}
