#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -ev

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

cp docker-compose-template.yml docker-compose.yml
CA1_PRIVATE_KEY=$(ls crypto-config/peerOrganizations/org1.example.com/ca/ | grep _sk)
CA2_PRIVATE_KEY=$(ls crypto-config/peerOrganizations/org2.example.com/ca/ | grep _sk)
CA3_PRIVATE_KEY=$(ls crypto-config/peerOrganizations/org3.example.com/ca/ | grep _sk)
sed -i "s/CA1_PRIVATE_KEY/$CA1_PRIVATE_KEY/g" docker-compose.yml
sed -i "s/CA2_PRIVATE_KEY/$CA2_PRIVATE_KEY/g" docker-compose.yml
sed -i "s/CA3_PRIVATE_KEY/$CA3_PRIVATE_KEY/g" docker-compose.yml

docker-compose -f docker-compose.yml down

docker-compose -f docker-compose.yml up -d ca1.example.com ca2.example.com ca3.example.com orderer.example.com peer0.org1.example.com peer0.org2.example.com peer0.org3.example.com couchdb1 couchdb2 couchdb3 cli
docker ps -a

# wait for Hyperledger Fabric to start
# incase of errors when running later commands, issue export FABRIC_START_TIMEOUT=<larger number>
export FABRIC_START_TIMEOUT=10
#echo ${FABRIC_START_TIMEOUT}
sleep ${FABRIC_START_TIMEOUT}

# Create the channel1
docker exec cli peer channel create -o orderer.example.com:7050 -c mychannel1 -f /etc/hyperledger/configtx/channel1.tx
sleep 3

# Create the channel2
docker exec cli peer channel create -o orderer.example.com:7050 -c mychannel2 -f /etc/hyperledger/configtx/channel2.tx
sleep 3

# Join peer0.org1.example.com to the channel.
docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com peer channel join -b /etc/hyperledger/configtx/mychannel1.block
sleep 3
docker exec -e "CORE_PEER_LOCALMSPID=Org2MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org2.example.com/msp" peer0.org2.example.com peer channel join -b /etc/hyperledger/configtx/mychannel1.block
sleep 3

docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com peer channel join -b /etc/hyperledger/configtx/mychannel2.block
sleep 3
docker exec -e "CORE_PEER_LOCALMSPID=Org3MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org3.example.com/msp" peer0.org3.example.com peer channel join -b /etc/hyperledger/configtx/mychannel2.block
sleep 3

# Anchor peer tx
docker exec -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com peer channel update -f /etc/hyperledger/configtx/Org1MSPanchorsInChannel1.tx -o orderer.example.com:7050 -c mychannel1 
sleep 3
docker exec -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org2.example.com/msp" peer0.org2.example.com peer channel update -f /etc/hyperledger/configtx/Org2MSPanchorsInChannel1.tx -o orderer.example.com:7050 -c mychannel1 
sleep 3
docker exec -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.example.com/msp" peer0.org1.example.com peer channel update -f /etc/hyperledger/configtx/Org1MSPanchorsInChannel2.tx -o orderer.example.com:7050 -c mychannel2 
sleep 3
docker exec -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org3.example.com/msp" peer0.org3.example.com peer channel update -f /etc/hyperledger/configtx/Org3MSPanchorsInChannel2.tx -o orderer.example.com:7050 -c mychannel2
sleep 3