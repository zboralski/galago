package hipaa

// Sanitizer handles the removal or redaction of PHI from data.
// In patient care, we often anonymize data for research or sharing.
// This sanitizer uses the detector to identify and redact PHI.
type Sanitizer struct {
	detector *Detector
}

// NewSanitizer creates a new sanitizer with the given detector.
// Combining detection and sanitization ensures comprehensive protection.
// It's like a multi-step process in medical diagnosis and treatment.
func NewSanitizer(detector *Detector) *Sanitizer {
	return &Sanitizer{detector: detector}
}

// Sanitize removes PHI from the given data.
// Redaction prevents accidental disclosure, preserving utility while protecting privacy.
// This is essential when handling data that might contain patient information.
func (s *Sanitizer) Sanitize(data string) string {
	return s.detector.SanitizePHI(data)
}

// IsSafe checks if the data is free of PHI.
// Before proceeding with data processing, we verify safety.
// This gives confidence that no PHI is present.
func (s *Sanitizer) IsSafe(data string) bool {
	return !s.detector.ContainsPHI(data)
}