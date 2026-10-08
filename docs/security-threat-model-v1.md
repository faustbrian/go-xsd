# XSD security threat model

Contract version: 1

## Scope and assets

This model covers the root module, schema parser and model, compiler, datatype
translation, instance validation, serialization, builders and built-in
resolvers. It refines the shared
[ecosystem model](https://github.com/faustbrian/go-library-tools/blob/main/docs/ecosystem/security/threat-model.md).
Protect application availability, schema and assessment integrity, resource
contents and identities, diagnostic data, source, release artifacts and CI
credentials. Input includes schema/instance bytes, caller-owned trees and
models, names and namespaces, patterns, composition references, resolver
requests/results and injected collaborators.

The library does not authenticate tenants or authorize application resources.
Applications must establish those decisions before supplying documents or
granting resolver capabilities. Successful parsing is not authorization.

## Owned boundaries and controls

| Boundary | Control and limitation |
| --- | --- |
| Schema and instance XML | Explicit byte, depth and element/node budgets; DTD directives are refused and external entities are not expanded. Standard decoder token allocation is distinct from library-owned copies. |
| Schema parser ownership | Independent namespace-copy/declaration and retained string/copy-work admission; annotation source spans are admitted before raw capture. Unchanged parent scopes remain aliases. Compiler allowances are per document, not graph cumulative. |
| Schema graph and compilation | Deny resolution by default; explicit resolver injection, graph/document/reference/component/particle limits and independent particle-copy admission. Per-URI admission precedes root parsing, reference cache/dispatch and returned-identity processing. Compilation failures publish no partial Set. |
| Caller trees and validator ownership | Byte/copy-work and namespace-entry admission precede known clone capacities. Explicit tree namespaces and inherited reader namespaces remain intentionally different. Diagnostics and identity work have separate limits. |
| Pattern translation | Source/output limits and inclusive class-subtraction depth 256 bound owned work. Context-aware translation checks owned loops; standard-library sorting, compilation and matching remain synchronous. |
| Memory and Catalog configuration | Construction owns its copies and therefore requires count and cumulative input-byte admission before allocation/URI processing; Memory also needs a per-resource byte cap. Compiler limits alone cannot prove constructor admission. |
| File capability | Explicit root, per-resource bytes and independent request-URI admission before parsing/filesystem work; descriptor-relative confined opens, caller-owned Close and fresh content ownership. No implicit file capability is granted by compilation. |
| Serialization | Model preflight rejects cycles and bounds depth, component work and output before publication; builders and explicit model access remain caller-controlled data operations. |
| Default reporting | ParseError, Diagnostic, Location and public compiler/built-in resolver failures use supported safe output boundaries. Sentinel/cause inspection and explicit detailed projections remain trusted operations, not redacted data. |
| Automation and publication | Immutable action/tool identities, minimum permissions, required-result aggregation, isolated execution resources and applicable dependency/history/current-tree checks. Release qualification is independent of successful compilation or a source review. |

See [security and limits](security.md), [resolution](resolution.md) and
[validation](validation.md) for configurable policies and exact semantics.
Controls changed for the next major are not promises about an older published
release. Bind the release verdict to the actual version and immutable source.

## Residual ownership and review conditions

These records describe responsibilities or open work, not blanket acceptance of
an unverified vulnerability. The maintainer is `faustbrian`; applications own
the capabilities and collaborators they inject.

| Boundary | Owner, rationale and mitigation | Review condition |
| --- | --- | --- |
| Blocking readers, OS reads and injected resolvers | Application and maintainer: a context cannot forcibly interrupt an arbitrary synchronous collaborator. Use cooperative bounded readers/resolvers and caller-owned transport deadlines; check cancellation before and after owned work and discard partial failures. | Review when adding a blocking operation, changing resolver lifecycle or claiming interruption beyond owned checkpoints. |
| Network resolution | Application: arbitrary injected network code is outside built-in capability confinement. Restrict destinations, redirects, DNS/proxy behavior, credentials, decompression and response size; maintain deadlines. Default Deny does not secure a separately injected client. | Review whenever a network capability or destination policy changes. |
| Raw diagnostic data | Application and maintainer: exported data and trusted unwrapping retain useful detail. Keep these out of untrusted reporting. Invalid fmt verb/type diagnostics can bypass supported formatting protection; use valid formats rather than treating all verbs as safe. | Review new output interfaces, logging handlers, projections or error types. |
| Schema namespace/model copies | Maintainer: candidate parser controls independently bound namespace entries and model-string/copy work. Decoder token allocations and exact heap measurement remain outside this accounting; use finite source/element/depth limits too. | Review changed accounting, QName/annotation semantics and compiler propagation; runtime/release qualification remains separate from these source controls. |
| Constructor admission | Maintainer: ownership copies must be bounded at their constructor, not by a later compiler. Verify finite policies, exact/one-over refusals, nil failure results and independent successful storage. | Review changes to collection limits, URI normalization or ownership; qualify the actual implementation before release. |
| Release and maintainer integrity | Maintainer: source review does not prove artifact or publication integrity. Verify exact source, selected gates, public module resolution, signing/checksums and affected consumers; use private coordinated disclosure for confirmed findings. | Every affected release, changed dependency/workflow/signing identity or security incident. |

## Qualification and disclosure

Current source controls, ordinary tests, scanner results, residual decisions,
remote-main CI and public consumption are distinct evidence boundaries. A
pending or failed selected gate is not a pass; this document does not waive
coverage, mutation or other configured release requirements. No passing
whole-module security verdict or public next-major release is implied here.

Report privately under [SECURITY.md](../SECURITY.md). Apply the shared
severity, acknowledgement, remediation, embargo and advisory procedure there.
Record confirmed affected/fixed versions and safe upgrade guidance; do not
publish private reporter data, credentials or exploit-enabling artifacts.
