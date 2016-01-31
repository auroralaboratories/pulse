#!/usr/bin/env bash

PROJECT="${PROJECT:-$(basename $PWD)}"
ORG_PATH="github.com/auroralaboratories"
REPO_PATH="${ORG_PATH}/${PROJECT}"

export GOPATH=${PWD}/gopath
export PATH="$GOPATH/bin:$PATH"

rm -f $GOPATH/src/${REPO_PATH}
mkdir -p $GOPATH/src/${ORG_PATH}
ln -s ${PWD} $GOPATH/src/${REPO_PATH}

eval $(go env)

if [ -s DEPENDENCIES ]; then
  echo 'Processing dependencies...'
  for f in $(find . -name DEPENDENCIES); do
    for d in $(cat ${f} | sort | uniq); do
      go get $d
    done
  done
fi

case "$1" in
test)
  go test -test.v
  ;;
*)
  echo "Unknown action '${1}'"
  exit 1
  ;;
esac

# vendor the dependencies
echo 'Vendoring...'
# remove all .git directories except the local projects (that would be bad :)
find gopath -type d | grep -v "${REPO_PATH}" | grep -v ^\./\.git$ | grep \.git$ | xargs rm -rf
