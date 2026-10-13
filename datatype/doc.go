// Package datatype implements the supported XML Schema 1.0 Second Edition
// lexical/value-space operations, exact decimal and integer values, partial
// calendar/duration orders, and bounded translation to Go regular expressions.
// Prefix binding, schema resolution and complete instance validation belong to
// the caller/compiler/validator, not standalone lexical checks.
//
// Decimal and Integer values are immutable and their zero values denote zero.
// Facet structs and regexp values remain caller-owned configuration. Decimal
// and integer parsing cap lexical bytes; standalone calendar/lexical helpers
// do not all share that cap or accept contexts. Admit caller input separately.
// Detailed errors may contain lexical values. Pattern translation has its own
// byte/depth limits and cooperative context checkpoints, not interruptible
// standard-library compilation/matching. Standard Unicode category tables are
// trusted shared inputs; the fixed block inventory is not generated at runtime.
package datatype
