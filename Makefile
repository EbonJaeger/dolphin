PKGNAME=mcdolphin
MODULE=gitlab.com/EbonJaeger/dolphin
VERSION="1.3.0"

PREFIX?=/usr
BINDIR?=$(DESTDIR)$(PREFIX)/bin

GO?=go
GOFLAGS?=

GOSRC!=find . -name '*.go'
GOSRC+=go.mod go.sum

mcdolphin: $(GOSRC)
	$(GO) build $(GOFLAGS) \
		-ldflags " \
		-X main.Version=$(VERSION)" \
		-o $(PKGNAME) \
		./cmd/mcdolphin

all: mcdolphin

RM?=rm -f

clean:
	$(GO) mod tidy
	$(RM) $(PKGNAME)
	$(RM) -r vendor

install: all
	install -Dm755 $(PKGNAME) $(BINDIR)/$(PKGNAME)

uninstall:
	$(RM) $(BINDIR)/$(PKGNAME)

check:
	$(GO) get -u golang.org/x/lint/golint
	$(GO) get -u github.com/securego/gosec/cmd/gosec
	$(GO) get -u honnef.co/go/tools/cmd/staticcheck
	$(GO) get -u gitlab.com/opennota/check/cmd/aligncheck
	$(GO) fmt -x ./...
	$(GO) vet ./...
	golint -set_exit_status `go list ./... | grep -v vendor`
	gosec -exclude=G204 ./...
	staticcheck ./...
	aligncheck ./...
	$(GO) test -cover ./...

vendor: check clean
	$(GO) mod vendor

.DEFAULT_GOAL := all

.PHONY: all clean install uninstall check vendor