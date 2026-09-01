# Specification provenance

`manifest.tsv` pins XML Schema 1.0 Second Edition, XML 1.0 Fifth Edition,
Namespaces in XML 1.0 Third Edition, supporting schemas, and XSTS. Dated W3C
Recommendation and XSTS URLs are immutable inputs. Namespace resource URLs are
digest-pinned inputs; drift fails verification instead of silently changing
the implementation target. Mutable errata and update pages are separate
review authorities in `monitoring.json`.

The normative language is in XSD Structures and XSD Datatypes. The primer,
built-in datatype schema, and examples are supporting material. The W3C schema
for schemas is normative only for the syntactic constraints it expresses.

`requirements/xsd-1.0.tsv` is the live implementation and evidence matrix.
Rows remain `missing` or `partial` until evidence directly covers the stated
requirement. The canonical
[specification decision register](../docs/specification-decisions.md) records
scope, interpretation, security, resource, compatibility, and wire decisions;
`decisions.md` remains a compatibility pointer.

Run `make provenance` to validate the local records. Set `VERIFY_REMOTE=1` to
download every pinned resource and verify its current bytes. Remote checking
is deliberately opt-in and is never part of parsing or compilation.

## Decision conformance matrix

| Decision | Evidence boundary |
| --- | --- |
| XSD-DEC-001 | Supported XSD feature line and exclusions |
| XSD-DEC-002 | Explicit resolver ownership |
| XSD-DEC-003 | XSTS applicability and matrix binding |
| XSD-DEC-004 | Bounded XSTS baseline claim |
| XSD-DEC-005 | Normative source and errata precedence |
| XSD-DEC-006 | Defensive XML DTD policy |
| XSD-DEC-007 | Namespace expanded-name identity |
| XSD-DEC-008 | XSD regular-expression dialect |
| XSD-DEC-009 | XSD identity XPath subset |
| XSD-DEC-010 | Datatype value-space comparison |
| XSD-DEC-011 | Deterministic non-canonical serialization |
| XSD-DEC-012 | Resource and diagnostic limits |
| XSD-DEC-013 | Validation entry-point parity and peer comparison |
