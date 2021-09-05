GO ?= go
GOFLAGS := -v
PKG := ./...
TESTS := ".*"
GOIMPORTS := goimports
PROGNAME := biscuit
VERSION := $(shell git describe --long --tags --always)
GOVERSIONLDFLAG := -ldflags="-X main.Version=$(VERSION)"

.PHONY: build
build: doc.go test
	$(GO) install $(GOVERSIONLDFLAG) $(GOFLAGS)

.PHONY: test
test:
	$(GO) test $(GOFLAGS) $(PKG)
	$(GO) test $(GOFLAGS) $(PKG)

doc.go: data
	/bin/echo -e '/*\n' > doc.go
	cat data/usage.txt >> doc.go
	/bin/echo -e '\n*/\npackage main' >> doc.go

.PHONY: check
check:
	$(GO) vet --all . 2>&1

.PHONY: fmt
fmt:
	gofmt -s -w .
	$(GOIMPORTS) -w .

.PHONY: clean
clean:
	rm doc.go
	rm -f $(GOPATH)/bin/$(PROGNAME)
	$(GO) clean $(GOFLAGS) -i $(PKG)

.PHONY: cross
cross:
	gox $(GOVERSIONLDFLAG) \
		-output 'build/{{.Dir}}/{{.OS}}_{{.Arch}}/biscuit' \
		-os "linux darwin windows" \
		-arch "amd64 arm arm64 386" \
		-osarch '!darwin/arm !darwin/386 !darwin/arm64'
	./cross.sh

.PHONY: docker-build
docker-build:
	docker build -f Dockerfile.e2e -t $(PROGNAME)/local .

.PHONY: docker-cross
docker-cross: docker-build
	mkdir build_docker_cross || /bin/true
	docker run -v $(shell pwd)/build_docker_cross/:/tmp/build/ $(PROGNAME)/local /bin/bash -xe -c \
		"rm -f build && ln -s /tmp/build/ build && make cross"


.PHONY: localstack
localstack:
	docker network create localstack
	docker run --name localstack --network localstack --rm -d -p 4566:4566 -p 4571:4571 localstack/localstack:0.12.17.5

.PHONY: docker
docker:
	docker build -t ghcr.io/dcoker/biscuit .
