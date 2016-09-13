OUT=ag
PKG=github.com/arigatomachine/cli

TARGETS=\
	darwin-amd64 \
	linux-amd64
GO_REQUIRED_VERSION=1.7.1

PUBLIC_KEY?=data/keys/production.json

VERSION=$(shell git describe --tags --abbrev=0 | sed 's/^v//')

all: binary
ci: binary vet fmtcheck lint test

.PHONY: all ci

#################################################
# Bootstrapping for base golang package deps
#################################################

BOOTSTRAP=\
	github.com/Masterminds/glide \
	github.com/golang/lint/golint \
	github.com/jteeuwen/go-bindata/...

$(BOOTSTRAP):
	go get -u $@
bootstrap: $(BOOTSTRAP)

.PHONY: bootstrap

#################################################
# Build targets
#################################################

VERSION_FLAG=-X $(PKG)/config.Version=$(VERSION)
STATIC_FLAGS=-extldflags "-static" -w -s
GO_BUILD=go build -i -v

binary: bindata vendor
	$(GO_BUILD) -o ${OUT} -ldflags='$(VERSION_FLAG)' ${PKG}

static: bindata vendor
	$(GO_BUILD) -o ${OUT}-v${VERSION} \
		-ldflags='$(STATIC_FLAGS) $(VERSION_FLAG)' ${PKG}

.PHONY: binary static

#################################################
# Build targets for releasing
#################################################

GO_VERSION=$(shell go version)
gocheck:
ifeq (,$(findstring $(GO_REQUIRED_VERSION),$(GO_VERSION)))
	$(error "Go Version $(GO_REQUIRED_VERSION) is required.")
endif

OS=$(word 1, $(subst -, ,$*))
ARCH=$(word 2, $(subst -, ,$*))
BINARY=-o builds/$(OUT)-$(OS)-$(ARCH)
$(addprefix release-,$(TARGETS)): release-%: gocheck bindata vendor
	GOOS=$(OS) GOARCH=$(ARCH) $(GO_BUILD) $(BINARY)

release-all: $(addprefix release-,$(TARGETS))

.PHONY: gocheck $(addprefix release-,$(TARGETS)) release-all

#################################################
# Code generation and dependency grabbing
#################################################

bindata: data/bindata.go
data/bindata.go: data/ca_bundle.pem data/public_key.json
	go-bindata -pkg data -o $@ $^

data/public_key.json: $(PUBLIC_KEY)
	ln -sf ../$< $@

vendor: glide.lock
	glide install

.PHONY: bindata

#################################################
# Cleanup
#################################################

clean:
	@rm -f ${OUT} ${OUT}-v*
	@rm -f data/bindata.go
	@rm -f data/public_key.json
	@rm -rf builds/*

.PHONY: clean

#################################################
# Test and linting
#################################################

GO_FILES=$(shell find . -name '*.go' | grep -v /vendor/ | \
		grep -v /data/bindata.go)

EACH_FILE=\
	@RES=$$(for file in ${GO_FILES} ;  do \
		$(2) $$file ; \
	done) ; \
	if test -n "$$RES"; then \
		echo "$(1) problems:" ; \
		echo "$$RES" ; \
		exit 1 ; \
	fi ;

test: bindata vendor
	@go test -short $$(glide nv)

vet:
	@go vet $$(glide nv)

fmtcheck:
	$(call EACH_FILE,gofmt,gofmt -l -s)

lint:
	$(call EACH_FILE,golint,golint)

.PHONY: vet fmtcheck lint test

#################################################
# Docker targets
#################################################

PWD=$(shell pwd)
IMAGE=arigato/cli:latest
RUN_IN_DOCKER=\
	docker run --name cli --rm \
		-v $(PWD):/go/src/github.com/arigatomachine/cli \
		-v $(PWD)/builds:/builds \
		$(IMAGE) $(1)

docker-build:
	$(call RUN_IN_DOCKER,binary)

docker-test:
	$(call RUN_IN_DOCKER,ci)

container:
	docker build -t arigato/cli:latest .

.PHONY: docker-build docker-test container

#################################################
# Distribution via npm
#################################################

NPM_DEPS=\
	builds/npm/package.json \
	builds/npm/README.md \
	builds/npm/LICENSE.md \
	builds/npm/scripts/install.js \
	npm-bin
npm: $(NPM_DEPS)

builds/npm builds/npm/bin builds/npm/scripts:
	mkdir -p $@

builds/npm/package.json: npm/package.json.in builds/npm
	sed 's/VERSION/$(VERSION)/' < $< > $@

builds/npm/README.md: npm/README.md builds/npm
	cp $< $@

builds/npm/LICENSE.md: LICENSE.md builds/npm
	cp $< $@

builds/npm/scripts/install.js: npm/install.js builds/npm/scripts
	cp $< $@

npm-bin: builds/npm/bin builds/ag-*
	cp builds/ag-* builds/npm/bin/

.PHONY: npm npm-bin
