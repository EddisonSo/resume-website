#!/bin/bash

HUBU=eddisonso
IMG=eddison-resume

docker buildx create --name mybuilder
docker buildx use mybuilder
docker buildx build --platform linux/amd64,linux/arm64 -t $HUBU/$IMG:latest --push .
