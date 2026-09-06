# Migration and compatibility

The released module follows stable v1 compatibility and requires the Go
version declared in `go.mod`. Incompatible exported API or documented behavior
changes require a new major version. Compatible releases may reject additional
invalid schemas as documented XML Schema rules become enforced.

Review the [specification decision register](specification-decisions.md) before
upgrading whenever parsing, validation, resolution, or diagnostic behavior can
affect compatibility.

When migrating from generic XML decoding, separate schema parsing from
instance validation, assign an absolute system URI, inject a resolver for every
dependency, compile once, and handle structured diagnostics rather than string
matching errors. Pin a module version and run representative schemas through
the conformance matrix before upgrading.
