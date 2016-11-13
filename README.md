# torus-cli

A secure, shared workspace for secrets.

[Homepage](https://torus.sh) |
[Documentation](https://torus.sh/docs) |
[Twitter](https://twitter.com/toruscli) |
[Security Disclosure](./docs/security.md) |
[Code of Conduct](./.github/CONDUCT.md) |
[Contribution Guidelines](./.github/CONTRIBUTING.md)

[![Travis](https://img.shields.io/travis/manifoldco/torus-cli/master.svg)](https://travis-ci.org/manifoldco/torus-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/manifoldco/torus-cli)](https://goreportcard.com/report/github.com/manifoldco/torus-cli)
[![npm](https://img.shields.io/npm/v/torus-cli.svg)](https://www.npmjs.com/package/torus-cli)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](./LICENSE.md)

![](./graphic.png)

## Installation & signup

[Manifold](https://www.manifold.co) provides binaries of `torus-cli` for OS X and Linux on `amd64`.

After installing, create an account with:
```
torus signup
```

### OS X

- [homebrew](http://brew.sh): `brew install manifoldco/brew/torus`
- [npm](https://www.npmjs.com): `npm install -g torus-cli`
- bare zip archives per release version are available on https://get.torus.sh/

### Linux

- RPM based distributions: Use the following repository configuration:
```
$ sudo tee /etc/yum.repos.d/torus.repo <<-'EOF'
[torus]
name=torus-cli repository
baseurl=https://get.torus.sh/rpm/$basearch/
enabled=1
gpgcheck=0
EOF
```
- [npm](https://www.npmjs.com): `npm install -g torus-cli`
- bare zip archives per release version are available on https://get.torus.sh/

## Security Disclosure

Please follow our security disclosure document [found here](./docs/security.md).

## License

Manifold's torus-cli is released under the [BSD 3-Clause License](./LICENSE.md).
