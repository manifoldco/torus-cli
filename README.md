# Arigato CLI and Daemon

This repository contains the code for both the arigato daemon and cli.

The daemon is written in Go and is responsible for pulling data from the
server, performing authorized actions against the server on behalf of the
client, and performing any crytographic operations. A secondary responsibility
is to maintain the encrypted on-disk cache in the case of a cloud outage.

The CLI is written using Node.js and communicates with the deamon using a local
unix socket only accessible the user running the daemon. It's the nice fluffly
front-end that is responsible for doing fun things like auto completion and
ensuring a great user experience.

### Structure

The top level is responsible for packaging and releasing the cli and daemon.
All component specific code and dependencies live in their respective folders.
