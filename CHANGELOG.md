# Changelog

All notable released changes will be documented here. This project is not yet
at its first stable release.

## Unreleased

### Changed

- Adopt `go-library-tools` v1.0.13 for shared repository verification while
  retaining package-owned conformance, interoperability, fuzz, benchmark, and
  mutation evidence.

- Pin CI policy to canonical `go-library-tools` revision
  `0bbff11f25d74203018019fa5b26ae1443310fe7`, including offline and online
  specification decision enforcement.

### Added

- Add machine-enforced [specification decisions](docs/specification-decisions.md),
  conformance bindings, append-only decision history, and separately monitored
  XSD, XSTS, XML, and Namespaces authorities. The JAXP lane remains a
  validity-only maintained-peer comparison, and XSTS results are explicitly
  bounded to the documented matrix rather than universal XSD conformance.

  Decision records: XSD-DEC-001 sha256:19b15b0bd7434c126bbdb9db9f3734542e21c511143a9257665b97024b05a84c;
  XSD-DEC-002 sha256:63c571267586feb8322b0f6906c4bcfdf75319ebbad4b41de319fb16a87dd19c;
  XSD-DEC-003 sha256:04525d003609b510a47435b78b4e8b6ed78396736721d4a93dbcb89f0efd4af0;
  XSD-DEC-004 sha256:8468098c06390f78827e924e7016f63ce99adf55a2445e54ae3d7cdac79d63b7;
  XSD-DEC-005 sha256:f32588e1e92c5d00a0283f98bde8dab4101a8f53c9fc2236d7bb4e657afb2d3f;
  XSD-DEC-006 sha256:00b0ad09f357d201fa52f7ae16a78025d12a413161edcb79dfc14713f02abb5d;
  XSD-DEC-007 sha256:4f64b7f6669be863ee08acad02f0d81eceb4eda575822afbca8230f00af2db90;
  XSD-DEC-008 sha256:a59307791251be1eccadac67afd10cbc2fc7327c55718272a8085970dec123a7;
  XSD-DEC-009 sha256:724e67326f218ab1609b03b042295f33aff31ceb41134151164958b6d52ed073;
  XSD-DEC-010 sha256:cd1a11dccde7706717bb424e7b673e466080e7985d5c13b7cf612f474995f499;
  XSD-DEC-011 sha256:c9fd12a12a0355471cf92bd550074b97440361401a33a11889f8cea2b56951c6;
  XSD-DEC-012 sha256:f4cdd5fd86022abcb75d0a29d424c1e68e0ba2a11fc463d5a7a023e953de13e6;
  XSD-DEC-013 sha256:7da45bc35a12cf7041458bd8a81b7f75b16a759823da9b09630e8bb42e198463.

### Documentation

- Remove the archived monorepo documentation link; package guidance remains in
  the repository-owned documentation.

## 1.0.0 - 2026-08-25

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Link the package README to the repository-wide Golib documentation portal.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-xsd` identity while preserving its documented API and behavior.
- Validate native and JAXP benchmark output with the standard shell toolchain
  so clean Linux CI runners do not require an undeclared ripgrep installation.
- Link conformance and migration guidance directly to the canonical
  specification decision register.
- Delegate local mutation checks to the canonical exact-100 repository runner
  instead of a reduced package-local efficacy threshold.
- Avoid retaining a pre-compilation complex-type lookup that must be refreshed
  after recursive base compilation.
- Execute API compatibility tooling against the isolated module graph so owned
  dependency source changes cannot conflict with release checksums.
- Require exact per-production-package coverage with the pinned XSTS corpus,
  including serializer partial-write and reflection-budget failure paths.
- Clarify simple-content base resolution and simplify equivalent serializer
  initialization so strict static analysis remains clean.
- Run Java interoperability, official XSTS conformance, and reference
  benchmarks through the root gate using one digest-pinned, network-isolated
  Eclipse Temurin container instead of relying on host Java.
- Separate official XSTS conformance from JAXP differential interoperability
  so both results remain attributable.
- Use the repository-pinned current `apidiff` revision for the canonical API
  compatibility gate.

### Fixed

- Preserve root confinement for platform-native file URIs and accept resources
  whose size exactly matches the configured byte limit.
- Reject simple types without a restriction, list, or union during parsing
  instead of returning a document that deterministic serialization rejects.
- Propagate resolver file-close failures without discarding read failures and
  verify differential-corpus manifest cleanup.
- Bound Unicode range-table expansion before iteration so malformed or
  corrupted tables fail closed instead of amplifying work.

### Added

- Add a canonical, evidence-linked specification decision register covering
  the supported feature line, source precedence, XSTS, secure XML handling,
  namespace identity, regex, XPath, value spaces, serialization, and limits.
- Add module-local MIT license metadata for clean consumer and supply-chain
  tooling.
- Establish the pinned XML Schema 1.0 specification and evidence matrix.
- Add a pinned, fail-closed public API compatibility baseline for the complete
  multi-package module.
- Add secure parsing, bounded resolution and compilation, immutable schema
  sets, instance validation, datatype support, deterministic serialization,
  and checked builders.
- Complete the XML Schema 1.0 requirement matrix with executable evidence.
- Add correctness-gated JAXP reference benchmarks and a public `wsdl`
  consumer contract.
