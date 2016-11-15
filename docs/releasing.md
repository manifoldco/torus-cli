# Releasing

Releases are coordinated using a "release issue" which tracks the current RC
and the state of all manual qa. Our desired state is for the role of "release
manager" to pass between maintainers ensuring *everyone* is capable of
releasing.

A release manager is responsible for:

- Creating the release issue
- Coordinating the tagging of release candidates and deployment
- Tracking and triaging bugs; while coordinating fixes for blockers
- Curating the change log
- Signing and pubishing the release.

The release manager role rotates on a per-release basis.

**The Flow**:

1. Create a release issue containing the targeted semver versions *and* current
   RC status using the [release template](./release-issue.md).
2. Curate a changelist against the current "stable" version using the
   [changelog template](../changelog.md).
3. Tag release candidates for the targeted components (registry, cli, etc) and
   make them available (deploy/distribute).
4. Execute manual [qa checklist](./qa.md). If bugs are found, track bugs in the checklist
   by linking to the bug issue. Repeat step 2-4 until checklist passes.
5. Tag production releases and deploy to hosted registry. Build the CLI for
   production and publish to npm with an appropriate tag.

**Releasing the CLI**

TL;DR:

The following command will build the v0.2.0 tag for production.

```
git checkout v0.2.0
make clean release-all RELEASE_CONFIRM=yes RELEASE_ENV=prod
```

You will need the following:

- Docker
- AWS SDK installed locally (e.g. `aws-cli/1.10.56 Python/2.7.10 Darwin/15.6.0 botocore/1.4.46`)
- Correct AWS environment variables set (`aws iam get-user` is successful)
  - Note: this includes `AWS_DEFAULT_REGION=us-east-1`
- You belong to the CLIDevelopers group on AWS

The steps for packaging the CLI:

- Make sure you've pulled latest master
- Tag master (e.g. `git tag v0.5.0-rc`)
- Push master and the tag to github (`git push --tags origin maser`)
- Checkout the ref you want to release `git checkout v0.17.0-rc`
- Run `make release-all RELEASE_CONFIRM=yes RELEASE_ENV=[stage|prod]`
- Release artifacts publish to s3:
  - the releases.arigato.sh for prerelease/stage
  - get.torus.sh (with cloudfront cdn) for prod
- NPM and brew only publish to *real* NPM/our brew repo in prod. For stage they
  are stored on the CDN.
