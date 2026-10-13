SHELL := /usr/bin/env bash
GOLIB ?= golib

.PHONY: source-docs-check check check-full ci cohesion inventory repository-check specification-check specification-local

check: source-docs-check
	$(GOLIB) check --local --all

check-full: source-docs-check
	$(GOLIB) check --all

ci: source-docs-check repository-check cohesion specification-local check

cohesion:
	$(GOLIB) cohesion check

inventory repository-check:
	$(GOLIB) repository check

specification-check:
	$(GOLIB) specification check --online

specification-local:
	$(GOLIB) specification check

source-docs-check:
	$(GOLIB) source-docs check
