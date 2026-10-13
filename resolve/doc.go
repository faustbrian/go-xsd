// Package resolve defines explicit XML Schema resource-loading capabilities.
// Deny is the default; Memory copies finite caller resources, Catalog maps
// locationless import namespaces, and File opts into one opened os.Root.
// Network access is application-injected, never implicit.
//
// Successful Resource bytes transfer to the caller and must remain stable.
// Resolver contexts must be non-nil; implementations own blocking I/O admission
// and cancellation. Memory is immutable; Catalog and Chain inherit delegate
// concurrency and lifetime obligations. File must be closed by its owner and
// does not turn arbitrary blocking filesystem reads into cancellable work.
// Supported default built-in errors use a fixed category; unwrapped causes
// remain trusted data. Confinement follows os.Root, not a filesystem sandbox.
package resolve
