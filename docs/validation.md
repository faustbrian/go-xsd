# Validation

Create a `validate.Validator` from an immutable compiled set. `Validate`
accepts caller-provided XML bytes, `ValidateReader` incrementally reads an
`io.Reader`, and `ValidateTree` accepts a caller-owned expanded-name tree.
All three entry points share the same validation engine and deterministic
diagnostics.

Reader validation does not require the caller to buffer the complete XML
instance. Parsing still builds a bounded internal tree: byte, depth, node,
attribute, text, diagnostic, XPath, and identity-value limits remain in
force. Context cancellation and reader errors propagate to the caller, and
DTDs remain forbidden.

`MaxBytes` independently bounds serialized input and cumulative validator-owned
string payload occurrences: expanded node/attribute names, attribute values,
namespace prefixes/URIs, retained text (including recopied text prefixes), and
location system IDs. This is conservative cumulative retention/copy work, not
final retained bytes or an exact heap-size limit: segmented text recopies its
previous prefix and can require a larger allowance than one text token.
`MaxTextBytes` remains
an independent bound on instance text. Tree admission checks known attribute
counts and direct-child node lower bounds before allocating clone capacity.

`MaxNamespaceEntries` independently bounds cumulative namespace work; zero
selects 1,000,000 entries and negative values are invalid. Each copied scope
entry and each declaration insertion, including a rebinding, counts once.
Declarations are not ordinary attributes. Reader scopes inherit their parents;
each tree node must supply its complete explicit scope and does not inherit one.
Limits apply before owned retention, not before the standard XML decoder's
token allocations. Refusal returns a resource error and no partial result,
without modifying the caller's tree or performing implicit I/O.

Diagnostics contain severity, stable code, message, instance path, system ID,
line, column, and byte offset. A validation result may contain multiple schema
errors. Resource-limit or parsing failures are returned as Go errors.

The validator covers the features identified in the requirement matrix,
including supported simple types and facets, particles, wildcards,
substitution groups, anonymous types, and the implemented identity XPath
subset. Complete XSD assessment semantics remain a release blocker.
The published support claim is therefore limited to the versioned requirement
matrix and decision register; XSTS and peer results do not widen that boundary.
