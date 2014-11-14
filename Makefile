VERSION := 0.0.2

GOFLAGS := -ldflags "-X main.Version $(VERSION) -X main.Release true"
GOOS    := $(shell go env GOOS)
GOARCH  := $(shell go env GOARCH)

ARCHIVE := pico-cdn-$(VERSION).$(GOOS).$(GOARCH).tar.gz
DISTDIR := dist/$(GOOS)-$(GOARCH)

.PHONY: clean default deps fmt release start test

default: deps
	go build $(GOFLAGS)

start:
	clear
	go test
	go build
	pico-cdn master

test:
	go test

archive: dist/$(ARCHIVE)

clean:
	go clean ./...
	git clean -fdx

deps:
	go get -d

dist/$(ARCHIVE): $(DISTDIR)/pico-cdn
	tar -C $(DISTDIR) -czvf $@ .

$(DISTDIR)/pico-cdn: deps
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS) -o $@