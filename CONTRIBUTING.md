# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected module's goals and docs.
2. Run `make inventory` and the narrow baseline gate for the module.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
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

Required mutation gates must finish with zero surviving viable mutants.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Classify the change under [proportional assurance](AGENTS.md#proportional-assurance)
and select assertions for the affected observable behavior. The root module's
explicit `coverage.modules` evidence mode still executes the complete module
coverage command with all production targets and reports genuine counts.
Missing or malformed profiles, absent packages and test failures remain errors.
Only universal exact-percentage acceptance changes; other configured gates,
mutation acceptance and fail-closed CI aggregation are unchanged. Coverage
collection does not certify assertion quality, security adequacy or readiness.

### Development tooling route

CI uses the reviewed, immutable Tools source
`606c3e9da5112086217f7108f335c2f88528bc14` for both the reusable workflow and
`tooling_sha`, with `source_bootstrap: true`. This route builds tooling from
that pristine source and runs the full `golib check --all`, not `--local`.
It is an explicit development-tooling route, not stable-binary qualification.
`public_dependencies: true` selects the official public proxy and checksum
database instead of a private bootstrap archive; XSD's external module
dependency is the published `go.uber.org/goleak v1.3.0`.
The declared v1.8.5 version and checksum remain published history; that binary
cannot parse the optional coverage policy and is not an equivalent executor.
For local development, build `cmd/golib` inside a separately checked-out
pristine copy of that pinned Tools source, using task-owned disposable Go
caches. Set Make's `GOLIB` variable to the resulting executable for the
commands below; do not run them with the incompatible declared binary.

Source-bootstrap CI, stable SDK qualification and publication, XSD v2 release
qualification/publication, a clean public v2 consumer, and WSDL's separately
released v2 adoption remain distinct boundaries. Published WSDL and the Tools
compatibility consumer still use unsuffixed XSD v1; they do not prove candidate
v2 interoperability. Failed strict coverage run `37493400899` at
`ff892bb2784214d7a7035b97ffabd94b00a1da83` remains failed. This policy change
does not repair that historical result or close material uncovered behavior.

Run during development:

```bash
make inventory
make check
```

Before submitting a repository-wide change:

```bash
make ci
```

The full scheduled and release gate is `make ci`. Report every unavailable or
failing command; do not describe partial results as release-ready.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
