# Builders and serialization

[Documentation index](README.md) · [Repository entry](../README.md)

`builder.Schema` offers checked methods for named elements, attributes, simple
restrictions, and complex types. Inputs are deeply copied. `Build` serializes
and compiles the complete result before returning an isolated document, so
unresolved types and invalid content models are reported instead of emitted.

Restriction facets include owned namespace maps and annotation content; changing
the caller's inputs after admission cannot alter a later build. Returned
documents are independently mutable and do not alter the retained builder.
Construction is trusted: recursive content passed to `AddComplexType` must be
finite and acyclic. Copying occurs before `Build`, without a separate clone
budget or context; later serializer/compiler limits do not bound that phase.

`xsd.Marshal` writes deterministic UTF-8 XML. Namespace declarations are
sorted, missing QName prefixes are allocated deterministically, and component
order follows the model's stable category and source order. Parse the output or
compile it before publication when constructing raw model structs directly.
Serialization rejects cyclic raw models and applies finite output, depth, and
component-work limits. `xsd.MarshalWithOptions` allows callers to lower those
limits for untrusted or tenant-controlled models.

The convenience methods intentionally cover a conservative surface. Raw
structs remain available for advanced schemas, but they are not valid by
construction; `Build` is the required validity boundary for either path.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
