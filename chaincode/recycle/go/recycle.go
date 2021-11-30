// SmartContract(CC) of Safe Purging for Hamful Recycle & Recycling from Factory
// 공장 한 곳과 재활용 물질 처리 업체 사이에서 사용하는 체인 코드입니다.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct{}

type Recycle struct {
	ObjectType string `json:"docType"`
	CasNo      string `json:"casNo"`    // 재활용 물질 고유번호
	Phase      string `json:"phase"`    // 재활용 물질 상태 (고체, 액체, 기체)
	Name       string `json:"name"`     // 재활용 물질 이름
	Color      string `json:"color"`    // 재활용 물질 색상
	Quantity   string `json:"quantity"` // 공장이 보유하고 있는 재활용 물질의 양
}

type EmissionRecord struct {
	ObjectType string `json:"docType"`
	CasNo      string `json:"casNo"`    // 재활용 물질 고유번호
	Phase      string `json:"phase"`    // 재활용 물질 상태 (고체, 액체, 기체)
	Quantity   string `json:"quantity"` // 재활용 물질 처리량
	Cost       string `json:"cost"`     // 재활용 물질 처리 비용
}

func (t *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (t *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	if function == "initLedger" {
		return initLedger(APIstub)
	} else if function == "registerRecycle" {
		// 재활용 물질 등록
		return registerRecycle(APIstub, args)
	} else if function == "createRecycle" {
		// 재활용 물질 생성
		return createRecycle(APIstub, args)
	} else if function == "queryAllRecycles" {
		// 보유한 재활용 물질 조회
		return queryAllRecycles(APIstub)
	} else if function == "purgeRecycle" {
		// 재활용 물질 처리
		return purgeRecycle(APIstub, args)
	} else if function == "queryAllEmitionRecords" {
		// 재활용 물질 처리 내역 조회
		return queryAllEmitionRecords(APIstub, args)
	} else {
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	objectType := "recycle"

	recycles := []Recycle{
		{ObjectType: objectType, CasNo: "scrap001", Phase: "Solid", Name: "Carbon Steel(Metal)", Color: "None", Quantity: "1000"},
	}

	i := 0
	for i < len(recycles) {
		recycleAsBytes, _ := json.Marshal(recycles[i])
		APIstub.PutState(objectType+"-"+recycles[i].CasNo+"-"+recycles[i].Phase, recycleAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

// 공장에서 처리해야 하는 재활용 물질의 종류를 등록하는 함수
func registerRecycle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 0: CasNo, 1: Phase, 2: Name, 3: Color
	if len(args) != 4 {
		return shim.Error("Incorrect number of argument Expecting 4")
	}

	objectType := "recycle"
	recycle := Recycle{ObjectType: objectType, CasNo: args[0], Phase: args[1], Name: args[2], Color: args[3], Quantity: "0"}

	recycleAsBytes, _ := json.Marshal(recycle)
	APIstub.PutState(objectType+"-"+recycle.CasNo+"-"+recycle.Phase, recycleAsBytes)

	return shim.Success(nil)
}

// 공장을 가동하면서 쌓이는 재활용 물질을 기록하는 함수
func createRecycle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 0: CasNo, 1: Phase, 2: Quantity to add
	if len(args) != 3 {
		return shim.Error("Incorrect number of argument Expecting 3")
	}

	objectType := "recycle"

	casNo := args[0]
	phase := args[1]

	recycleAsBytes, err := APIstub.GetState(objectType + "-" + casNo + "-" + phase)
	if err != nil {
		return shim.Error(err.Error())
	} else if recycleAsBytes == nil {
		return shim.Error("일치하는 재활용 물질이 없습니다.")
	}

	recycle := Recycle{}
	json.Unmarshal(recycleAsBytes, &recycle)

	q1, _ := strconv.Atoi(recycle.Quantity)
	q2, _ := strconv.Atoi(args[2])

	quantity := q1 + q2
	quantityToString := strconv.Itoa(quantity)

	recycle.Quantity = quantityToString
	recycleAsBytes, _ = json.Marshal(recycle)

	APIstub.PutState(objectType+"-"+casNo+"-"+phase, recycleAsBytes)

	return shim.Success(nil)
}

// 공장이 보유한 모든 재활용 물질을 조회하는 함수
func queryAllRecycles(APIstub shim.ChaincodeStubInterface) sc.Response {

	queryString := fmt.Sprintf("{\"selector\": {\"docType\": \"recycle\"}}")
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllRecycles:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// 재활용 물질 처리 함수
func purgeRecycle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 0: CasNo, 1: Phase, 2: Quantity to purge, 3: Cost to purge
	if len(args) != 4 {
		return shim.Error("Incorrect number of argument Expecting 4.")
	}

	casNo := args[0]
	phase := args[1]

	// 공장이 보유한 해당 재활용 물질에서 처리한 양 만큼 제거
	recycleAsBytes, err := APIstub.GetState("recycle" + "-" + casNo + "-" + phase)
	if err != nil {
		return shim.Error(err.Error())
	} else if recycleAsBytes == nil {
		return shim.Error("일치하는 재활용 물질이 없습니다.")
	}

	recycle := Recycle{}
	json.Unmarshal(recycleAsBytes, &recycle)

	q1, _ := strconv.Atoi(recycle.Quantity)
	q2, _ := strconv.Atoi(args[2])

	quantity := q1 - q2
	quantityToString := strconv.Itoa(quantity)

	recycle.Quantity = quantityToString
	recycleAsBytes, _ = json.Marshal(recycle)

	APIstub.PutState("recycle"+"-"+casNo+"-"+phase, recycleAsBytes)

	// 재활용 물질 처리 내역에 기록
	emissionRecord := EmissionRecord{ObjectType: "emissionRecord", CasNo: casNo, Phase: phase, Quantity: args[2], Cost: args[3]}
	emissionRecordAsBytes, _ := json.Marshal(emissionRecord)

	unixTime := time.Now().Unix()
	unixTimeToString := strconv.FormatInt(unixTime, 10)

	APIstub.PutState("emissionRecord"+unixTimeToString, emissionRecordAsBytes)

	return shim.Success(nil)
}

// 재활용 물질 처리 내역을 조회하는 함수
func queryAllEmitionRecords(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	queryString := fmt.Sprintf("{\"selector\": {\"docType\": \"emissionRecord\"}}")
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllEmitionRecords:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// The main function is only relevant in unit test mode. Only included here for completenes
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
