OUT = ag
PKG = github.com/arigatomachine/cli
PKG_LIST = $(shell go list ${PKG}/... | grep -v /vendor/ | grep -v /data/bindata.go)
GO_FILES = $(shell find . -name '*.go' | grep -v /vendor/ | grep -v /data/bindata.go)
VERSION = $(shell git describe --tags --abbrev=0 | sed 's/^v//')

PUBLIC_KEY ?= data/keys/production.json

all: binary

binary: bindata
	go build -i -v -o ${OUT} -ldflags="-X ${PKG}/config.Version=${VERSION}" ${PKG}

test: bindata
	@go test -short $$(glide nv)

vet:
	@go vet ${PKG_LIST}

fmtcheck:
	@FMT=$$(for file in ${GO_FILES} ;  do \
		gofmt -l -s $$file ; \
	done) ; \
	if test -n "$$FMT"; then \
		echo "gofmt problems:" ; \
		echo "$$FMT" ; \
		exit 1 ; \
	fi ;

lint:
	@LINT=$$(for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done) ; \
	if test -n "$$LINT"; then \
		echo "go lint problems:" ; \
		echo "$$LINT" ; \
		exit 1 ; \
	fi ;

bindata: data/bindata.go
data/bindata.go: data/ca_bundle.pem data/public_key.json
	go-bindata -pkg data -o $@ $^

data/public_key.json: $(PUBLIC_KEY)
	ln -sf ../$< $@

static: vet fmtcheck lint bindata
	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X ${PKG}/config.Version=${VERSION}" ${PKG}


npm: builds/npm/package.json builds/npm/README.md builds/npm/LICENSE.md builds/npm/scripts/install.js npm-bin

builds/npm builds/npm/bin builds/npm/scripts:
	mkdir -p $@

builds/npm/package.json: npm/package.json.in builds/npm
	sed 's/VERSION/$(VERSION)/' < $< > $@

builds/npm/README.md: npm/README.md
	cp $< $@

builds/npm/LICENSE.md: LICENSE.md
	cp $< $@

builds/npm/scripts/install.js: npm/install.js builds/npm/scripts
	cp $< $@

npm-bin: builds/npm/bin builds/ag-*
	cp builds/ag-* builds/npm/bin/

clean:
	@rm -f ${OUT} ${OUT}-v*
	@rm -f data/bindata.go
	@rm -f data/public_key.json
	@rm -rf builds/*

.PHONY: static vet fmtcheck lint test bindata
