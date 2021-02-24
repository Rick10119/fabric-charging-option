/*
 * The sample smart contract for documentation topic:
 * charging option
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the states of the CO_state
const ISSUE = 1     //
const BUY = 2       //co paid by the user and therefore able to list
const LIST = 3
const TIMELEN = 10  //number of time slots
const L_CS = 9      //number of charging stations
const R_ek = 6 // revenue of charging
const w = 4 // average unit cost of time
const mu = 1 // average charging rate

// Define the ChargingOption structure, with 6 properties.  Structure tags are used by encoding/json library
// There is another ID of the ChargingOption, as the key
type ChargingOption struct {
	ID_car   string `json:"ID_car"`
	ID_cs    string `json:"ID_cs"`
	T_arrive string `json:"T_arrive"`
	T_leave  string `json:"T_leave"`
	CO_price string `json:"CO_price"`
	CO_state string `json:"CO_state"`
}

type StationState struct {
	ID_cs    string `json:"ID_cs"`
	T        string `json:"Time"`
	N        string `json:"Number of charging slots"`
	Lambda_b string `json:"Lambda_b"`
	Lambda_k string `json:"Lambda_k"`
	CO_price string `json:"CO_price"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryCO" {
		return s.queryCO(APIstub, args)
	} else if function == "queryPrice" {
		return s.queryPrice(APIstub, args)
	} else if function == "queryList" {
		return s.queryList(APIstub, args)
	}else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "buy" {
		return s.buy(APIstub, args)
	} else if function == "list" {
		return s.list(APIstub, args)
	} else if function == "delist" {
		return s.delist(APIstub, args)
	} else if function == "confirm" {
		return s.confirm(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryCO(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	coAsBytes, _ := APIstub.GetState(args[0])

	return shim.Success(coAsBytes) //coAsBytes does not include the co_id
}

func (s *SmartContract) queryList(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {


	queryString := fmt.Sprintf("{\"selector\":{\"CO_state\":\"3\"}}")


	queryResults, err := getQueryResultForQueryString(APIstub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults) //coAsBytes does not include the co_id
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	var price_1b int
	var price_2b int
	var N int
	var Lambda_b int
	var Lambda_k int
	var stationstate StationState
	// J is written as 1/4 to guarantee the determinacy of the result

	for i := 0; i < L_CS; i++ {
		for j := 0; j < TIMELEN; j++ {

			N = 20
			Lambda_b = i/2
			Lambda_k = 15
			price_1b = R_ek * (L_CS - 1)/L_CS * Lambda_k / (mu*N - Lambda_b)
			price_2b = w / mu  * (Lambda_k * Lambda_k) / (mu*N - Lambda_b - Lambda_k)/(mu*N - Lambda_b - Lambda_k)/4/L_CS
			stationstate = StationState{
				ID_cs:    strconv.Itoa(i),
				T:        strconv.Itoa(j),
				N:        strconv.Itoa(N),
				Lambda_b: strconv.Itoa(Lambda_b),
				Lambda_k: strconv.Itoa(Lambda_k),
				CO_price: strconv.Itoa(price_1b + price_2b),
			}
			coAsBytes, _ := json.Marshal(stationstate)
			APIstub.PutState("Station"+strconv.Itoa(i)+"Slot"+strconv.Itoa(j), coAsBytes)
			fmt.Println("Added:", stationstate)
		}
	}

	return shim.Success(nil)
}

func (s *SmartContract) buy(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	// create a new charging option
	var chargingoption = ChargingOption{ID_car: args[1], ID_cs: args[2], T_arrive: args[3], T_leave: args[4]}

	chargingoption.CO_state = strconv.Itoa(ISSUE)
	// to calculate the price
	var price int
	price = 0
	t_arrive, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid time amount, expecting a integer value")
	}
	t_leave, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("Invalid time amount, expecting a integer value")
	}

	var price_1b int
	var price_2b int
	var N int
	var Lambda_b int
	var Lambda_k int
	for i := t_arrive; i < t_leave; i++ {
		// to get the state of the charging station and deserialize into stationstateT
		stationStateID := "Station" + args[2] + "Slot" + strconv.Itoa(i)
		ssAsBytes, _ := APIstub.GetState(stationStateID)
		stationstateT := StationState{}
		_ = json.Unmarshal(ssAsBytes, &stationstateT)

		// read the price and state now
		priceT, _ := strconv.Atoi(stationstateT.CO_price)
		Lambda_b, _ = strconv.Atoi(stationstateT.Lambda_b)
		Lambda_k, _ = strconv.Atoi(stationstateT.Lambda_k)
		N, _ = strconv.Atoi(stationstateT.N)
		// add the price of the current time slot T
		price = price + priceT

		// change the state(Lambda_b) of the particular charging station by +1
		Lambda_b = Lambda_b + 1
		// calculate the new price
		price_1b = R_ek * (L_CS - 1)/L_CS * Lambda_k / (mu*N - Lambda_b)
		price_2b = w / mu  * (Lambda_k * Lambda_k) / (mu*N - Lambda_b - Lambda_k)/(mu*N - Lambda_b - Lambda_k)/4/L_CS
		// update the state of the station at T
		stationstateT.Lambda_b = strconv.Itoa(Lambda_b)
		stationstateT.CO_price = strconv.Itoa(price_1b + price_2b)

		// serialize the new state and update it
		ssAsBytes, _ = json.Marshal(stationstateT)
		APIstub.PutState(stationStateID, ssAsBytes)
	}

	chargingoption.CO_price = strconv.Itoa(price)
	// the problem about lock???
	coAsBytes, _ := json.Marshal(chargingoption)
	APIstub.PutState(args[0], coAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) queryPrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	t_arrive, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Invalid time amount, expecting a integer value")
	}
	t_leave, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid time amount, expecting a integer value")
	}
	if t_arrive >= t_leave {
		return shim.Error("Invalid time amount, t_arrive should be smaller than t_leave")
	}

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer

	// for each station, calculate the price
	for j := 0; j < L_CS; j++ {
		var price = 0
		for i := t_arrive; i < t_leave; i++ {
			// to get the state of the charging station
			coAsBytes, _ := APIstub.GetState("Station" + strconv.Itoa(j) + "Slot" + strconv.Itoa(i))

			chargingoptionT := ChargingOption{}

			_ = json.Unmarshal(coAsBytes, &chargingoptionT)

			priceT, _ := strconv.Atoi(chargingoptionT.CO_price)

			price = price + priceT
		}

		buffer.WriteString("Station: ")
		buffer.WriteString(strconv.Itoa(j))

		buffer.WriteString(", Price: ")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(strconv.Itoa(price))
		buffer.WriteString("\n")

	}
	fmt.Printf("- queryPrice:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) list(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// fetch the co according to the ID_CO
	coAsBytes, _ := APIstub.GetState(args[0])
	chargingoption := ChargingOption{}
	_ = json.Unmarshal(coAsBytes, &chargingoption)

	// the user can list his/her co only if the co has been paid (and then set delivered by cs)
	if chargingoption.CO_state != strconv.Itoa(BUY) {
		return shim.Error("The charging option is not paid yet!")
	}
	chargingoption.CO_state = strconv.Itoa(LIST)
	chargingoption.CO_price = args[1]

	coAsBytes, _ = json.Marshal(chargingoption)
	APIstub.PutState(args[0], coAsBytes)


	return shim.Success(nil)
}

func (s *SmartContract) delist(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	coAsBytes, _ := APIstub.GetState(args[0])
	chargingoption := ChargingOption{}

	_ = json.Unmarshal(coAsBytes, &chargingoption)
	if chargingoption.CO_state != strconv.Itoa(LIST) {
		return shim.Error("The charging option is NOT listed yet!")
	}
	chargingoption.CO_state = strconv.Itoa(ISSUE)
	chargingoption.ID_car = args[1]

	coAsBytes, _ = json.Marshal(chargingoption)
	APIstub.PutState(args[0], coAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) confirm(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	coAsBytes, _ := APIstub.GetState(args[0])
	chargingoption := ChargingOption{}

	_ = json.Unmarshal(coAsBytes, &chargingoption)
	if chargingoption.CO_state == strconv.Itoa(BUY) {
		return shim.Error("Confirmation is not needed!")
	}
	chargingoption.CO_state = strconv.Itoa(BUY)

	coAsBytes, _ = json.Marshal(chargingoption)
	APIstub.PutState(args[0], coAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		buffer.WriteString("\"Listed Charging Option\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
	}

	return &buffer, nil
}
