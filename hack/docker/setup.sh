#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/pharmer/csi-linode

source "$REPO_ROOT/hack/libbuild/common/pharmer_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
IMG=csi-linode

DIST=$GOPATH/src/github.com/pharmer/csi-linode/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
  export $(cat $DIST/.tag | xargs)
fi

clean() {
  pushd $GOPATH/src/github.com/pharmer/csi-linode/hack/docker
  rm csi-linode Dockerfile
  popd
}

build_binary() {
  pushd $GOPATH/src/github.com/pharmer/csi-linode
  ./hack/builddeps.sh
  ./hack/make.py build
  detect_tag $DIST/.tag
  popd
}

build_docker() {
  pushd $GOPATH/src/github.com/pharmer/csi-linode/hack/docker
  cp $DIST/csi-linode/csi-linode-alpine-amd64 csi-linode
  chmod 755 csi-linode

  cat >Dockerfile <<EOL
FROM alpine

RUN set -x \
  && apk add --update --no-cache ca-certificates tzdata e2fsprogs findmnt


ADD csi-linode /usr/bin/csi-linode

ENTRYPOINT ["csi-linode"]

EOL
  local cmd="docker build -t pharmer/$IMG:$TAG ."
  echo $cmd
  $cmd

  rm csi-linode Dockerfile
  popd
}

build() {
  build_binary
  build_docker
}

docker_push() {
  if [ "$APPSCODE_ENV" = "prod" ]; then
    echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"
    exit 0
  fi
  if [ "$TAG_STRATEGY" = "git_tag" ]; then
    echo "Are you trying to 'release' binaries to prod?"
    exit 1
  fi
  hub_canary
}

docker_release() {
  if [ "$APPSCODE_ENV" != "prod" ]; then
    echo "'release' only works in PROD env."
    exit 1
  fi
  if [ "$TAG_STRATEGY" != "git_tag" ]; then
    echo "'apply_tag' to release binaries and/or docker images."
    exit 1
  fi
  hub_up
}

source_repo $@
