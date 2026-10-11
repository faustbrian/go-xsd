# Contributing

## Before Editing

1. Read the README, [`AGENTS.md`](AGENTS.md) and the affected API contract.
2. Start with `go test ./...` or a focused package/test selector.
3. Inspect direct consumers when changing a public or composition contract.
4. Preserve unrelated work and generated/corpus provenance.

An ordinary fix stays in this repository. It does not require sibling checkouts,
Tools edits, ecosystem catalog updates or agent-goal ledgers. Change manifests
only when their owned facts change, not merely because source changed.

## Changes

Keep commits focused and conventional. Update changelogs for material public
behavior, compatibility or adoption changes, not CI-only administration.
Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable entries in the
[specification decision register](docs/specification-decisions.md), and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.

Mutation is optional developer-local or selected scheduled assurance, not a PR
prerequisite. Investigate viable survivors and add focused regressions for real
defects; moving a campaign off PRs does not waive a confirmed correctness or
security defect. Equivalent/invalid mutants retain reviewed narrow dispositions.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Classify the change under [proportional assurance](AGENTS.md#proportional-assurance)
and select assertions for the affected observable behavior. Explicit coverage
collection retains the root module's `coverage.modules` evidence mode: it
executes the complete module coverage command with all production targets and
reports genuine counts.
Missing or malformed profiles, absent packages and test failures remain errors.
Coverage collection does not certify assertion quality, security adequacy or
readiness, and is not required for every ordinary local change.

### Public tooling and bounded local verification

CI consumes the checksum-pinned public Tools `v2.1.1` binary. The reusable
workflow and `tooling_sha` both select immutable source
`eff8b2127c361ee2f422123824a1d337cc277ac4`, with `source_bootstrap: false`;
ordinary contributions do not require a Tools source build or sibling checkout.
`public_dependencies: true` selects the official public proxy and checksum
database instead of a private bootstrap archive; XSD's external module
dependency is the published `go.uber.org/goleak v1.3.0`.

For ordinary source work, `go test ./...` runs deterministic tests, fuzz seeds
and saved corpus regressions. With the qualified verifier selected by repository
policy, use:

```bash
make inventory
make check
make ci
```

`check` and `ci` use bounded local checks, not full mutation, conformance or
interoperability campaigns. `ci` retains offline specification validation;
`make specification-check` retains explicit online provenance checks.
Documentation changes need relevant structure, links and examples, not unrelated
runtime work. Public API, concurrency and external-boundary changes retain their
affected compatibility, race or integration checks.

For active exploration, select one fuzz target and a small budget; long campaigns
belong in scheduled assurance. Optional local mutation selects one package with
the qualified verifier. These examples are not all-change requirements:

```bash
GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzParseSchema$' -fuzztime=2s -parallel=1 -timeout=2m .
golib mutation --module . --package .
```

`make check-full` retains explicit full assurance; relevant release qualification
uses `golib release dry-run --module .`. Select these for their material risk or
release boundary, not every commit. Report unavailable/failed checks honestly;
a fast local pass is not public-release or hosted-CI qualification.

Tools adoption, XSD release qualification and clean public consumption, and any
separately released WSDL adoption remain distinct boundaries. Failed strict
coverage run `37493400899` at `ff892bb2784214d7a7035b97ffabd94b00a1da83`
remains failed; these routing changes do not repair that historical result or
close material uncovered behavior.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
