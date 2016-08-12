FROM golang:1.7rc6-wheezy

ARG NODEJS_VERSION=v4.4.7
ENV NODEJS_VERSION=$NODEJS_VERSION
ENV NODEJS_TAR_NAME=node-$NODEJS_VERSION-linux-x64.tar.gz
ENV NODEJS_DOWNLOAD_URL=https://nodejs.org/dist/$NODEJS_VERSION/$NODEJS_TAR_NAME
ENV NODEJS_SHASUM_URL=https://nodejs.org/dist/$NODEJS_VERSION/SHASUMS256.txt.asc

# gpg keys listed at https://github.com/nodejs/node
RUN set -ex \
    && for key in \
        9554F04D7259F04124DE6B476D5A82AC7E37093B \
        94AE36675C464D64BAFA68DD7434390BDBE9B9C5 \
        0034A06D9D9B0064CE8ADF6BF1747F4AD2306D93 \
        FD3A5288F042B6850C66B31F09FE44734EB7990E \
        71DCFD284A79C3B38668286BC97EC7A07EDE3FC1 \
        DD8F2338BAE7501E3DD5AC78C273792F7D83545D \
        B9AE9905FFD7803F25714661B63B535A4C206CA9 \
        C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8 \
    ; do \
        gpg --keyserver ha.pool.sks-keyservers.net --recv-keys "$key"; \
    done

# Create work directory for the CLI, ag-app user, and download+verify+install
# nodejs
RUN mkdir -p /go/src/github.com/arigatomachine/cli \
        && mkdir -p /builds \
        && curl -fsSL "$NODEJS_DOWNLOAD_URL" -o $NODEJS_TAR_NAME  \
        && curl -SLO "$NODEJS_SHASUM_URL" \
        && gpg --batch --decrypt --output SHASUMS256.txt SHASUMS256.txt.asc \
        && grep " $NODEJS_TAR_NAME\$" SHASUMS256.txt | sha256sum -c - \
        && tar --strip-components 1 -C /usr/local -xzf $NODEJS_TAR_NAME \
        && rm $NODEJS_TAR_NAME

ENV PATH="/usr/local/bin:/go/src/github.com/arigatomachine/cli/cli/node_modules/.bin:$PATH"

# Now install our go specific dependencies
RUN go get -u github.com/Masterminds/glide \
        && go get -u github.com/golang/lint/golint

VOLUME /go/src/github.com/arigatomachine/cli
VOLUME /builds

WORKDIR /go/src/github.com/arigatomachine/cli

ENTRYPOINT ["/bin/bash", "./docker/init.sh"]
CMD ["test"]

STOPSIGNAL SIGTERM
