#!/bin/bash -e

CURDIR="$(dirname $0)"
TAG_DEST=$(git rev-parse --short HEAD)
export CLOUDSDK_CORE_PROJECT=eoscanada-shared-services

pushd ${CURDIR}/.. >/dev/null
  PROJDIR="$(basename $(pwd))"
popd >/dev/null

pushd ${CURDIR} >/dev/null
  cloud-build-local \
        --config cloudbuild.yaml \
        --dryrun=false \
        --substitutions=TAG_NAME=${TAG_DEST},_APP=${PROJDIR} \
        ..
popd >/dev/null
echo "TAG_DEST: ${TAG_DEST}"

