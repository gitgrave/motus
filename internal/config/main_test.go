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
	// LoadFromEnv tests also need a valid CSRF secret in production mode.
	// Tests that specifically test CSRF validation override this via t.Setenv.
	if os.Getenv("MOTUS_CSRF_SECRET") == "" {
		_ = os.Setenv("MOTUS_CSRF_SECRET", "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
		defer func() { _ = os.Unsetenv("MOTUS_CSRF_SECRET") }()
	}
	os.Exit(m.Run())
}
