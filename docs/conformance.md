# Support and conformance

The stable target is the documented surface of XML Schema 1.0 Second Edition
plus reviewed errata, using XML 1.0 Fifth Edition and Namespaces in XML 1.0
Third Edition at the parsing boundary. XML Schema 1.1 is outside stable scope.
Exact source versions and digests are recorded in
`specification/manifest.tsv`; mutable errata and update pages are monitored
separately in `specification/monitoring.json`.

The [specification decision register](specification-decisions.md) records the
scope choices, alternatives, consequences, and executable evidence behind this
support matrix.

`specification/requirements/xsd-1.0.tsv` is the support contract. `implemented`
means the row has executable evidence for its stated scope. `partial` means a
useful subset exists but the broad feature is not complete. `missing` means no
support is claimed.

The official XSTS 2007-06-20 suite is pinned by digest. `make xsts` downloads
that exact archive, confines resource access to the extracted suite root, and
runs every accepted valid or invalid expectation. All 24,696 accepted
expectations passed with no failures or skips in the recorded baseline; 90
upstream `queried` expectations are reported separately. This is evidence for
the matrix rows exercised by the suite, not a universal XSD conformance claim.
A regression or newly identified semantic gap
must reopen the affected row instead of weakening its executable evidence.
New support claims require a normative matrix row, focused tests, and
applicable XSTS evidence. The measured result is recorded in the
[XSTS baseline](xsts-baseline.md).

`make differential` runs a shared positive and negative corpus through byte,
incremental-reader, and caller-owned tree validation, then runs the same
corpus through the JDK JAXP XML Schema reference implementation. The Java
reference runs without network access in the digest-pinned Eclipse Temurin 25
container declared by `scripts/run-java-reference.sh`.
The peer lane compares validity outcomes only; JAXP diagnostics and resource
policy are not treated as equivalent or normative.
