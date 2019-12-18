#!/bin/bash

function printHelp() {
  echo "使用方法：$0 [option] [NodeNumber]"
  echo "option选项有："
  echo "generate：生成区块链需要的文件"
  echo "up：启动区块链"
  echo "down：关闭区块链"
  echo "NodeNumber是要加入区块链的节点个数，必须为整数数字，不小于2且不大于20"
}

function generatebc() {
  if [ -e "./crypto-config" ] && [ -d "./crypto-config" ]; then
    echo "文件已存在，不需要创建"
    return 0
  fi
  echo
  echo "####################正在创建证书和交易文件####################"
  echo
  cryptogen generate --config=./crypto_config.yaml
  configtxgen -profile Genesis -outputBlock ./channel-artifacts/genesis.block
  configtxgen -profile MyChannel -outputCreateChannelTx ./channel-artifacts/MyChannel.tx -channelID mychannel
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH1.tx -channelID mychannel -asOrg LSH1
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH2.tx -channelID mychannel -asOrg LSH2
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH3.tx -channelID mychannel -asOrg LSH3
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH4.tx -channelID mychannel -asOrg LSH4
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH5.tx -channelID mychannel -asOrg LSH5
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH6.tx -channelID mychannel -asOrg LSH6
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH7.tx -channelID mychannel -asOrg LSH7
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH8.tx -channelID mychannel -asOrg LSH8
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH9.tx -channelID mychannel -asOrg LSH9
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH10.tx -channelID mychannel -asOrg LSH10
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH11.tx -channelID mychannel -asOrg LSH11
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH12.tx -channelID mychannel -asOrg LSH12
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH13.tx -channelID mychannel -asOrg LSH13
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH14.tx -channelID mychannel -asOrg LSH14
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH15.tx -channelID mychannel -asOrg LSH15
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH16.tx -channelID mychannel -asOrg LSH16
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH17.tx -channelID mychannel -asOrg LSH17
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH18.tx -channelID mychannel -asOrg LSH18
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH19.tx -channelID mychannel -asOrg LSH19
  configtxgen -profile MyChannel -outputAnchorPeersUpdate ./channel-artifacts/LSH20.tx -channelID mychannel -asOrg LSH20
  echo
  echo "####################创建证书和交易文件完毕####################"
  echo
}

function upthebc() {
  echo
  echo "####################正在创建docker容器####################"
  echo
  countainernum=$(expr 2 + "$NodeNum")
  countainers="orderer.LSH cli"
  for (( i = 1; i <= NodeNum; i++ )); do
      countainers="$countainers peer1.LSH$i"
  done
  docker-compose -f /home/forsim/go/LSH/bc/docker-compose.yaml up -d $countainers
  docker_stat=$?
  # 等1秒看看起了多少个容器
  sleep 1s
  docker_numb=$(docker ps -aq | wc -l)
  # 应该有22个容器成功运行起来的，如果没有说明出了问题
  if [ "$docker_numb" -ne $countainernum ] || [ "$docker_stat" -ne 0 ]; then
    echo "docker_numb=$docker_numb"
    echo "containernum=$countainernum"
    echo "docker_stat=$docker_stat"
    echo
    echo "!!!!!!!!!!!!!!!!!!!!docker容器创建失败!!!!!!!!!!!!!!!!!!!!"
    echo
    exit 1
  fi
  set -x
  docker ps
  set +x
  echo
  echo "####################docker容器创建成功####################"
  echo
  docker exec cli ./cli.sh construct $NodeNum
  sleep 5s
  docker exec cli ./cli.sh init $NodeNum
}

function downthebc() {
  echo
  echo "####################正在关闭docker容器与网络####################"
  echo
  docker-compose -f /home/forsim/go/LSH/bc/docker-compose.yaml down -v
  CONTAINER_IDS=$(docker ps -a | awk '($2 ~ /dev-peer.*/) {print $1}')
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == "" ]; then
    echo
    echo "####################没有发现链码容器####################"
    echo
  else
    docker rm -f $CONTAINER_IDS
  fi
  # 每次实例化链码都会生成新的docker镜像，必须把镜像删掉，否则以后实例化同名的链码将不会被编译
  # 所有实例化链码的镜像都是以"dev-peer."开头的
  DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*/) {print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == "" ]; then
    echo
    echo "####################没有发现链码镜像####################"
    echo
  else
    docker rmi -f $DOCKER_IMAGE_IDS
  fi
  echo
  echo "####################成功关闭docker容器与网络####################"
  echo
}

MODE="$1"
NodeNum="$2"
if [[ $NodeNum -eq "" ]]; then
  NodeNum=20
fi

case $MODE in
"generate")
  generatebc
  ;;
"up")
  upthebc
  ;;
"down")
  downthebc
  ;;
*)
  printHelp
  ;;
esac
