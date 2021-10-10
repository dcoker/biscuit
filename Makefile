GO ?= go
GOFLAGS := -v
PKG := ./...
TESTS := ".*"
GOIMPORTS := goimports
VERSION ?= $(shell git describe --long --tags --always)

.PHONY: build
	$(GO) install $(GOVERSIONLDFLAG) $(GOFLAGS)

.PHONY: test
test:
	$(GO) test $(GOFLAGS) $(PKG)

.PHONY: generate
generate:
	go generate ./...

.PHONY: check
check:
	$(GO) vet --all . 2>&1

.PHONY: fmt
fmt:
	gofmt -s -w .
	$(GOIMPORTS) -w .

.PHONY: clean
clean:
	rm -f $(GOPATH)/bin/$(PROGNAME)
	$(GO) clean $(GOFLAGS) -i $(PKG)

.PHONY: goreleaser-test
goreleaser-test:
	git tag ${VERSION}
	docker run --rm \
		-v $(shell pwd):/go/src/github.com/dcoker/biscuit \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/dcoker/biscuit \
		-e GITHUB_TOKEN=${GITHUB_TOKEN} \
		goreleaser/goreleaser release --rm-dist --snapshot


.PHONY: localstack
localstack:
	docker network create localstack
	docker run --name localstack --network localstack --rm -d -p 4566:4566 -p 4571:4571 localstack/localstack:0.12.17.5

.PHONY: docker
docker:
	docker build -t ghcr.io/dcoker/biscuit .

.PHONY: lint
lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run --config .github/golangci.yml -v
