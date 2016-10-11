OUT=torus
PKG=github.com/arigatomachine/cli

TARGETS=\
	darwin-amd64 \
	linux-amd64
GO_REQUIRED_VERSION=1.7.1

VERSION=$(shell git describe --tags --abbrev=0 | sed 's/^v//')

all: binary
ci: binary vet fmtcheck simple lint test

.PHONY: all ci

#################################################
# Bootstrapping for base golang package deps
#################################################

BOOTSTRAP=\
	github.com/Masterminds/glide \
	github.com/golang/lint/golint \
	honnef.co/go/simple/cmd/gosimple \
	github.com/jteeuwen/go-bindata/...

$(BOOTSTRAP):
	go get -u $@
bootstrap: $(BOOTSTRAP)

.PHONY: bootstrap

#################################################
# Build targets
#################################################

VERSION_FLAG=-X $(PKG)/config.Version=$(VERSION)
STATIC_FLAGS=-w -s
GO_BUILD=CGO_ENABLED=0 go build -i -v

binary: generated vendor
	$(GO_BUILD) -o ${OUT} -ldflags='$(VERSION_FLAG)' ${PKG}

static: generated vendor
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
$(addprefix release-,$(TARGETS)): release-%: gocheck generated vendor
	GOOS=$(OS) GOARCH=$(ARCH) $(GO_BUILD) $(BINARY) \
		-ldflags='$(STATIC_FLAGS) $(VERSION_FLAG)' ${PKG}

release-all: $(addprefix release-,$(TARGETS))

.PHONY: gocheck $(addprefix release-,$(TARGETS)) release-all

#################################################
# Code generation and dependency grabbing
#################################################

TOOLS=tools/bin

GENERATED_FILES=\
	data/zz_generated_bindata.go \
	envelope/zz_generated_envelope.go \
	primitive/zz_generated_primitive.go
generated: $(GENERATED_FILES)

data/zz_generated_bindata.go: data/ca_bundle.pem data/public_key.json
	go-bindata -pkg data -o $@ $^

primitive/zz_generated_primitive.go envelope/zz_generated_envelope.go: $(TOOLS)/primitive-boilerplate primitive/primitive.go
	$^

vendor: glide.lock
	glide install

$(TOOLS)/primitive-boilerplate: $(wildcard tools/primitive-boilerplate/*.go) $(wildcard tools/primitive-boilerplate/*.tmpl)
	$(GO_BUILD) -o $@ ./tools/primitive-boilerplate

.PHONY: generated

#################################################
# Cleanup
#################################################

clean:
	@rm -f ${OUT} ${OUT}-v*
	@rm -f $(GENERATED_FILES)
	@rm -f $(TOOLS)/*
	@rm -rf builds/*

.PHONY: clean

#################################################
# Test and linting
#################################################

GO_FILES=$(shell find . -name '*.go' | grep -v /vendor/ | \
		grep -v /data/zz_generated_bindata.go)

EACH_FILE=\
	@RES=$$(for file in ${GO_FILES} ;  do \
		$(2) $$file ; \
	done) ; \
	if test -n "$$RES"; then \
		echo "$(1) problems:" ; \
		echo "$$RES" ; \
		exit 1 ; \
	fi ;

test: generated vendor
	@CGO_ENABLED=0 go test -short $$(glide nv)

vet:
	@go vet $$(glide nv)

fmtcheck:
	$(call EACH_FILE,gofmt,gofmt -l -s)

simple:
	$(call EACH_FILE,gosimple,gosimple)

lint:
	$(call EACH_FILE,golint,golint)

.PHONY: vet fmtcheck simple lint test

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

docker-release-all:
	$(call RUN_IN_DOCKER,release-all)

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

npm-bin: builds/npm/bin builds/torus-*
	cp builds/torus-* builds/npm/bin/

.PHONY: npm npm-bin
