package hipaa

import (
	"regexp"
	"strings"
)

// Detector is responsible for identifying Protected Health Information (PHI) in data.
// As a healthcare professional, I understand the critical importance of safeguarding patient data.
// This detector uses regular expressions to scan for common PHI patterns like Social Security Numbers,
// email addresses, phone numbers, and medical terms that could indicate sensitive information.
type Detector struct {
	patterns []*regexp.Regexp
}

// NewDetector creates a new PHI detector with predefined patterns.
// These patterns are based on common PHI identifiers that healthcare providers must protect.
// The list is configurable in the sense that it can be extended, but starts with essential ones.
func NewDetector() *Detector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), // SSN
		regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`), // Email
		regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b`), // US Phone
		regexp.MustCompile(`\b\d{10}\b`), // Phone without dashes
		regexp.MustCompile(`(?i)\b(patient|diagnosis|treatment|medication|symptom)\b`), // Medical terms
		regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`), // Date of birth format
	}
	return &Detector{patterns: patterns}
}

// ContainsPHI checks if the given data contains any PHI patterns.
// In clinical practice, we must be vigilant about any data that could be linked to a patient.
// This function returns true if PHI is detected, allowing for appropriate handling.
func (d *Detector) ContainsPHI(data string) bool {
	for _, pattern := range d.patterns {
		if pattern.MatchString(data) {
			return true
		}
	}
	return false
}

// SanitizePHI removes or redacts PHI from the data.
// When dealing with patient information, it's better to err on the side of caution.
// This replaces detected PHI with placeholders to prevent accidental exposure.
func (d *Detector) SanitizePHI(data string) string {
	sanitized := data
	for _, pattern := range d.patterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			return strings.Repeat("*", len(match))
		})
	}
	return sanitized
}