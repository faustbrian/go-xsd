SHELL := /usr/bin/env bash

.PHONY: benchmark conformance fuzz interoperability

BENCH_TIME ?= 100ms

benchmark:
	./scripts/check-benchmarks.sh "$(BENCH_TIME)"

fuzz:
	./scripts/check-fuzz.sh

conformance:
	./scripts/check-provenance.sh
	./scripts/check-xsts.sh

interoperability:
	./scripts/check-differential.sh
