SHELL := /usr/bin/env bash
GOLIB ?= golib

.PHONY: check check-full ci cohesion inventory repository-check specification-check specification-local

check:
	$(GOLIB) check --local --all

check-full:
	$(GOLIB) check --all

ci: repository-check cohesion specification-local check

cohesion:
	$(GOLIB) cohesion check

inventory repository-check:
	$(GOLIB) repository check

specification-check:
	$(GOLIB) specification check --online

specification-local:
	$(GOLIB) specification check
