#!/bin/bash

docker build -t hyperledger/fabric-tools-new:latest -f ./tool/Dockerfile .
docker build -t hyperledger/fabric-ccenv-new:latest -f ./ccenv/Dockerfile .

# 替换原本的ccenv
docker tag hyperledger/fabric-ccenv:latest hyperledger/fabric-ccenv-backup:latest
# 偷梁换柱
docker tag hyperledger/fabric-ccenv-new:latest hyperledger/fabric-ccenv:latest