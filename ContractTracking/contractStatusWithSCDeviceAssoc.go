package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type ContractsTotal struct {
	TotalRecords int `json:"totalRecords"`
	Contracts    []struct {
		ContractNumber       string `json:"contractNumber"`
		ContractStatus       string `json:"contractStatus"`
		ContractBillToID     string `json:"contractBillToID"`
		ContractBillToName   string `json:"contractBillToName"`
		ContractBillToGUID   string `json:"contractBillToGUID"`
		ContractBillToGUName string `json:"contractBillToGUName"`
		EndCustomers         []struct {
			Country string `json:"country"`
			ID      string `json:"id"`
			Name    string `json:"name"`
		} `json:"endCustomers"`
		ServiceLevel       []string `json:"serviceLevel"`
		ServiceSKU         []string `json:"serviceSKU"`
		ServiceDescription []string `json:"serviceDescription"`
		ContractEndDate    string   `json:"contractEndDate"`
		EarliestEndDate    string   `json:"earliestEndDate,omitempty"`
		ListPrice          float64  `json:"listPrice"`
		ContractLabel      string   `json:"contractLabel"`
		Currency           string   `json:"currency"`
	} `json:"contracts"`
}

type IndividualContracts struct {
	CustomerName            string
	CustomerID              string
	CustomerCountry         string
	ContractNumber          string
	ContractStatus          string
	ContractBillToID        string
	ContractBillToName      string
	ContractBillToGUID      string
	ContractBillToGUName    string
	ContractEarliestEndDate string
	ContractListPrice       string
	ContractLabel           string
	ContractEndDate         string
	EarliestEndDate         string
	ListPrice               float64
	Currency                string
	ServiceSKU              string
	ServiceLevel            string
	ServiceDescription      string
}

type ContractDetails struct {
	TotalRecords int `json:"totalRecords"`
	Instances    []struct {
		LastDateOfSupport    time.Time `json:"lastDateOfSupport,omitempty"`
		SerialNumber         string    `json:"serialNumber,omitempty"`
		ParentSerialNumber   string    `json:"parentSerialNumber,omitempty"`
		Minor                bool      `json:"minor"`
		InstanceNumber       string    `json:"instanceNumber"`
		ParentInstanceNumber string    `json:"parentInstanceNumber"`
		InstalledBaseStatus  string    `json:"installedBaseStatus"`
		EndCustomer          struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Address struct {
				Address1   string `json:"address1"`
				City       string `json:"city"`
				Country    string `json:"country"`
				State      string `json:"state"`
				PostalCode string `json:"postalCode"`
			} `json:"address"`
		} `json:"endCustomer"`
		ServiceSKU         string    `json:"serviceSKU"`
		ServiceLevel       string    `json:"serviceLevel"`
		ServiceDescription string    `json:"serviceDescription"`
		StartDate          time.Time `json:"startDate"`
		EndDate            time.Time `json:"endDate"`
		Contract           struct {
			Number               string `json:"number"`
			LineStatus           string `json:"lineStatus"`
			BillToGlobalUltimate struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"billToGlobalUltimate"`
			BillTo struct {
				Address1   string `json:"address1"`
				City       string `json:"city"`
				Country    string `json:"country"`
				State      string `json:"state"`
				PostalCode string `json:"postalCode"`
				Location   string `json:"location"`
				Name       string `json:"name"`
			} `json:"billTo"`
		} `json:"contract"`
		MacID                          string `json:"macId"`
		Quantity                       int    `json:"quantity"`
		SalesOrderNumber               string `json:"salesOrderNumber"`
		PurchaseOrderNumber            string `json:"purchaseOrderNumber"`
		MaintenanceSalesOrderNumber    string `json:"maintenanceSalesOrderNumber"`
		MaintenancePurchaseOrderNumber string `json:"maintenancePurchaseOrderNumber"`
		Product                        struct {
			Number      string `json:"number"`
			Description string `json:"description"`
			Family      string `json:"family"`
			Group       string `json:"group"`
			SubType     string `json:"subType"`
			BillTo      struct {
				Address1   string `json:"address1"`
				Address2   string `json:"address2"`
				City       string `json:"city"`
				Country    string `json:"country"`
				State      string `json:"state"`
				PostalCode string `json:"postalCode"`
				Location   string `json:"location"`
				Name       string `json:"name"`
			} `json:"billTo"`
			ShipTo struct {
				Address1   string `json:"address1"`
				Address2   string `json:"address2"`
				City       string `json:"city"`
				Country    string `json:"country"`
				State      string `json:"state"`
				PostalCode string `json:"postalCode"`
				Location   string `json:"location"`
				Name       string `json:"name"`
			} `json:"shipTo"`
		} `json:"product"`
		ItemType string `json:"itemType"`
		Warranty struct {
			Type    string    `json:"type"`
			Status  string    `json:"status"`
			EndDate time.Time `json:"endDate"`
		} `json:"warranty"`
		ShipDate              time.Time `json:"shipDate"`
		EndUserGlobalUltimate struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"endUserGlobalUltimate"`
		ResellerGlobalUltimate struct {
		} `json:"resellerGlobalUltimate,omitempty"`
		Distributor struct {
			BillToID      string `json:"billToId"`
			BillToName    string `json:"billToName"`
			BillToAddress struct {
				Address1   string `json:"address1"`
				City       string `json:"city"`
				Country    string `json:"country"`
				State      string `json:"state"`
				PostalCode string `json:"postalCode"`
			} `json:"billToAddress"`
			BillToGlobalUltimate struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"billToGlobalUltimate"`
		} `json:"distributor"`
		CartonID                string `json:"cartonId,omitempty"`
		ReplacementSerialNumber string `json:"replacementSerialNumber,omitempty"`
	} `json:"instances"`
}

type ContractDetailElements struct {
	ContractNumber          string
	ContractLineStatus      string
	SerialNumber            string
	ParentSerialNumber      string
	Minor                   bool
	InstanceNumber          string
	ParentInstanceNumber    string
	InstalledBaseStatus     string
	EndCustomerID           string
	EndCustomerName         string
	ProductNumber           string
	ProductDescription      string
	ProductFamily           string
	ProductGroup            string
	ProductSubtype          string
	UUID                    string
	DBInsertionTimeStamp    string
	CompanyID               string
	DisplayName             string
	Description             string
	ContractStatus          string
	ContractEndDate         string
	ContractEarliestEndDate string
	ContractListPrice       float64
	ContractLabel           string
	ConnectWiseCompanyID    string
	ServiceSKU              string
	ServiceLevel            string
	ServiceDescription      string
	DeviceLevelStartDate    string
	DeviceLevelEndDate      string
}

type Device struct {
	UUID                   string
	Timestamp              int
	SystemModel            string
	EndpointModel          string
	SystemVersion          string
	FirmwareVersion        string
	Manufacturer           string
	DisplayName            string
	Description            string
	Categories             string
	SWproductID            string
	CompanyID              string
	BillingName            string
	ServiceBoard           string
	TicketType             string
	AgreementID            string
	SnmpVersion            string
	SubType                string
	NetworkAddress         string
	NetworkMask            string
	MacAddress             string
	ComponentSerialNumber  string
	ParentSerialNumber     string
	IsParentDevice         string
	EntPhysicalDescription string
	Memory                 string
	Vendor                 string
	HostSerialNumber       string
	Systeminfo             string
}

type Company struct {
	ConnectWiseCompanyID string
	CustomerName         string
	ContractNumber       string
}

type AccessToken struct {
	Access_token string
	Token_type   string
	Expires_in   int
}

///Variables//
var authToken = "null"
var ContractsTotal1 = ContractsTotal{}
var IndividualContracts1 = IndividualContracts{}
var ContactDetails1 = ContractDetails{}
var ContractDetailElements1 = ContractDetailElements{}
var Device1 = Device{}
var Company1 = Company{}

/////////////

// Function to return all contracts currently associated with SMP's bill to ID of 1397311
// This API requires that the Bill To number is sent both as Request-Id in the header and in the billToLocation in the body
// The API requires either a serial number or a bill to ID in the body and with the goal of retrieving all contracts
// the bill to ID is used in the body.
func CCWContractStatusDump() {
	url := "https://api.cisco.com/ccw/renewals/api/v1.0/search/contractSummary"
	method := "POST"

	apiAuthToken()

	payload := strings.NewReader("{\n  \"billToLocation\": [\n    1397311\n  ],	\n  \"offset\": 0,\n  \"limit\": 1000,\n  \"configurations\": false\n}")
	// For loop is necessary a maximum of 1000 records will be returned via the API on a single request.  The offset is
	// used as a type of pagination and by setting it to 1000 on the second loop the records starting at 1000 will be
	// returned (1000 record number not page).  At the time of this development - 1025 records exists and thus a loop thru
	// on a second iteration - changing only the offset in the POST payload - is sufficient.
	for x := 1; x <= 2; x++ {
		if x == 1 {
			// No action necessary
		} else if x == 2 {
			payload = strings.NewReader("{\n  \"billToLocation\": [\n    1397311\n  ],	\n  \"offset\": 1000,\n  \"limit\": 1000,\n  \"configurations\": false\n}")
		}

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Request-Id", "1397311")

		req.Header.Add("Authorization", "Bearer "+authToken)

		// req.Header.Add("Authorization", "Bearer 2wjiceJJpU3c41spLlJ6CLXg8EyU")
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)

		if err != nil {
			fmt.Println("within res reattempt conditional")
			time.Sleep(1 * time.Second)
			res, err = client.Do(req)
			if err != nil {
				fmt.Println("within res reattempt2 conditional")
				time.Sleep(1 * time.Second)
				res, err = client.Do(req)
			}
			if err != nil {
				fmt.Println("within res reattempt3 conditional")
				time.Sleep(1 * time.Second)
				res, err = client.Do(req)
			}
			if err != nil {
				fmt.Println("within res reattempt4 conditional")
				time.Sleep(1 * time.Second)
				res, err = client.Do(req)
			}
			if err != nil {
				fmt.Println(err)
			}
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		// check for response error
		if err != nil {
			fmt.Println("Error:", err)
		}

		ContractsTotal1 = ContractsTotal{}
		json.Unmarshal(body, &ContractsTotal1)

		//fmt.Println("Length of contracts: ", len(ContractsTotal1.Contracts))
		for _, v := range ContractsTotal1.Contracts {
			IndividualContracts1 = IndividualContracts{}
			// BELIEVE THIS ONLY OCCURS FOR SMP CONTRACTS - NEED TO CONFIRM AND DEAL WITH THESE INSTANCES - TWO RECORDS MATCH
			if len(v.EndCustomers) == 0 {
				continue
			}
			IndividualContracts1.CustomerName = v.EndCustomers[0].Name
			IndividualContracts1.CustomerID = v.EndCustomers[0].ID
			IndividualContracts1.CustomerCountry = v.EndCustomers[0].Country
			IndividualContracts1.ContractNumber = v.ContractNumber
			IndividualContracts1.ContractStatus = v.ContractStatus
			IndividualContracts1.ContractBillToID = v.ContractBillToID
			IndividualContracts1.ContractBillToName = v.ContractBillToName
			IndividualContracts1.ContractBillToGUID = v.ContractBillToGUID
			IndividualContracts1.ContractBillToGUName = v.ContractBillToGUName
			IndividualContracts1.ContractEndDate = v.ContractEndDate
			IndividualContracts1.EarliestEndDate = v.EarliestEndDate
			IndividualContracts1.ListPrice = v.ListPrice
			IndividualContracts1.ContractLabel = v.ContractLabel
			IndividualContracts1.Currency = v.Currency
			IndividualContracts1.ContractStatus = v.ContractStatus

			CCWPerContractDetails(IndividualContracts1.ContractNumber)
			time.Sleep(1 * time.Second)
		}
	}
}

// Function takes individual contract numbers passed in from the CCWContractStatusDump function and conducts a query to
// obtain all associated devices, contract expiration dates, and other details available in the search lines API
func CCWPerContractDetails(ContractNumber string) {
	apiUrl := "https://api.cisco.com/ccw/renewals/api/v1.0/search/lines"

	apiAuthToken()

	stringEx := "{\"contractNumbers\":[\"" + ContractNumber + "\"],}"
	//fmt.Println(stringEx)

	reqBody := strings.NewReader(stringEx)

	// create a request object
	req, _ := http.NewRequest(
		"POST",
		apiUrl,
		reqBody,
	)

	// add a request header
	req.Header.Add("Request-Id", "1397311")
	req.Header.Add("Authorization", "Bearer "+authToken)
	req.Header.Add("Content-Type", "application/json")

	// send an HTTP using `req` object
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("within res reattempt conditional")
		time.Sleep(1 * time.Second)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("within res reattempt2 conditional")
			time.Sleep(1 * time.Second)
			res, err = http.DefaultClient.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt3 conditional")
			time.Sleep(1 * time.Second)
			res, err = http.DefaultClient.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt4 conditional")
			time.Sleep(1 * time.Second)
			res, err = http.DefaultClient.Do(req)
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	// close response body
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	// check for response error
	if err != nil {
		fmt.Println("Error:", err)
	}

	ContactDetails1 = ContractDetails{}
	json.Unmarshal(body, &ContactDetails1)

	fmt.Println(ContactDetails1)

	//fmt.Println("Total records in contract: ", ContactDetails1.TotalRecords)
	ContractDetailElements1 = ContractDetailElements{}

	if ContactDetails1.TotalRecords == 0 {
		ContractDetailElements1.ContractNumber = ContractNumber
		ContractDetailElements1.SerialNumber = "null"
		ContractDetailElements1.ContractStatus = IndividualContracts1.ContractStatus
		ContractDetailElements1.ContractEndDate = IndividualContracts1.ContractEndDate
		ContractDetailElements1.ContractEarliestEndDate = IndividualContracts1.EarliestEndDate
		ContractDetailElements1.ContractListPrice = IndividualContracts1.ListPrice
		ContractDetailElements1.ContractLabel = IndividualContracts1.ContractLabel
		ContractDetailElements1.EndCustomerName = IndividualContracts1.CustomerName
		ContractDetailElements1.EndCustomerID = IndividualContracts1.CustomerID

		uuid, err := exec.Command("uuidgen").Output()
		if err != nil {
			log.Fatal(err)
		}

		ContractDetailElements1.UUID = string(uuid)

		ContractDetailElements1.DBInsertionTimeStamp = time.Now().Format(time.RFC3339)

		CompanyIDToContractAssocQuery(ContractDetailElements1)

	}

	for _, v := range ContactDetails1.Instances {
		ContractDetailElements1 = ContractDetailElements{}

		ContractDetailElements1.ContractNumber = v.Contract.Number
		ContractDetailElements1.ContractLineStatus = v.Contract.LineStatus
		// ContractDetailElements1.LastDateOfSupport = v.LastDateOfSupport.Format(time.RFC3339)
		ContractDetailElements1.SerialNumber = v.SerialNumber
		ContractDetailElements1.ParentSerialNumber = v.ParentSerialNumber
		ContractDetailElements1.Minor = v.Minor
		ContractDetailElements1.InstanceNumber = v.InstanceNumber
		ContractDetailElements1.ParentInstanceNumber = v.ParentInstanceNumber
		ContractDetailElements1.InstalledBaseStatus = v.InstalledBaseStatus
		ContractDetailElements1.EndCustomerID = v.EndCustomer.ID
		ContractDetailElements1.EndCustomerName = v.EndCustomer.Name
		ContractDetailElements1.ProductNumber = v.Product.Number
		ContractDetailElements1.ProductDescription = v.Product.Description
		ContractDetailElements1.ProductFamily = v.Product.Family
		ContractDetailElements1.ProductGroup = v.Product.SubType
		ContractDetailElements1.ServiceSKU = v.ServiceSKU
		ContractDetailElements1.ServiceLevel = v.ServiceLevel
		ContractDetailElements1.ServiceDescription = v.ServiceDescription
		ContractDetailElements1.ContractStatus = IndividualContracts1.ContractStatus
		ContractDetailElements1.ContractEndDate = IndividualContracts1.ContractEndDate
		ContractDetailElements1.ContractEarliestEndDate = IndividualContracts1.EarliestEndDate
		ContractDetailElements1.ContractListPrice = IndividualContracts1.ListPrice
		ContractDetailElements1.ContractLabel = IndividualContracts1.ContractLabel
		ContractDetailElements1.DeviceLevelStartDate = v.StartDate.Format(time.RFC3339)
		ContractDetailElements1.DeviceLevelEndDate = v.EndDate.Format(time.RFC3339)

		if ContractDetailElements1.SerialNumber == "FCH1938V27W" {
			fmt.Println("#######SN FCH1938V27W found - here is SerialNumber field: ", ContractDetailElements1.SerialNumber)
		}

		uuid, err := exec.Command("uuidgen").Output()
		if err != nil {
			fmt.Println(err)
		}

		ContractDetailElements1.UUID = string(uuid)

		ContractDetailElements1.DBInsertionTimeStamp = time.Now().Format(time.RFC3339)

		if len(ContractDetailElements1.SerialNumber) == 0 {
			ContractDetailElements1.SerialNumber = "null"
		}

		fmt.Println("Customer Name: ", ContractDetailElements1.EndCustomerName)
		fmt.Println("Contract Number: ", ContractDetailElements1.ContractNumber)

		CompanyIDToContractAssocQuery(ContractDetailElements1)
		time.Sleep(1 * time.Second)
	}
}

func CompanyIDToContractAssocQuery(ContractDetailElements1 ContractDetailElements) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{}
	// Query of DyanmoDB CompanyID to ContractID assoociation table for purpose of LMDevice lookup
	// if ContractDetailElements1.SerialNumber != "null" {
	queryInput = &dynamodb.QueryInput{
		TableName: aws.String("smartcentre_ccw_company_associations"),
		KeyConditions: map[string]*dynamodb.Condition{
			"ContractNumber": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(ContractDetailElements1.ContractNumber),
					},
				},
			},
		},
	}
	var result, err2 = svc.Query(queryInput)

	if err2 != nil {
		fmt.Println("Query of martcentre_ccw_company_associations  failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}

	// If no intel is returned from the LMDevice table - indicating it is not monitored by LM and is not a SMARTcentre
	// managed device - proceed to DynamoDB insert of CCW contract details
	if len(result.Items) == 0 {
		ContractDetailElements1.CompanyID = "NoCurrentAssociation"
		apiUpdateDynamoDB(ContractDetailElements1)
	}

	// Iterate thru DynamoDB Returned Rows
	for _, v := range result.Items {
		Company1 = Company{}
		err = dynamodbattribute.UnmarshalMap(v, &Company1)
		if err != nil {
			fmt.Println("Unmarshal to Device1 failed:")
			fmt.Println((err.Error()))
			os.Exit(1)
		}

		fmt.Println("#########Company1###########")
		fmt.Println(Company1)
		fmt.Println("##########################")

		if len(Company1.ConnectWiseCompanyID) > 0 {
			ContractDetailElements1.CompanyID = Company1.ConnectWiseCompanyID
		} else {
			ContractDetailElements1.CompanyID = "NoCurrentAssociation"
			apiUpdateDynamoDB(ContractDetailElements1)
		}
		if ContractDetailElements1.SerialNumber == "null" {
			apiUpdateDynamoDB(ContractDetailElements1)
		} else {
			LogicMonitorDeviceQuery(ContractDetailElements1)
		}
	}
	// } else {
	// 	// If device has no serial number - proceed to DynamoDB record insertion of CCW contract intel
	// 	ContractDetailElements1.CompanyID = "NoCurrentAssociation"
	// 	apiUpdateDynamoDB(ContractDetailElements1)
	// }
	fmt.Println("#############ContractDetailElements############")
	fmt.Println(ContractDetailElements1)
	fmt.Println(ContractDetailElements1.EndCustomerName)
	fmt.Println("###################################")

}

func LogicMonitorDeviceQuery(ContractDetailElements1 ContractDetailElements) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{}
	// Query of SMARTcentre managed devices in DynamoDB LogicMonitor table for displayname and any other correlating
	// intel of interest
	// Intel will only be available for devices that have a serial number and are monitored by LogicMonitor
	if ContractDetailElements1.SerialNumber != "null" {
		fmt.Println("###Within LM Device query conditional#######")
		fmt.Println("CompanyID: ", ContractDetailElements1.CompanyID)
		fmt.Println("ComponentSerialNumber: ", ContractDetailElements1.SerialNumber)
		queryInput = &dynamodb.QueryInput{
			TableName: aws.String("LMDevice"),
			IndexName: aws.String("CompanyID-ComponentSerialNumber-index"),
			KeyConditions: map[string]*dynamodb.Condition{
				"CompanyID": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(ContractDetailElements1.CompanyID),
						},
					},
				},
				"ComponentSerialNumber": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(ContractDetailElements1.SerialNumber),
						},
					},
				},
			},
		}
		var result, err2 = svc.Query(queryInput)

		if err2 != nil {
			fmt.Println("Query of LMDevice  failed:")
			fmt.Println((err.Error()))
			os.Exit(1)
		}

		// If no intel is returned from the LMDevice table - indicating it is not monitored by LM and is not a SMARTcentre
		// managed device - proceed to DynamoDB insert of CCW contract details
		if len(result.Items) == 0 {
			apiUpdateDynamoDB(ContractDetailElements1)
		}

		// Iterate thru DynamoDB Returned Rows
		for _, v := range result.Items {
			Device1 = Device{}
			err = dynamodbattribute.UnmarshalMap(v, &Device1)
			if err != nil {
				fmt.Println("Unmarshal to Device1 failed:")
				fmt.Println((err.Error()))
				os.Exit(1)
			}

			ContractDetailElements1.DisplayName = Device1.DisplayName
			ContractDetailElements1.Description = Device1.Description
			ContractDetailElements1.ConnectWiseCompanyID = Device1.CompanyID
			apiUpdateDynamoDB(ContractDetailElements1)
		}
	} else {
		// If device has no serial number - proceed to DynamoDB record insertion of CCW contract intel
		apiUpdateDynamoDB(ContractDetailElements1)
	}
}

func apiUpdateDynamoDB(ContractDetailElements1 ContractDetailElements) {

	// If displayname is not populated the DB insert will fail.  Should revisit this logic but temporarily setting
	// displayname to null in such cases.
	if len(ContractDetailElements1.DisplayName) == 0 {
		ContractDetailElements1.DisplayName = "null"
	}

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Update item in table smartcentre_ccw_contract_details
	tableName := "smartcentre_ccw_contract_details"

	av, err := dynamodbattribute.MarshalMap(ContractDetailElements1)
	if err != nil {
		fmt.Println("Got error marshalling item:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully updated contract element detail")

}

func apiAuthToken() {

	// API key removed at end of string for GitHub upload
	url := "https://cloudsso.cisco.com/as/token.oauth2?grant_type=client_credentials&client_id=bghqpdv3tqg66eedcuqd3hqr&client_secret"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	// The auth API experiences timeouts occasionally in the network request.  To ensure the flow continues the
	// logc below will re-attempt the API request four times before eventually erroring out.  A two second delay is
	// added between each of the two additional requests in an attempt to allow network timeouts to subside.
	if err != nil {
		fmt.Println("within res reattempt conditional")
		time.Sleep(1 * time.Second)
		res, err = client.Do(req)
		if err != nil {
			fmt.Println("within res reattempt2 conditional")
			time.Sleep(1 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt3 conditional")
			time.Sleep(1 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt4 conditional")
			time.Sleep(1 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	AccessToken1 := AccessToken{}

	json.Unmarshal(body, &AccessToken1)

	authToken = AccessToken1.Access_token
}

func main() {
	CCWContractStatusDump()
}
