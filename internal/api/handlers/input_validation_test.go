package handlers_test

import (
	"strings"
	"testing"

	"github.com/tamcore/motus/internal/api/handlers"
)

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "My Geofence", false},
		{"valid unicode", "Büro Frankfurt", false},
		{"valid with newline", "Line1\nLine2", false},
		{"valid with tab", "Col1\tCol2", false},
		{"empty string is allowed", "", false},
		{"at max length", strings.Repeat("a", 200), false},
		{"exceeds max length", strings.Repeat("a", 201), true},
		{"contains <", "hello <world>", true},
		{"contains >", "foo > bar", true},
		{"contains NUL", "foo\x00bar", true},
		{"contains control char 0x01", "foo\x01bar", true},
		{"contains control char 0x1F", "foo\x1fbar", true},
		{"script tag", "<script>alert(1)</script>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateDisplayName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDisplayName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "A nice area near the office.", false},
		{"empty string is allowed", "", false},
		{"valid with newlines", "Line1\nLine2\nLine3", false},
		{"at max length", strings.Repeat("a", 2000), false},
		{"exceeds max length", strings.Repeat("a", 2001), true},
		{"contains <", "<b>bold</b>", true},
		{"contains >", "a > b", true},
		{"contains NUL", "foo\x00bar", true},
		{"contains control char", "foo\x02bar", true},
		{"script payload", "<script>alert(1)</script>", true},
		{"img onerror payload", `<img src=x onerror=alert(1)>`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescription(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
