// Package compile parses explicitly identified resources and resolves XML
// Schema 1.0 document graphs into immutable sets. New selects finite independent
// graph/copy/parser limits and a deny-all resolver when none is injected. Parsing
// does not grant I/O authority; custom resolvers own capabilities, cancellation,
// resource admission and concurrency safety.
//
// Effective document identity includes URI and adopted namespace, so one
// chameleon resource may appear in several compiled namespaces. Set accessors
// return copies suitable for caller mutation, not original lexical documents;
// accessor copying has no context or per-call budget. Compiler calls own their
// per-operation state and publish no partial set on error. Default public error
// output is a fixed category; explicit unwrapping exposes trusted causes.
package compile
