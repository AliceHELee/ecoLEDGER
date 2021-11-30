#!/bin/bash

set -ev

# Install chaincode
docker exec cli peer chaincode install -n waste -v 1.0 -p github.com/waste/go

# Instantiate chaincode
docker exec cli peer chaincode instantiate -n waste -v 1.0 -C mychannel1 -c '{"Args":[]}' -P 'OR ("Org1MSP.member", "Org2MSP.member")'
sleep 5

docker exec cli peer chaincode invoke -n waste -C mychannel1 -c '{"Args":["initLedger"]}'
sleep 5

# 공장에서 처리해야 하는 폐기물의 종류를 등록하는 함수 호출 (registerWaste)
# 0: CasNo, 1: Phase, 2: Name, 3: Color
docker exec cli peer chaincode invoke -n waste -C mychannel1 -c '{"Args":["registerWaste", "506-77-4", "Gas", "Cyanogen chioride", "None"]}'
sleep 5

# 공장을 가동하면서 쌓이는 폐기물을 기록하는 함수 호출 (createWaste)
# 0: CasNo, 1: Phase, 2: Quantity to add
docker exec cli peer chaincode invoke -n waste -C mychannel1 -c '{"Args":["createWaste", "506-77-4", "Gas", "500"]}'
sleep 5

# 공장이 보유한 모든 폐기물을 조회하는 함수 호출 (queryAllWastes)
docker exec cli peer chaincode query -n waste -C mychannel1 -c '{"Args":["queryAllWastes"]}'

# 폐기물 처리 함수 호출 (purgeWaste)
# 0: CasNo, 1: Phase, 2: Quantity to purge, 3: Cost to purge
docker exec cli peer chaincode invoke -n waste -C mychannel1 -c '{"Args":["purgeWaste", "506-77-4", "Gas", "300", "1700"]}'
sleep 5

# 공장이 보유한 모든 폐기물을 조회하는 함수를 한 번 더 호출해서 폐기물 보유 량이 변경되었는지 확인
docker exec cli peer chaincode query -n waste -C mychannel1 -c '{"Args":["queryAllWastes"]}'

# 폐기물 처리 내역을 조회하는 함수 호출 (queryAllEmitionRecords)
docker exec cli peer chaincode query -n waste -C mychannel1 -c '{"Args":["queryAllEmitionRecords"]}'

echo '-------------------------------------END-------------------------------------'