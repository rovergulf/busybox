#!/usr/bin/env bash

set -e

RUN_STARTED=$(date)
REGISTRY_IMAGE="${OCI_MAIN_ROVERGULF}/busybox"
VERSION=$(cat ./version | awk '{print $1}')

echo "[$(date)] Start building ${REGISTRY_IMAGE}:${VERSION} docker image"
docker build --no-cache -t $REGISTRY_IMAGE:$VERSION -t $REGISTRY_IMAGE:latest . || exit 1


echo "[$(date)] push image to registry"
docker push $REGISTRY_IMAGE:$VERSION || exit 2
docker push $REGISTRY_IMAGE:latest || exit 3

echo "[$(date)] Successfully pushed registry image. Run started at [${RUN_STARTED}]"
