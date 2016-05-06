GO ?= go
GOFLAGS := -v
PKG := ./...
TESTS := ".*"
GOIMPORTS := ../../../../bin/goimports
GOLINT := ../../../../bin/golint
GLOCK := ../../../../bin/glock
BINDATA := ../../../../bin/go-bindata
PROGNAME := biscuit
VERSION := $(shell git describe --long --tags --dirty --always)

.PHONY: build
build: doc.go bindata.go test
	$(GO) install -ldflags="-X main.Version=$(VERSION)" $(GOFLAGS)

$(GLOCK):
	go get -v github.com/robfig/glock

.PHONY: glock-sync
glock-sync: $(GLOCK)
	$(GLOCK) sync github.com/dcoker/$(PROGNAME)

.PHONY: glock-save
glock-save: $(GLOCK)
	$(GLOCK) save github.com/dcoker/$(PROGNAME)

.PHONY: setup
setup: $(GLOCK) glock-sync
	$(GLOCK) install github.com/dcoker/$(PROGNAME)

.PHONY: test
test:
	$(GO) test $(GOFLAGS) -i $(PKG)
	$(GO) test $(GOFLAGS) $(PKG)

bindata.go: data
	$(BINDATA) -o bindata.go -prefix data -ignore=\\.gitignore data/...
	gofmt -s -w bindata.go

doc.go: data
	/bin/echo -e '/*\n' > doc.go
	cat data/usage.txt >> doc.go
	/bin/echo -e '\n*/\npackage main\n' >> doc.go

.PHONY: check
check:
	$(GO) tool vet --all . 2>&1
	$(GOLINT) $(PKG)

.PHONY: fmt
fmt:
	gofmt -s -w .
	$(GOIMPORTS) -w .

.PHONY: clean
clean:
	rm doc.go bindata.go
	$(GO) clean $(GOFLAGS) -i $(PKG)

docker-build:
	docker build -f Dockerfile.e2e -t $(PROGNAME)/local .
