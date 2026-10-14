# XSTS baseline

[Documentation index](README.md) · [Repository entry](../README.md)

The XSTS harness was verified against every test-set metadata file in the
pinned archive. These historical results are conformance evidence, not proof
of filesystem confinement.

The next-major harness uses one `os.Root` directory capability for metadata,
schemas, instances and imported schemas. Relative symlinks within the suite
root are supported; outside-root targets and absolute symlinks are rejected.
`xsdtest.Run` reads caller-trusted metadata and decodes its XML before checking
cancellation at group traversal. Startup decode errors therefore precede group
cancellation and return an empty report. This initial ingestion has no separate
tenant-style byte or decode-work quota; callers must trust their suite inputs.
The official corpus is authenticated by its pinned digest before extraction.
Fixture-access refusals are reported as errors, not passing invalid-schema
expectations. Ordinary schema invalidity retains its conformance meaning.
The caller selects and trusts the root directory. This is not a sandbox for
device files, mounted filesystems or malicious changes to the root itself.
Confinement relies on Go's `os.Root` guarantees on the supported native
platforms; it must not be claimed for JavaScript targets.

- Date: 2026-07-19
- Archive: XSTS 2007-06-20
- SHA-256: `902176b25e4111cf96b08663107521a4992e8ea67aad6b815592a6a5b4b9ea06`
- Filter: none
- Accepted expectations passed: 24,696
- Failed: 0
- Skipped accepted expectations: 0
- Excluded queried expectations: 90

Queried expectations are disputed upstream metadata rather than accepted
conformance requirements. The harness reports them separately and fails if
any accepted valid or invalid expectation fails or is skipped.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
