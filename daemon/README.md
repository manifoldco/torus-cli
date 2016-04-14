# Arigato Daemon

Simple go service for storing passwords in guarded memory using libsodium
through c-go. Communicates with clients using a local unix socket.

## Building and Linting

You can build the daemon using make with `make build`. You can install
dependencies using `glide install`.

To lint you can just run `make lint` and to run vet `make vet`.

You must have the CLI repository checked out in your $GOPATH.

For example: `~/code/arigato/src/github.com/arigatomachine/cli`.

## Dependencies

* Go 1.6
* [Glide](https://github.com/Masterminds/glide) for package vendoring
* [libsodium](https://download.libsodium.org/doc/)
* [golint](https://github.com/golang/lint) for linting
