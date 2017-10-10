# torus-cli

A secure, shared workspace for secrets.

[Homepage](https://torus.sh) |
[Documentation](https://torus.sh/docs) |
[Twitter](https://twitter.com/toruscli) |
[Security Disclosure](./internal/security.md) |
[Code of Conduct](./.github/CONDUCT.md) |
[Contribution Guidelines](./.github/CONTRIBUTING.md)

[![Travis](https://img.shields.io/travis/manifoldco/torus-cli/master.svg)](https://travis-ci.org/manifoldco/torus-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/manifoldco/torus-cli)](https://goreportcard.com/report/github.com/manifoldco/torus-cli)
[![npm](https://img.shields.io/npm/v/torus-cli.svg)](https://www.npmjs.com/package/torus-cli)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](./LICENSE.md)

![](./graphic.png)

## Installation & signup

[Manifold](https://www.manifold.co) provides binaries of `torus-cli` for OS X, Linux and Windows on `amd64`.

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
- DEB based distributions: Use the following repository configuration:
```
DISTRO=$(lsb_release -i | awk '{print tolower($3)}')
CODENAME=$(lsb_release -c | awk '{print $2}')
sudo tee /etc/apt/sources.list.d/torus.list <<< "deb https://get.torus.sh/$DISTRO/ $CODENAME main"
```
- [npm](https://www.npmjs.com): `npm install -g torus-cli`
- bare zip archives per release version are available on https://get.torus.sh/

### Windows (Alpha)

Install torus via npm using `npm install -g torus-cli` or manally using the
steps below!

- Get the desired version on [https://get.torus.sh/](https://get.torus.sh/)
- Unzip the file
- Put the `torus.exe` file in your path
  - **System Settings**
  - **Advanced System Settings**
  - **Advanced**
  - **Environment Variables**
  - Edit **Path** in **System Variables** and add the full path to the folder where your `torus.exe` file is

#### Security note

Currently on Windows, the Daemon will create a named pipe using the default security attributes. This means, that the LocalSystem account, administrators, and the creator will be granted full control. All members of the Everyone group and the anonymous account are granted read access.

More information can be found [here](https://msdn.microsoft.com/en-us/library/windows/desktop/aa365150(v=vs.85).aspx).

## Security Disclosure

Please follow our security disclosure document [found here](./internal/security.md).

## License

Manifold's torus-cli is released under the [BSD 3-Clause License](./LICENSE.md).
