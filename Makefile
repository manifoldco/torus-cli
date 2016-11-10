OUT=torus
PKG=github.com/manifoldco/torus-cli

GO_REQUIRED_VERSION=1.7.1
LINUX=\
	linux-amd64
TARGETS=\
	darwin-amd64 \
	$(LINUX)

VERSION?=$(shell git describe --tags --abbrev=0 | sed 's/^v//')

all: binary
ci: binary vet fmtcheck simple lint misspell ineffassign test

.PHONY: all ci

#################################################
# Bootstrapping for base golang package deps
#################################################

BOOTSTRAP=\
	github.com/Masterminds/glide \
	github.com/golang/lint/golint \
	honnef.co/go/simple/cmd/gosimple \
	github.com/jteeuwen/go-bindata/... \
	github.com/client9/misspell/cmd/misspell \
	github.com/gordonklaus/ineffassign \
	github.com/alecthomas/gometalinter

$(BOOTSTRAP):
	go get -u $@
bootstrap: $(BOOTSTRAP)

.PHONY: bootstrap $(BOOTSTRAP)

#################################################
# Build targets for local usage
#################################################

VERSION_FLAG=-X $(PKG)/config.Version=$(VERSION)
STATIC_FLAGS=-w -s $(VERSION_FLAG)
GO_BUILD=CGO_ENABLED=0 go build -i -v

binary: generated vendor
	$(GO_BUILD) -o ${OUT} -ldflags='$(VERSION_FLAG)' ${PKG}

static: generated vendor
	$(GO_BUILD) -o ${OUT}-v${VERSION} -ldflags='$(STATIC_FLAGS)' ${PKG}

.PHONY: binary static

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

PRIMITIVE_BOILERPLATE=tools/primitive-boilerplate
$(TOOLS)/primitive-boilerplate: $(wildcard $(PRIMITIVE_BOILERPLATE)/*.go) $(wildcard $(PRIMITIVE_BOILERPLATE)/*.tmpl)
	$(GO_BUILD) -o $@ ./$(PRIMITIVE_BOILERPLATE)

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

misspell:
	@gometalinter --disable-all --vendor --enable=misspell ./...

ineffassign:
	@gometalinter --disable-all --vendor --enable=ineffassign ./...


.PHONY: vet fmtcheck simple lint misspell ineffassign test

#################################################
# Docker targets
#################################################

PWD=$(shell pwd)
IMAGE=manifoldco/torus-cli:latest
RUN_IN_DOCKER=\
	docker run --name cli --rm \
		-v $(PWD):/go/src/github.com/manifoldco/torus-cli \
		-v $(PWD)/builds:/builds \
		$(IMAGE) $(1)

docker-build:
	$(call RUN_IN_DOCKER,binary)

docker-test:
	$(call RUN_IN_DOCKER,ci)

docker-release-all:
	$(call RUN_IN_DOCKER,release-all)

container:
	docker build -t $(IMAGE) .

rpm-container:
	docker build -t manifoldco/torus-rpm packaging/rpm

deb-container:
	docker build -t manifoldco/torus-deb packaging/deb

.PHONY: docker-build docker-test container rpm-container deb-container

#################################################
# Build targets for releasing
#################################################

RELEASE_ENV?=stage
ifeq (stage,$(RELEASE_ENV))
	TORUS_S3_BUCKET=s3://releases.arigato.sh
	BOTTLE_URL=https://s3.amazonaws.com/releases.arigato.sh
else ifeq (prod,$(RELEASE_ENV))
	TORUS_S3_BUCKET=s3://get.torus.sh
	BOTTLE_URL=https://get.torus.sh
endif

tagcheck:
ifneq (v$(VERSION),$(shell git describe --tags --dirty))
	$(error "VERSION $(VERSION) is not git HEAD")
endif

envcheck:
ifeq (,$(TORUS_S3_BUCKET))
	$(error "Unknown RELEASE_ENV $(RELEASE_ENV)")
endif
ifeq (prod,$(RELEASE_ENV))
ifneq (,$(findstring -rc,$(VERSION)))
	$(error "You can't release an rc version to prod")
endif
endif
ifneq (yes,$(RELEASE_CONFIRM))
	$(error "Set RELEASE_CONFIRM=yes to really release")
endif

gocheck:
ifeq (,$(findstring $(GO_REQUIRED_VERSION),$(shell go version)))
ifeq (,$(BYPASS_GO_CHECK))
	$(error "Go Version $(GO_REQUIRED_VERSION) is required.")
endif
endif

OS=$(word 1, $(subst -, ,$*))
ARCH=$(word 2, $(subst -, ,$*))
BUILD_DIR=builds/bin/$(VERSION)/$(OS)/$(ARCH)
BINARY=-o $(BUILD_DIR)/$(OUT)
$(addprefix binary-,$(TARGETS)): binary-%: gocheck generated vendor
	GOOS=$(OS) GOARCH=$(ARCH) $(GO_BUILD) $(BINARY) \
		-ldflags='$(STATIC_FLAGS)' ${PKG}

BUILD_DIRS=\
	builds/dist/$(VERSION) \
	builds/dist/rpm \
	builds/deb \
	builds/dist/ubuntu \
	builds/dist/debian \
	builds/dist/brew/$(VERSION) \
	builds/dist/npm/$(VERSION) \
	builds/dist/brew/bottles
$(BUILD_DIRS):
	@mkdir -p $@

$(addprefix zip-,$(TARGETS)): zip-%: binary-% builds/dist/$(VERSION)
	zip -j builds/dist/$(VERSION)/$(OUT)_$(VERSION)_$(OS)_$(ARCH).zip \
		$(BUILD_DIR)/$(OUT)

release-binary: $(addprefix zip-,$(TARGETS))
	pushd builds/dist/$(VERSION) && \
		shasum -a 256 *.zip > $(OUT)_$(VERSION)_SHA256SUMS

$(addprefix rpm-,$(LINUX)): rpm-%: binary-% builds/dist/rpm rpm-container
	docker run -v $(PWD):/torus manifoldco/torus-rpm /bin/bash -c " \
		rpmbuild -D '_sourcedir /torus' \
			-D 'VERSION $(subst -,_,$(VERSION))' \
			-D 'REAL_VERSION $(VERSION)' \
			-D 'ARCH $(ARCH)' \
			-bb packaging/rpm/torus.spec && \
		cp -R ~/rpmbuild/RPMS/* /torus/builds/dist/rpm/ \
	"

$(addprefix yum-,$(LINUX)): yum-%: rpm-%
	docker run -v $(PWD):/torus manifoldco/torus-rpm /bin/bash -c " \
		cd builds/dist/rpm/x86_64/ && \
		createrepo_c . \
	"

$(addprefix deb-,$(LINUX)): deb-%: binary-% builds/deb deb-container
	docker run -v $(PWD):/torus manifoldco/torus-deb /bin/bash -c " \
		mkdir -p deb-tmp/torus/DEBIAN && \
		mkdir -p deb-tmp/torus/usr/bin && \
		cp /torus/builds/bin/$(VERSION)/$(OS)/$(ARCH)/torus \
			deb-tmp/torus/usr/bin/ && \
		sed 's/VERSION/$(VERSION)/' < /torus/packaging/deb/control.in | \
			sed 's/ARCH/$(ARCH)/' > deb-tmp/torus/DEBIAN/control && \
		cd deb-tmp && \
		dpkg-deb -b torus && \
		cp torus.deb /torus/builds/deb/torus_$(VERSION)_$(ARCH).deb \
	"

builds/dist/debian/conf/distributions: packaging/deb/distributions.debian
	mkdir -p $(@D)
	cp $< $@

builds/dist/ubuntu/conf/distributions: packaging/deb/distributions.ubuntu
	mkdir -p $(@D)
	cp $< $@

apt-repo: $(addprefix deb-,$(LINUX)) builds/dist/debian/conf/distributions builds/dist/ubuntu/conf/distributions
	docker run -v $(PWD):/torus manifoldco/torus-deb /bin/bash -c " \
		cd /torus/builds/dist/debian && \
		reprepro includedeb jessie /torus/builds/deb/*.deb && \
		cd /torus/builds/dist/ubuntu && \
		reprepro includedeb xenial /torus/builds/deb/*.deb && \
		reprepro includedeb trusty /torus/builds/deb/*.deb \
	"

GIT_SHA=$(shell curl -L https://github.com/manifoldco/torus-cli/archive/v$(VERSION).tar.gz | shasum -a 256 | cut -d" " -f1)
BOTTLE_SHA=$(shell shasum -a 256 builds/torus-$(VERSION).sierra.bottle.tar.gz | cut -d" " -f 1)
builds/torus-$(VERSION).rb: packaging/homebrew/torus.rb.in builds/torus-$(VERSION).sierra.bottle.tar.gz
	sed 's/{{VERSION}}/$(VERSION)/' < $< | \
		sed 's/{{SHA256}}/$(GIT_SHA)/' | \
		sed 's/{{BOTTLE_SHA256}}/$(BOTTLE_SHA)/' | \
		sed 's|{{BOTTLE_URL}}|$(BOTTLE_URL)/brew/bottles|' > $@

builds/bottle/torus/$(VERSION)/INSTALL_RECEIPT.json: packaging/homebrew/INSTALL_RECEIPT.json.in
	mkdir -p builds/bottle/torus/$(VERSION)
	sed 's/{{VERSION}}/$(VERSION)/' < $< | \
		sed 's/{{GO_VERSION}}/$(GO_REQUIRED_VERSION)/' | \
		sed 's/{{MTIME}}/$(shell git log -n1 --format=%at v$(VERSION))/' > $@

builds/torus-$(VERSION).sierra.bottle.tar.gz: binary-darwin-amd64 builds/bottle/torus/$(VERSION)/INSTALL_RECEIPT.json
	mkdir -p builds/bottle/torus/$(VERSION)/bin
	cp builds/bin/0.16.0/darwin/amd64/torus builds/bottle/torus/$(VERSION)/bin
	tar zcf $@ -C builds/bottle torus

release-homebrew: envcheck tagcheck release-homebrew-bottle release-homebrew-$(RELEASE_ENV)

BOTTLE_VERSIONS=$(foreach release,sierra el_capitan yosemite,builds/dist/brew/bottles/torus-$(VERSION).$(release).bottle.tar.gz)
$(BOTTLE_VERSIONS): builds/torus-$(VERSION).sierra.bottle.tar.gz builds/dist/brew/bottles
	cp $< $@

release-homebrew-bottle: $(BOTTLE_VERSIONS)

release-homebrew-stage: builds/torus-$(VERSION).rb builds/dist/brew/$(VERSION)
	cp $< builds/dist/brew/$(VERSION)/torus.rb

builds/homebrew-git:
	git clone --depth=1 git@github.com:manifoldco/homebrew-brew.git \
		builds/homebrew-git
homebrew-git: builds/homebrew-git
	cd builds/homebrew-git && git pull

release-homebrew-prod: builds/torus-$(VERSION).rb homebrew-git
	cp $< builds/homebrew-git/Formula/torus.rb
	pushd builds/homebrew-git && \
		git add Formula/torus.rb && \
		git commit -m "Update torus to v$(VERSION)" && \
		git push origin master

release-npm: envcheck tagcheck release-npm-$(RELEASE_ENV)

release-npm-stage: builds/torus-npm-$(VERSION).tar.gz builds/dist/npm/$(VERSION)
	cp $< builds/dist/npm/$(VERSION)/torus.tar.gz

release-npm-prod: builds/torus-npm-$(VERSION).tar.gz
	npm publish $<

RELEASE_TARGETS=\
	release-binary \
	release-npm \
	release-homebrew \
	apt-repo \
	$(addprefix yum-,$(LINUX))
release-all: envcheck tagcheck $(RELEASE_TARGETS)
	pushd builds/dist && aws s3 cp --recursive . $(TORUS_S3_BUCKET)

.PHONY: envcheck tagcheck gocheck release-all release-binary
.PHONY: $(addprefix binary-,$(TARGETS)) $(addprefix zip-,$(TARGETS))
.PHONY: $(addprefix yum-,$(TARGETS)) $(addprefix rpm-,$(TARGETS))
.PHONY: release-npm-stage release-npm-prod

#################################################
# Distribution via npm
#################################################

NPM_DEPS=\
	builds/npm/package.json \
	builds/npm/README.md \
	builds/npm/LICENSE.md \
	builds/npm/scripts/install.js \
	builds/npm/bin/torus-darwin-amd64 \
	builds/npm/bin/torus-linux-amd64
npm: $(NPM_DEPS)

builds/npm builds/npm/bin builds/npm/scripts:
	mkdir -p $@

builds/npm/README.md builds/npm/LICENSE.md: builds/npm/%: builds/npm
	cp $* $@

builds/npm/package.json: packaging/npm/package.json.in builds/npm
	sed 's/VERSION/$(VERSION)/' < $< > $@

builds/npm/scripts/install.js: packaging/npm/install.js builds/npm/scripts
	cp $< $@

builds/npm/bin/torus-darwin-amd64: builds/bin/$(VERSION)/darwin/amd64/torus builds/npm/bin
	cp $< $@

builds/npm/bin/torus-linux-amd64: builds/bin/$(VERSION)/linux/amd64/torus builds/npm/bin
	cp $< $@

builds/torus-npm-$(VERSION).tar.gz: npm
	tar czf $@ -C builds npm/

.PHONY: npm
