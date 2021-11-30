#!/bin/bash

set -ev

# Install chaincode
docker exec cli peer chaincode install -n recycle -v 1.0 -p github.com/recycle/go

# Instantiate chaincode
docker exec cli peer chaincode instantiate -n recycle -v 1.0 -C mychannel2 -c '{"Args":[]}' -P 'OR ("Org1MSP.member", "Org3MSP.member")'
sleep 5

docker exec cli peer chaincode invoke -n recycle -C mychannel2 -c '{"Args":["initLedger"]}'
sleep 5

# 공장에서 처리해야 하는 재활용 물질의 종류를 등록하는 함수 호출 (registerRecycle)
# 0: CasNo, 1: Phase, 2: Name, 3: Color
docker exec cli peer chaincode invoke -n recycle -C mychannel2 -c '{"Args":["registerRecycle", "scrap001", "Solid", "Carbon Steel(Metal)", "Steel"]}'
sleep 5

# 공장을 가동하면서 쌓이는 재활용 물질을 기록하는 함수 호출 (createRecycle)
# 0: CasNo, 1: Phase, 2: Quantity to add
docker exec cli peer chaincode invoke -n recycle -C mychannel2 -c '{"Args":["createRecycle", "scrap001", "Solid", "500"]}'
sleep 5

# 공장이 보유한 모든 재활용 물질을 조회하는 함수 호출 (queryAllRecycles)
docker exec cli peer chaincode query -n recycle -C mychannel2 -c '{"Args":["queryAllRecycles"]}'

# 재활용 물질 처리 함수 호출 (purgeRecycle)
# 0: CasNo, 1: Phase, 2: Quantity to purge, 3: Cost to purge
docker exec cli peer chaincode invoke -n recycle -C mychannel2 -c '{"Args":["purgeRecycle", "scrap001", "Solid", "300", "1200"]}'
sleep 5

# 공장이 보유한 모든 재활용 물질을 조회하는 함수를 한 번 더 호출해서 재활용 물질 보유 량이 변경되었는지 확인
docker exec cli peer chaincode query -n recycle -C mychannel2 -c '{"Args":["queryAllRecycles"]}'

# 재활용 물질 처리 내역을 조회하는 함수 호출 (queryAllEmitionRecords)
docker exec cli peer chaincode query -n recycle -C mychannel2 -c '{"Args":["queryAllEmitionRecords"]}'

echo '-------------------------------------END-------------------------------------'