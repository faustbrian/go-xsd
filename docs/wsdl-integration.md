# WSDL integration

[Documentation index](README.md) · [Repository entry](../README.md)

WSDL tooling should compile embedded and imported schemas once, retain the
immutable `compile.Set`, and query declarations by expanded QName. It should
not duplicate chameleon include, group expansion, datatype, or substitution
logic.

Provide WSDL and XSD resources through one application-owned resolver so base
URI and access policy stay consistent. Do not infer namespace prefixes from
compiled QNames; prefixes are serialization details, while component identity
is the namespace URI plus local name.

The set accessors expose global elements, attributes, simple and complex types,
model groups, attribute groups, notation declarations, and substitution
membership. Corresponding `*Names` methods provide deterministic expanded-name
inventories so generators do not need access to mutable compiler maps.
These are isolated, resolved component projections, not lossless source models
or original XML bytes. Named groups are expanded in compiled complex content;
retain the parsed source document separately when source-level IDs or lexical
metadata matter.
`compile/wsdl_contract_test.go` fixes this public consumer surface as an
executable contract. Future `wsdl` requirements must extend that contract
and the requirement matrix instead of reaching into compiler internals.

Published WSDL v3.0.1 declares XSD v2.0.0 as its minimum. After the XSD v2.0.2
tag and release are available, explicitly select that XSD patch to adopt the
restriction-builder ownership repair; the WSDL dependency minimum is unchanged.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
