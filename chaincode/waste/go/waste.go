// SmartContract(CC) of Safe Purging for Hamful Waste & Recycling from Factory
// 공장 한 곳과 폐기물 처리 업체 사이에서 사용하는 체인 코드입니다.
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

type Waste struct {
	ObjectType string `json:"docType"`
	CasNo      string `json:"casNo"`    // 폐기물 고유번호
	Phase      string `json:"phase"`    // 폐기물 상태 (고체, 액체, 기체)
	Name       string `json:"name"`     // 폐기물 이름
	Color      string `json:"color"`    // 폐기물 색상
	Quantity   string `json:"quantity"` // 공장이 보유하고 있는 폐기물의 양
}

type EmissionRecord struct {
	ObjectType string `json:"docType"`
	CasNo      string `json:"casNo"`    // 폐기물 고유번호
	Phase      string `json:"phase"`    // 폐기물 상태 (고체, 액체, 기체)
	Quantity   string `json:"quantity"` // 폐기물 처리량
	Cost       string `json:"cost"`     // 폐기물 처리 비용
}

func (t *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (t *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	if function == "initLedger" {
		return initLedger(APIstub)
	} else if function == "registerWaste" {
		// 폐기물 등록
		return registerWaste(APIstub, args)
	} else if function == "createWaste" {
		// 폐기물 생성
		return createWaste(APIstub, args)
	} else if function == "queryAllWastes" {
		// 보유한 폐기물 조회
		return queryAllWastes(APIstub)
	} else if function == "purgeWaste" {
		// 폐기물 처리
		return purgeWaste(APIstub, args)
	} else if function == "queryAllEmitionRecords" {
		// 폐기물 처리 내역 조회
		return queryAllEmitionRecords(APIstub, args)
	} else {
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	objectType := "waste"

	wastes := []Waste{
		{ObjectType: objectType, CasNo: "108-39-4", Phase: "Liquid", Name: "m-Crezol(Phenol)", Color: "yellow", Quantity: "1000"},
		{ObjectType: objectType, CasNo: "506-77-4", Phase: "Oil", Name: "Cyanogen chloride", Color: "None", Quantity: "1000"},
	}

	i := 0
	for i < len(wastes) {
		wasteAsBytes, _ := json.Marshal(wastes[i])
		APIstub.PutState(objectType+"-"+wastes[i].CasNo+"-"+wastes[i].Phase, wasteAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

// 공장에서 처리해야 하는 폐기물의 종류를 등록하는 함수
func registerWaste(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 0: CasNo, 1: Phase, 2: Name, 3: Color
	if len(args) != 4 {
		return shim.Error("Incorrect number of argument Expecting 4")
	}

	objectType := "waste"
	waste := Waste{ObjectType: objectType, CasNo: args[0], Phase: args[1], Name: args[2], Color: args[3], Quantity: "0"}

	wasteAsBytes, _ := json.Marshal(waste)
	APIstub.PutState(objectType+"-"+waste.CasNo+"-"+waste.Phase, wasteAsBytes)

	return shim.Success(nil)
}

// 공장을 가동하면서 쌓이는 폐기물을 기록하는 함수
func createWaste(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 0: CasNo, 1: Phase, 2: Quantity to add
	if len(args) != 3 {
		return shim.Error("Incorrect number of argument Expecting 3")
	}

	objectType := "waste"

	casNo := args[0]
	phase := args[1]

	wasteAsBytes, err := APIstub.GetState(objectType + "-" + casNo + "-" + phase)
	if err != nil {
		return shim.Error(err.Error())
	} else if wasteAsBytes == nil {
		return shim.Error("일치하는 폐기물이 없습니다.")
	}

	waste := Waste{}
	json.Unmarshal(wasteAsBytes, &waste)

	q1, _ := strconv.Atoi(waste.Quantity)
	q2, _ := strconv.Atoi(args[2])

	quantity := q1 + q2
	quantityToString := strconv.Itoa(quantity)

	waste.Quantity = quantityToString
	wasteAsBytes, _ = json.Marshal(waste)

	APIstub.PutState(objectType+"-"+casNo+"-"+phase, wasteAsBytes)

	return shim.Success(nil)
}

// 공장이 보유한 모든 폐기물을 조회하는 함수
func queryAllWastes(APIstub shim.ChaincodeStubInterface) sc.Response {

	queryString := fmt.Sprintf("{\"selector\": {\"docType\": \"waste\"}}")
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

	fmt.Printf("- queryAllWastes:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// 폐기물 처리 함수
func purgeWaste(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 0: CasNo, 1: Phase, 2: Quantity to purge, 3: Cost to purge
	if len(args) != 4 {
		return shim.Error("Incorrect number of argument Expecting 4.")
	}

	casNo := args[0]
	phase := args[1]

	// 공장이 보유한 해당 폐기물에서 처리한 양 만큼 제거
	wasteAsBytes, err := APIstub.GetState("waste" + "-" + casNo + "-" + phase)
	if err != nil {
		return shim.Error(err.Error())
	} else if wasteAsBytes == nil {
		return shim.Error("일치하는 폐기물이 없습니다.")
	}

	waste := Waste{}
	json.Unmarshal(wasteAsBytes, &waste)

	q1, _ := strconv.Atoi(waste.Quantity)
	q2, _ := strconv.Atoi(args[2])

	quantity := q1 - q2
	quantityToString := strconv.Itoa(quantity)

	waste.Quantity = quantityToString
	wasteAsBytes, _ = json.Marshal(waste)

	APIstub.PutState("waste"+"-"+casNo+"-"+phase, wasteAsBytes)

	// 폐기물 처리 내역에 기록
	emissionRecord := EmissionRecord{ObjectType: "emissionRecord", CasNo: casNo, Phase: phase, Quantity: args[2], Cost: args[3]}
	emissionRecordAsBytes, _ := json.Marshal(emissionRecord)

	unixTime := time.Now().Unix()
	unixTimeToString := strconv.FormatInt(unixTime, 10)

	APIstub.PutState("emissionRecord"+unixTimeToString, emissionRecordAsBytes)

	return shim.Success(nil)
}

// 폐기물 처리 내역을 조회하는 함수
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
