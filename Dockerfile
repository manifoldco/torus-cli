FROM golang:1.7.1-alpine

RUN apk add --no-cache bash git make

# Create work directory for the CLI and build output dest
RUN mkdir -p /go/src/github.com/arigatomachine/cli \
        && mkdir -p /builds \

ENV PATH="/usr/local/bin:$PATH"

# Now install our go specific dependencies
RUN go get -u github.com/Masterminds/glide \
        && go get -u github.com/golang/lint/golint

VOLUME /go/src/github.com/arigatomachine/cli
VOLUME /builds

WORKDIR /go/src/github.com/arigatomachine/cli

ENTRYPOINT ["/bin/bash", "./docker/init.sh"]
CMD ["test"]

STOPSIGNAL SIGTERM
