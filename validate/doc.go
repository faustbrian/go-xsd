// Package validate checks XML instances against immutable compiled XSD sets.
// New selects finite independent limits; Validate, ValidateReader and
// ValidateTree use the same validation engine with per-call state. Returned
// diagnostics are caller-owned and retain trusted detailed fields while their
// supported default formatting is redacted.
//
// Use a non-nil context and initialized Validator. Calls may run concurrently;
// input bytes/trees must remain unchanged while borrowed. Tree validation
// clones the supplied structure under work limits. Readers are not closed by
// the validator and must cooperate with cancellation of blocking reads. XML
// directives are forbidden; schema hints do not authorize file/network access.
// Resource refusal or cancellation returns an error, not a partial valid result.
package validate
