# Migration and compatibility

## Migrating to v2

Main now uses `github.com/faustbrian/go-xsd/v2`. Source stays at the repository
root; the new major is published with the root `v2.0.0` Git tag. Existing v1
consumers keep their original v1 imports and selected v1 dependency until they
deliberately migrate. No corrected v1 release is included.

Insert `/v2` before each package suffix, including
`/compile`, `/resolve` and `/validate`. Public APIs exposing XSD values acquire
new Go type identities; libraries exposing them need their own compatibility
decision rather than silently replacing a v1 dependency.

Review finite parser, compiler, validator and resolver allowances against
representative valid inputs. In addition to the parser allowances below,
compiler particle copies, validator bytes and namespace entries, and Memory
and Catalog constructor resources have independent limits. Configure explicit
finite policies for trusted workloads that exceed defaults; do not assume a
larger serialized-byte limit raises other allowances. Pattern class-subtraction
depth has a separate ceiling of 256.

Default compiler/resolver errors and Diagnostic, Location and ParseError
formatting, JSON and logging are redacted. Prefer sentinel/cause classification
with `errors.Is` and `errors.As`; use exported fields, unwrapping and explicit
detailed projections only at trusted inspection boundaries. Consumers relying
on old error text or structured diagnostic output must migrate deliberately.

WSDL currently exposes v1 XSD types and requires a separately released adoption.
XSD's local WSDL-shaped tests do not certify that public consumer migration.

V2 adds independent `ParseOptions.MaxNamespaceEntries`
and `MaxModelBytes` allowances (zero selects 1,000,000 entries and 64 MiB).
Valid schemas exceeding cumulative namespace-copy/declaration or retained
string/copy-work limits now return nil with `xsd.ErrLimitExceeded`; raise the
explicit finite allowance when a trusted schema needs more work. Source bytes,
depth and elements remain separate limits. Compiler users select these policies
through `Limits.MaxParseNamespaceEntries` and `MaxParseModelBytes`, independently
for each root or loaded document rather than the whole graph. Default Builder,
XSTS and WSDL callers keep their existing signatures and finite defaults;
WSDL-owned lower-limit propagation is a separate adoption boundary. XSD v2
publication does not certify that separate WSDL release.

The released module follows stable v2 compatibility and requires the Go
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
