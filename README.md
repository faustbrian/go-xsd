# xsd

[![CI](https://github.com/faustbrian/go-xsd/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-xsd/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-xsd/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-xsd.svg)](https://pkg.go.dev/github.com/faustbrian/go-xsd)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-xsd?sort=semver)](https://github.com/faustbrian/go-xsd/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`xsd` is a secure XML Schema 1.0 parser, compiler, validator, serializer,
and builder for Go. It is intended to provide the schema layer for `wsdl`
and SOAP tooling without performing implicit file or network access.

> [!NOTE]
> The released module follows stable v1 compatibility. The
> documented XML Schema 1.0 surface is evidence-mapped and the pinned XSTS
> baseline passes; neither statement claims support for XML Schema 1.1 or for
> behavior outside the published matrix.

## Install

The stable v1 release requires Go 1.27.0.

```sh
go get github.com/faustbrian/go-xsd@v1
```

## Quick start

The [compiler-checked package example](example_test.go) parses, compiles, and
validates a small schema without external I/O. See the [API guide](docs/api.md)
for package selection and ownership boundaries.

Parsing and validation reject DTDs. Compilation uses a deny-by-default
resolver and bounded schema graphs. See [the documentation index](docs/README.md),
[the live support matrix](specification/requirements/xsd-1.0.tsv), and
[specification provenance](specification/README.md). Observable ambiguities
and package policy are recorded in the
[specification decision register](docs/specification-decisions.md).

Use [GitHub issues](https://github.com/faustbrian/go-xsd/issues) for adoption
questions and defect reports, and follow the [support](SUPPORT.md) and
[security](SECURITY.md) policies.

Run `make check` for formatting, static analysis, tests, the race detector,
and provenance checks.

For ecosystem-wide selection and ownership guidance, see the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Protocols and descriptions family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).
