# Package map

[Documentation index](README.md) · [Repository entry](../README.md)

This map follows the current source modules, not a synchronized release.
Use the repository README for published installation versions; an unpublished
source major is not installable merely because its package appears here.
Internal packages and non-releasable fixtures are intentionally excluded.

## github.com/faustbrian/go-xsd/v2

`xsd` is a secure XML Schema 1.0 parser, compiler, validator, serializer, and builder for Go. It is intended to provide the schema layer for `wsdl` and SOAP tooling without performing implicit file or network access.

| Import path | Responsibility |
| --- | --- |
| [`github.com/faustbrian/go-xsd/v2`](../doc.go) | Use to parse, model, and deterministically serialize bounded XML Schema 1.0 documents. |
| [`github.com/faustbrian/go-xsd/v2/builder`](../builder/builder.go) | Use to construct checked XML Schema documents through a mutable builder and explicit Build validation boundary. |
| [`github.com/faustbrian/go-xsd/v2/compile`](../compile/doc.go) | Use to resolve and compile bounded schema graphs into immutable concurrent validation sets. |
| [`github.com/faustbrian/go-xsd/v2/datatype`](../datatype/doc.go) | Use for XML Schema datatype lexical validation, canonicalization, ordering, and immutable numeric values. |
| [`github.com/faustbrian/go-xsd/v2/resolve`](../resolve/doc.go) | Use to supply explicit deny-by-default, in-memory, catalog, chained, or root-confined file resolution. |
| [`github.com/faustbrian/go-xsd/v2/validate`](../validate/doc.go) | Use to validate bounded XML instances or caller-owned trees against an immutable compiled schema set. |
| [`github.com/faustbrian/go-xsd/v2/xsdtest`](../xsdtest/xsts.go) | Use in tests to execute the pinned XML Schema Test Suite and inspect attributable conformance results. |

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
