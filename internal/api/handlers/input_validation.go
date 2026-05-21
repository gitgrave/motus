package handlers

import (
	"errors"
	"strings"
)

const (
	maxDisplayNameLen = 200
	maxDescriptionLen = 2000
)

// ValidateDisplayName checks that a display name (geofence name, calendar
// name, etc.) is within length limits and contains no HTML-injectable or
// control characters. An empty string is accepted (callers check presence
// separately with a "required" guard).
func ValidateDisplayName(name string) error {
	if len(name) > maxDisplayNameLen {
		return errors.New("name exceeds maximum length")
	}
	return validateTextChars(name)
}

// ValidateDescription checks that a description field is within length
// limits and contains no HTML-injectable or control characters.
func ValidateDescription(desc string) error {
	if len(desc) > maxDescriptionLen {
		return errors.New("description exceeds maximum length")
	}
	return validateTextChars(desc)
}

// validateTextChars rejects strings containing angle brackets (< >) or
// control characters other than newline (\n) and horizontal tab (\t).
func validateTextChars(s string) error {
	if strings.ContainsAny(s, "<>") {
		return errors.New("value contains invalid characters")
	}
	for _, r := range s {
		if r < 0x20 && r != '\n' && r != '\t' {
			return errors.New("value contains invalid characters")
		}
	}
	return nil
}
