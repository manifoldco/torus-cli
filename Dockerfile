FROM golang:1.7.1-alpine

RUN apk add --no-cache git make

# Create work directory for the CLI and build output dest
RUN mkdir -p /go/src/github.com/arigatomachine/cli \
        && mkdir -p /builds \

ENV PATH="/usr/local/bin:$PATH"

# Now install our go specific dependencies
COPY Makefile /
RUN make -f /Makefile bootstrap && rm /Makefile

VOLUME /go/src/github.com/arigatomachine/cli
VOLUME /builds

WORKDIR /go/src/github.com/arigatomachine/cli

ENTRYPOINT ["make"]
CMD ["ci"]

STOPSIGNAL SIGTERM
