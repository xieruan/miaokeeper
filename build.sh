#!/bin/bash

COMMIT=$(git rev-list -1 HEAD)
GOOS="linux" GOARCH="amd64" go build -ldflags "-s -w" .
tar zcvf miaokeeper.tgz miaokeeper
rm miaokeeper

mc mv miaokeeper.tgz sideload/miaoko/miaokeeper.tgz