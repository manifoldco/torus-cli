OUT := ag-daemon
PKG := github.com/arigatomachine/cli/daemon
SHA := $(shell git describe --always --long --dirty)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
VERSION := $(shell cat ../cli/package.json \
  | grep version \
  | head -1 \
  | awk -F: '{ print $2 }' \
  | sed 's/[",version:]//g' \
  | tr -d '[[:space:]]')

all: server

server:
	go build -i -v -o ${OUT} -ldflags="-X ${PKG}/config.version=${VERSION}" ${PKG}

test:
	@go test -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
	    golint $$file ; \
	    done

static: vet lint
	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X ${PKG}/config.version=${VERSION}" ${PKG}

clean:
	-@rm ${OUT} ${OUT}-v*

.PHONY: run server static vet lint
