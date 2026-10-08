# Resolution and catalogs

`resolve.Resolver` is the only compilation I/O boundary. The default resolver
denies every request. `resolve.Memory` is suitable for tests, embedded schemas,
and applications that already control all resource bytes.

For the next major, both in-memory constructors admit the complete input before
allocating owned map capacity, parsing URIs, or copying content. `NewMemory`
defaults to 256 resources, 64 MiB cumulative URI/content input bytes and 16 MiB
content per resource. `NewCatalog` defaults to 256 mappings and 64 MiB cumulative
namespace/URI input bytes. Use `NewMemoryWithOptions` with `MemoryOptions` or
`NewCatalogWithOptions` with `CatalogOptions` to select different finite limits.
Limits are inclusive; zero selects the finite default and negatives are invalid.
Refusal returns nil and an error classified by `resolve.ErrLimitExceeded` with
safe default output. These are input allowances, not exact heap-size limits or
a bound on caller allocations. No constructor performs I/O. Resource and mapping
ownership remains independent of subsequent caller mutation. This intentionally
tightens acceptance for existing constructors; no release is implied.

`resolve.File` is an opt-in local filesystem capability. It accepts only
hostless absolute `file` URIs beneath one absolute configured root, confines
opens with `os.Root`, rejects symlink and traversal escapes, and caps each
resource with the inclusive `FileOptions.MaxBytes` limit (16 MiB by default).
`FileOptions.MaxURIBytes` independently admits request URI bytes before
parsing or filesystem work (inclusive 64 KiB by default). Zero selects the
default, negatives are invalid and refusal preserves `resolve.ErrLimitExceeded`.
Platform-native absolute paths, including Windows drive paths, remain confined
to the configured root. Close the resolver when the compiler no longer needs
it:

```go
files, err := resolve.NewFile(resolve.FileOptions{
    Root:     "/srv/schemas",
    MaxBytes: 4 << 20,
})
if err != nil {
    return err
}
defer files.Close()
```

`resolve.Catalog` maps import namespaces to absolute resource identities. It
enables `xs:import` without `schemaLocation` while leaving byte access with an
explicit underlying resolver:

```go
catalog, err := resolve.NewCatalog(map[string]string{
    "urn:orders": "file:///srv/schemas/orders.xsd",
}, files)
if err != nil {
    return err
}
```

Explicit schema locations pass through unchanged. A missing locationless
mapping behaves as an unavailable optional hint; compilation still rejects
any declaration that later depends on unresolved imported components.

Resolvers receive the absolute URI, requested namespace, and reference kind.
They must return bytes with the exact requested resource identity. Compiler
limits bound schemas, references, depth, components, particles, and total
content bytes. `compile.Limits.MaxURIBytes` separately bounds each root,
reference and returned identity before URI parsing, cache lookup or resolver
dispatch (inclusive 64 KiB by default). Zero selects the finite default;
negatives are invalid. A rejected reference never dispatches the resolver.
Returned identities are admitted before mismatch diagnostics, including
namespace-only imports. Refusal preserves `compile.ErrLimitExceeded` and
publishes no partial Set; observed cancellation retains priority.

There is no built-in HTTP resolver. Applications needing one should implement
`resolve.Resolver`, enforce an allowlist, cap response bytes and redirects,
and avoid forwarding ambient credentials. Use `resolve.File` instead of
mapping an untrusted schema location directly to a filesystem path.
