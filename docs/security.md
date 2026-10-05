# Security and limits

For the next major, `compile.Compiler.Compile` failures and built-in resolver
constructor/Resolve/Close failures use `xsd compile: failed` or
`xsd resolve: failed` for supported text/formatting, JSON and logging. Default
output never evaluates delegated error formatting, marshaling or logging hooks.
Exact standard `context.Canceled`/`context.DeadlineExceeded` remain unchanged;
other causes are wrapped without evaluating their text. `Compile.New` already
uses payload-free option errors and is unchanged.

The original cause remains available through explicit trusted unwrapping,
`errors.Is` and `errors.As`. Resolver chains still advance only on not-found
classification; access/operational errors stop. Error returns publish no partial
Set/Resource, including a partial result returned by a delegated resolver.
File-open not-found and operational errors retain both their resolver
classification and underlying OS cause. This does not retroactively restore
causes previously flattened by unrelated internal compiler formatting.
Explicit schema data, XSTS field projections and detailed reporting remain
trusted seams; invalid Go fmt verb/type diagnostics remain outside default
output protection as described below.

XML Schema pattern translation independently bounds source bytes (1 MiB),
translated bytes (8 MiB), and class-subtraction nesting (inclusive depth 256,
outer class depth one). Depth refusal occurs before descending into an excluded
class. `datatype.CompilePatternContext` cooperatively checks its context during
owned translation, class and rune-set work and before/after standard-library
compilation, preserving cancellation/deadline causes and returning no partial
regexp. Standard-library sorting/regexp compilation and matching are synchronous
bounded operations, not interruptible work. `CompilePattern` uses a background
context; compiler and validator input patterns use their actual operation context.

For the next major, `Diagnostic` and `Location` supported default formatting,
JSON and logging return only `xsd: diagnostic` and `xsd: location`. Quoted
formatting quotes the category; nil pointers retain safe nil/null forms.
All fields, including arbitrary severity/code strings and source coordinates,
remain available for explicit trusted inspection and detailed report projection.
Schema-wire location strings and the XSTS harness's explicit reports are unchanged.

This protection covers supported/default output, not invalid Go format
verb/type combinations. Go 1.27 processes `%p` on a struct value and invalid
`%w` through error diagnostics that bypass formatting methods and expose fields.
`%T` and `%p` on pointers retain intrinsic type/address output. Trusted callers
must use supported verbs, or format an explicitly redacted string before passing
it to arbitrary format strings; the same limitation applies to `ParseError`.

Instance ingestion bounds cumulative owned string payload and namespace scope
copy/declaration work independently of serialized input, text, nodes and
ordinary attributes. Tree clone cardinalities are admitted before owned map and
child-capacity allocation. These policies do not bound allocations performed by
the standard XML decoder before it yields a token; caller readers still own
their blocking-read cancellation. See [validation](validation.md) for accounting
and explicit tree namespace-scope semantics.

The parser and instance validator forbid DTD directives and do not expand
external entities. Parsing performs no implicit I/O. Compilation denies file
and remote resolution unless the caller injects a resolver.

For the next major, non-nil `ParseError` default `Error`, supported pointer/value `fmt` formats
(including Go-syntax formats), JSON and `slog` output expose only the fixed
`xsd: parse failed` category without evaluating cause formatting, marshaling
or logging callbacks. Nil and zero receivers return that category from
`Error`; nil `Unwrap` returns nil, `fmt` uses its safe `<nil>` placeholder for
a nil pointer, and JSON encodes a nil pointer as `null`.
Its exported `Location` and `Err` remain unchanged for explicit trusted
inspection through fields, `Unwrap`, `errors.Is` and `errors.As`. Do not log
those fields outside that trusted boundary. This is an intentional formatting
compatibility break, not a release announcement. Public compiler and built-in
resolver error boundaries have the separate cause-preserving protections above.

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
