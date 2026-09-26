# Copyright NU Cybernetics. PDUM — research prototype.
SHELL := /bin/bash
SNAPSHOT ?= $(shell ls -d data/snapshots/snap-* 2>/dev/null | sort | tail -1)

.PHONY: check install go-check ts-check build e2e schemas validate candidate audit-verify

install:
	pnpm install --frozen-lockfile

go-check:
	gofmt -l . | tee /dev/stderr | { ! grep -q .; }
	go build ./...
	go vet ./...
	go test ./...

ts-check:
	pnpm typecheck
	pnpm lint
	pnpm test

build:
	pnpm build

check: go-check ts-check build

schemas:
	pnpm build:jsonschema

validate:
	go run ./cmd/pdumctl snapshot validate $(SNAPSHOT)

candidate:
	go run ./cmd/pdumctl model run --snapshot $(SNAPSHOT) --out data/candidates/$$(date -u +cand-%Y-%m-%d-%H%M%S)

audit-verify:
	go run ./cmd/pdumctl audit verify data/audit/audit.jsonl

e2e:
	pnpm e2e
