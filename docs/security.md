# Security and limits

The parser and instance validator forbid DTD directives and do not expand
external entities. Parsing performs no implicit I/O. Compilation denies file
and remote resolution unless the caller injects a resolver.

For the next major, non-nil `ParseError` default `Error`, pointer/value `fmt` formats
(including Go-syntax formats), JSON and `slog` output expose only the fixed
`xsd: parse failed` category without evaluating cause formatting, marshaling
or logging callbacks. Nil and zero receivers return that category from
`Error`; nil `Unwrap` returns nil, `fmt` uses its safe `<nil>` placeholder for
a nil pointer, and JSON encodes a nil pointer as `null`.
Its exported `Location` and `Err` remain unchanged for explicit trusted
inspection through fields, `Unwrap`, `errors.Is` and `errors.As`. Do not log
those fields outside that trusted boundary. This is an intentional formatting
compatibility break, not a release announcement. It does not make standalone
`Location`, validation `Diagnostic`, compiler or resolver error output private.

Parser options bound bytes, XML element depth, and total elements before the
document model is built. Compiler options bound schema bytes, graph depth,
documents, references, components, and particles. Validation additionally
bounds nodes, attributes, text, diagnostics, identity values, and estimated
identity XPath steps. Keep limits finite when processing tenant or
internet-controlled documents.

Serialization preflights the complete in-memory model before writing. It
rejects cyclic models and bounds structural depth, component work, retained
output memory, and emitted bytes. `xsd.Marshal` uses conservative defaults;
use `xsd.MarshalWithOptions` to lower `MaxDepth`, `MaxComponents`, or
`MaxOutputBytes` at a trust boundary.

`make check` includes hostile-input and filesystem tests for implicit network access, file
and symlink escape, DTD and entity input, deep XML and schema models, recursive
and explosive particles, regex translation, identity XPath amplification, and
diagnostic growth.

An injected remote resolver remains part of the application's trust boundary.
It must defend against SSRF, redirects, DNS rebinding, credential forwarding,
and decompression bombs. The opt-in `resolve.File` resolver confines opens to
an explicit root and applies a per-resource byte limit. The package does not
make other injected resolvers safe.
