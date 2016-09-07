OUT := ag
PKG := github.com/arigatomachine/cli
SHA := $(shell git describe --always --long --dirty)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
VERSION := $(shell node -p -e "require('./cli/package.json').version")

all: binary

binary: generated
	go build -i -v -o ${OUT} -ldflags="-X ${PKG}/config.Version=${VERSION}" ${PKG}

test: generated
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

static: vet fmtcheck lint generated
	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X ${PKG}/config.Version=${VERSION}" ${PKG}

generated:
	go generate

clean:
	-@rm ${OUT} ${OUT}-v*

.PHONY: run server static vet fmtcheck lint generated test
