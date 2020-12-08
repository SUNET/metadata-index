ifeq ($(SHELL), cmd)
	VERSION := $(shell git describe --exact-match --tags 2>nil)
	HOME := $(HOMEPATH)
else ifeq ($(SHELL), sh.exe)
	VERSION := $(shell git describe --exact-match --tags 2>nil)
	HOME := $(HOMEPATH)
else
	VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
endif

PREFIX := /usr
BINDIR := $(PREFIX)/bin
ETCDIR := /etc
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l -s $(filter-out plugins/parsers/influx/machine.go, $(GOFILES)))
BUILDFLAGS ?=
INSTALL := install
INSTALL_EXEC := $(INSTALL) -D --mode 755
INSTALL_DATA := $(INSTALL) -D --mode 0644

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(shell go env GOPATH))/bin:$(PATH)
endif

LDFLAGS := -X github.com/sunet/metadata-index/pkg/meta.commit=$(COMMIT) -X github.com/sunet/metadata-index/pkg/meta.branch=$(BRANCH)
ifdef VERSION
	LDFLAGS += -X github.com/sunet/metadata-index/pkg/meta.version=$(VERSION)
endif

.PHONY: all
all:
	@$(MAKE) --no-print-directory mix #docs/mix.1

.PHONY: swag
swag:
	swag init -g pkg/api/api.go

.PHONY: mix
mix: swag
	go build $(GO_BUILD_FLAGS) -ldflags "$(LDFLAGS)" ./cmd/mix

docs/%.1: docs/%.ronn.1
	ronn -r $< > $@

.PHONY: install
install: mix
	$(INSTALL_EXEC) mix $(DESTDIR)$(BINDIR)


.PHONY: test
test:
	go test $(GO_BUILD_FLAGS) -cover -short ./...

.PHONY: testcover
testcover:
	go test $(GO_BUILD_FLAGS) -cover ./...

.PHONY: test-all
test-all: fmtcheck vet
	go test ./...

.PHONY: clean
clean:
	rm -f mix 
	rm -f mix.exe
	rm -f docs/mix.1

.PHONY: docker
docker:
	docker build -t "mix:$(COMMIT)" .
	docker tag mix:$(COMMIT) docker.sunet.se/mix:$(COMMIT)
	docker tag mix:$(COMMIT) docker.sunet.se/mix:latest
	docker push docker.sunet.se/mix:$(COMMIT)
	docker push docker.sunet.se/mix:latest

.PHONY: deb-source
deb-source:
	go mod vendor
	dpkg-buildpackage -S -k$(DEBSIGN_KEYID)

.PHONY: deb-gin
deb-bin:
	go mod vendor
	dpkg-buildpackage -k$(DEBSIGN_KEYID)
