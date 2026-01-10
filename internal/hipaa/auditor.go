package hipaa

import (
	"fmt"
	"time"
)

// Auditor logs HIPAA compliance events for auditing purposes.
// Just as we maintain detailed patient records in medicine, auditing ensures accountability.
// This auditor records events like PHI detection, encryption, and access attempts.
type Auditor struct {
	enabled bool
}

// NewAuditor creates a new auditor.
// In clinical settings, audit logs help track who accessed what and when.
// Set enabled to true to start logging events.
func NewAuditor(enabled bool) *Auditor {
	return &Auditor{enabled: enabled}
}

// LogEvent logs a HIPAA compliance event.
// Events include timestamps for traceability, similar to medical timestamps.
// This helps in reviewing compliance and identifying potential breaches.
func (a *Auditor) LogEvent(eventType, details string) {
	if !a.enabled {
		return
	}
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("[HIPAA AUDIT %s] %s: %s\n", timestamp, eventType, details)
}

// LogPHIDetected logs when PHI is detected.
// Detecting PHI is the first step in protecting patient privacy.
// This event alerts to potential risks.
func (a *Auditor) LogPHIDetected(location, dataSnippet string) {
	a.LogEvent("PHI_DETECTED", fmt.Sprintf("PHI detected at %s: %s", location, dataSnippet))
}

// LogDataEncrypted logs when data is encrypted.
// Encryption is our shield against unauthorized access, much like locked cabinets for records.
// This confirms protective measures are in place.
func (a *Auditor) LogDataEncrypted(location string) {
	a.LogEvent("DATA_ENCRYPTED", fmt.Sprintf("Data encrypted at %s", location))
}

// LogAccessAttempt logs attempts to access sensitive data.
// Monitoring access is crucial for security, just as we track patient visits.
// This helps in forensic analysis if needed.
func (a *Auditor) LogAccessAttempt(location, action string) {
	a.LogEvent("ACCESS_ATTEMPT", fmt.Sprintf("Access attempt at %s for %s", location, action))
}