# Migration and compatibility

The next-major candidate adds independent `ParseOptions.MaxNamespaceEntries`
and `MaxModelBytes` allowances (zero selects 1,000,000 entries and 64 MiB).
Valid schemas exceeding cumulative namespace-copy/declaration or retained
string/copy-work limits now return nil with `xsd.ErrLimitExceeded`; raise the
explicit finite allowance when a trusted schema needs more work. Source bytes,
depth and elements remain separate limits. Compiler users select these policies
through `Limits.MaxParseNamespaceEntries` and `MaxParseModelBytes`, independently
for each root or loaded document rather than the whole graph. Default Builder,
XSTS and WSDL callers keep their existing signatures and finite defaults;
WSDL-owned lower-limit propagation is a separate adoption boundary. No new
version or release qualification is implied by this candidate.

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
