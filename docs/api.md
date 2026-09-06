# API guide

The Go package documentation is authoritative for signatures. This guide maps
package responsibility and ownership boundaries.

| Package | Use |
| --- | --- |
| root | Parse, model, and deterministically serialize bounded XML Schema 1.0 documents. |
| `builder` | Construct checked schema documents through a mutable builder and explicit `Build` boundary. |
| `compile` | Resolve and compile bounded schema graphs into immutable sets safe for concurrent validation. |
| `datatype` | Validate, compare, and canonicalize XML Schema lexical and value spaces. |
| `resolve` | Supply explicit deny-by-default, in-memory, catalog, chained, or root-confined file resolution. |
| `validate` | Validate bounded XML instances or caller-owned trees against an immutable compiled set. |
| `xsdtest` | Execute the pinned XML Schema Test Suite and inspect attributable conformance results in tests. |

Parsing and validation perform no implicit I/O. A compiler borrows its
configured resolver and returns an immutable set. Validation results own their
diagnostics. Callers retain ownership of contexts, source bytes, readers, and
injected resolver resources.

Use the [compiler-checked package example](../example_test.go) for the smallest
complete parse, compile, and validation flow. The [architecture guide](architecture.md)
describes compilation in more detail, and the [security guide](security.md)
defines required limits and resolver hardening.
