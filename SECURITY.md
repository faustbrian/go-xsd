# Security Policy

## Reporting

Report suspected vulnerabilities through the
[private reporting form](https://github.com/faustbrian/go-xsd/security/advisories/new).
Do not open a public issue containing exploit
details, credentials, private fixtures, or affected deployment information.

Include the affected module and version, impact, reproduction, preconditions,
and any suggested mitigation. Reports are acknowledged as soon as practical;
timelines depend on severity and verification.

The shared [vulnerability-management procedure](https://github.com/faustbrian/go-library-tools/blob/main/docs/ecosystem/security/vulnerability-management.md)
defines severity, acknowledgement and remediation targets, private triage,
embargo, advisory publication and coordinated affected-module releases.
The maintainer owns repository triage; private case records identify exact
affected and fixed versions without disclosing reporter data or credentials.

## Supported Versions

Published v1.0.0 and v1.1.0 have the diagnostic-privacy and bounded-work gaps
described in [security guidance](docs/security.md#published-v1-limitations).
No corrected v1 release is available. The maintainer continues private report
triage, but remediation of these findings uses the incompatible `/v2` module
rather than a promised v1 backport. [V2.0.0](https://github.com/faustbrian/go-xsd/releases/tag/v2.0.0)
is published and verified through clean public-module consumption.

Applications remaining on v1 must provide the temporary admission and
reporting controls in that guidance. Migrate using
[the v2 guide](docs/migration.md). Compatibility decisions follow
[`COMPATIBILITY.md`](COMPATIBILITY.md).

## Security Gates

Releases require isolated tests, race and hostile-input checks, exact coverage
and mutation results, `govulncheck`, secret scanning, license verification,
SBOM generation, provenance validation, and clean-consumer resolution. A
missing scanner or unavailable service is a failed gate, not a warning.

Security fixes MUST include a regression test that does not publish weaponized
details or real secrets. Credentials MUST be redacted from logs and evidence.

## Repository Assurance

The repository [safety and concurrency policy](AGENTS.md#safety-and-concurrency)
and [supply-chain policy](AGENTS.md#dependencies-and-supply-chain) define shared
trust boundaries and release requirements. Package-specific security guidance
refines those rules for its owned boundary.

The versioned [XSD threat model](docs/security-threat-model-v1.md) records
concrete boundaries and residual responsibilities. It is not a release verdict
or evidence that every selected security gate passed.
