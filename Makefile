SHELL := /usr/bin/env bash

GO111MODULE  ?= on
GO           ?= go

.PHONY: format
format:
	go fmt $(pkgs)

.PHONY: style
style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: test
test:
	@$(GO) test $(pkgs)

.PHONY: test-short
test-short:
	@echo ">> running short tests"
	@$(GO) test -short $(pkgs)

.PHONY: vet
vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)
