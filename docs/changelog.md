# Example Changelog

This document contains an example changelog which is curated during the release
process of the CLI.

The template follows below:

### v0.4.0

#### Breaking Changes

- Generating policies via. allow/deny will require  >= v0.4.0.

#### Notable Changes

- Added feedback messages when generating a keypair or encrypting a secret.
- Added the ability to view members of a team and to remove them using ag teams:members and ag teams:remove.

#### Performance Improvements

- Reduced cpu usage of loading spinner by 0.5%

#### Bug Fixes

- If the CLI cancels mid-operation the daemon now cancels its on-going crypto operations.
- The CLI no longer checks the file permissions of the .arigato.json file
