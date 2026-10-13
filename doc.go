// Package xsd parses XML Schema 1.0 Second Edition documents into a mutable
// model and serializes that model deterministically, not as canonical XML or a
// byte-for-byte reproduction of the source. The pinned specification and live
// support matrix are in specification/; no XSD 1.1 support is implied.
//
// Parse requires a non-nil context, forbids XML directives/DTD processing and
// performs no external resolution. Finite source, depth, element, namespace
// copy and retained-string budgets are independent; they do not bound decoder
// token allocations or total process memory. Compile and resolve own explicit
// schema composition; validate owns instance checking against a compiled set.
//
// Document and nested model slices, maps and pointers are caller-mutable and
// need caller synchronization while parsing, serializing or compiling related
// data. Marshal checks cycles and structural budgets before serialization but
// does not replace schema compilation. Supported default diagnostic formatting
// is redacted; exported fields and unwrapped causes are trusted-inspection
// seams, not safe log projections. See docs/security.md for the exact boundary.
package xsd
