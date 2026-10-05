package xsd

import (
	"errors"
	"fmt"
)

var (
	// ErrDTDForbidden reports a DTD or other XML directive. XSD documents do
	// not need DTD processing and directives are rejected before compilation.
	ErrDTDForbidden = errors.New("xsd: DTD and XML directives are forbidden")
	// ErrNotSchema reports that the document element is not xs:schema.
	ErrNotSchema = errors.New("xsd: root element is not an XML Schema schema")
	// ErrLimitExceeded reports that an explicit parser, serializer, or compiler
	// bound was exceeded.
	ErrLimitExceeded = errors.New("xsd: resource limit exceeded")
)

// Location identifies an offset in an input resource. Line and Column are
// one-based when known; Offset is a zero-based byte offset.
// Fields are for explicit trusted inspection; supported default output is
// redacted. Invalid fmt verb/type diagnostics can expose raw fields.
type Location struct {
	SystemID string
	Line     int
	Column   int
	Offset   int64
}

// Format redacts caller-supplied resource identity and operational coordinates.
func (Location) Format(state fmt.State, verb rune) {
	if verb == 'q' {
		_, _ = state.Write([]byte(`"xsd: location"`))
		return
	}
	_, _ = state.Write([]byte("xsd: location"))
}

// MarshalJSON returns only the fixed category, not the trusted location fields.
func (Location) MarshalJSON() ([]byte, error) { return []byte(`"xsd: location"`), nil }

// Severity is the stable importance of a validation diagnostic.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Diagnostic is a stable schema or instance finding. Fields are available for
// explicit trusted inspection; supported default output exposes a fixed
// category. Invalid fmt verb/type diagnostics can expose raw fields.
type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
	Path     string
	Location Location
}

// Format redacts every finding field, including arbitrary severity and code.
func (Diagnostic) Format(state fmt.State, verb rune) {
	if verb == 'q' {
		_, _ = state.Write([]byte(`"xsd: diagnostic"`))
		return
	}
	_, _ = state.Write([]byte("xsd: diagnostic"))
}

// MarshalJSON returns only the fixed category, not the trusted finding fields.
func (Diagnostic) MarshalJSON() ([]byte, error) { return []byte(`"xsd: diagnostic"`), nil }

// ParseError retains trusted location and cause information for a parsing
// failure. Default text, supported formatting, and JSON expose a fixed category;
// inspect Location and Err explicitly only inside a trusted boundary.
// Invalid fmt verb/type diagnostics can expose raw fields without calling Format.
type ParseError struct {
	Location Location
	Err      error
}

// Error returns a fixed category, including for nil and zero receivers.
func (*ParseError) Error() string { return "xsd: parse failed" }

// Format keeps pointer and value formatting, including Go-syntax formatting,
// from traversing Location or invoking cause formatters. Quoting is retained.
func (ParseError) Format(state fmt.State, verb rune) {
	if verb == 'q' {
		_, _ = state.Write([]byte(`"xsd: parse failed"`))
		return
	}
	_, _ = state.Write([]byte("xsd: parse failed"))
}

// MarshalJSON keeps pointer and value encoding from traversing trusted fields.
// As usual, encoding/json encodes a nil pointer as null without calling it.
func (ParseError) MarshalJSON() ([]byte, error) {
	return []byte(`"xsd: parse failed"`), nil
}

// Unwrap supports errors.Is and errors.As.
func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
