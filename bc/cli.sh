#!/bin/bash

function constructbc() {
  echo
  echo "#################### 开 始 创 建 通 道 ####################"
  echo
  set -x
  peer channel create -o orderer.LSH:7050 -c mychannel -f ./channel-artifacts/MyChannel.tx
  res=$?
  set +x
  echo
  if [ $res -eq 0 ]; then
    echo "#################### 创 建 通 道 成 功 ####################"
  else
    echo "!!!!!!!!!!!!!!!!!!!! 创 建 通 道 失 败 !!!!!!!!!!!!!!!!!!!!"
  fi
  echo
  echo
  echo "#################### 节 点 开 始 加 入 通 道 ####################"
  echo
  for ((i = 1; i <= $NodeNum; i++)); do
    joinChannel $i
  done
  echo
  echo "#################### 开 始 安 装 链 码 ####################"
  echo
  for ((i = 1; i <= $NodeNum; i++)); do
    installCC $i
  done
  setEnv 1
  echo
  echo "#################### 正 在 初 始 化 链 码 ####################"
  echo
  inarg='{"Args":["'
  inarg="${inarg}${NodeNum}\"]}"
  set -x
  peer chaincode instantiate -o orderer.LSH:7050 -C mychannel -c "$inarg" -n mycc -v 1.0 --collections-config privconf.json
  res=$?
  set +x
  if [ $res -eq 0 ]; then
    echo "#################### 初 始 化 链 码 成 功 ####################"
  else
    echo "!!!!!!!!!!!!!!!!!!!! 初 始 化 链 码 失 败 !!!!!!!!!!!!!!!!!!!!"
  fi
}

function installCC() {
  setEnv $1
  echo
  echo "#################### LSH${1} 开 始 安 装 链 码 ####################"
  echo
  set -x
  peer chaincode install -n mycc -p chaincode/code02 -v 1.0
  res=$?
  set +x
  echo
  if [ $res -eq 0 ]; then
    echo "#################### LSH${1} 安 装 链 码 成 功 ####################"
  else
    echo "!!!!!!!!!!!!!!!!!!!! LSH${1} 安 装 链 码 失 败 !!!!!!!!!!!!!!!!!!!!"
  fi
}

function joinChannel() {
  setEnv $1
  echo
  echo "#################### LSH${1} 节 点 开 始 加 入 通 道 ####################"
  echo
  set -x
  peer channel join -b mychannel.block
  res=$?
  set +x
  echo
  if [ $res -eq 0 ]; then
    echo "#################### LSH${1} 节 点 加 入 通 道 成 功 ####################"
  else
    echo "!!!!!!!!!!!!!!!!!!!! LSH${1} 节 点 加 入 通 道 失 败 !!!!!!!!!!!!!!!!!!!!"
  fi
  echo
  echo
  echo "#################### LSH${1} 节 点 升 级 为 锚 节 点 ####################"
  echo
  set -x
  peer channel update -o orderer.LSH:7050 -c mychannel -f ./channel-artifacts/LSH${1}.tx
  res=$?
  set +x
  echo
  if [ $res -eq 0 ]; then
    echo "#################### LSH${1} 节 点 升 级 成 功 ####################"
  else
    echo "!!!!!!!!!!!!!!!!!!!! LSH${1} 节 点 升 级 失 败 !!!!!!!!!!!!!!!!!!!!"
  fi
  echo
}

function setEnv() {
  portNum=$(expr $1 + 6)
  CORE_PEER_LOCALMSPID=LSH"$1"MSP
  CORE_PEER_ADDRESS=peer1.LSH"$1":"${portNum}051"
  CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/LSH"$1"/users/Admin@LSH"$1"/msp
}

function initbc() {
  echo
  echo "#################### 各 节 点 开 始 编 译 ####################"
  echo
  for ((i = 2; i <= $NodeNum; i++)); do
    compileCC $i
    res=$?
    if [ $res -ne 0 ]; then
      echo
      echo "!!!!!!!!!!!!!!!!!!!! 编 译 失 败 !!!!!!!!!!!!!!!!!!!!"
      echo
      exit 1
    fi
  done
}

function compileCC() {
  setEnv $1
  echo
  echo "#################### LSH${1} 开 始 编 译 链 码 ####################"
  echo
  set -x
  peer chaincode invoke -C mychannel -n mycc -c '{"Args":["init"]}'
  res=$?
  set +x
  echo
  if [ $res -eq 0 ]; then
    echo "#################### LSH${1} 节 点 编 译 成 功 ####################"
    sleep 1s
  else
    echo "!!!!!!!!!!!!!!!!!!!! LSH${1} 节 点 编 译 失 败 !!!!!!!!!!!!!!!!!!!!"
    exit 1
    sleep 2s
  fi
  echo
}

MODE=$1
NodeNum=$2

case $MODE in
"construct")
  constructbc
  ;;
"init")
  initbc
  ;;
esac
