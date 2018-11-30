# Docker Machine driver for RancherVM

[![Latest Version](https://img.shields.io/github/release/llparse/docker-machine-driver-ranchervm.svg?maxAge=8600)][release]
[![License](https://img.shields.io/github/license/llparse/docker-machine-driver-ranchervm.svg?maxAge=8600)]()

## Prerequisites

* [Docker Machine](https://github.com/docker/machine/releases) v0.5.1 or later.

## Installation

For Unix systems, run the following command:

```console
NAME=docker-machine-driver-ranchervm && OS=`uname -s` && ARCH=`uname -m` && INSTALLPATH=/usr/local/bin/${NAME} && \
VERSION=`curl -sSfL https://raw.githubusercontent.com/llparse/${NAME}/master/VERSION` && \
curl -sSfL https://github.com/llparse/${NAME}/releases/download/${VERSION}/${NAME}-${OS}-${ARCH} \
  -o /tmp/${NAME} && chmod +x /tmp/${NAME} && sudo mv /tmp/${NAME} ${INSTALLPATH}
```

For other platforms, download the binary from the release page or build it yourself.
