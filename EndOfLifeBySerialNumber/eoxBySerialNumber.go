package main

import (
	"encoding/json"
	"encoding/xml"
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
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

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
	EntPhysicalDescription string
	Memory                 string
	Vendor                 string
	HostSerialNumber       string
	DBInsertionTimeStamp   string
	// Meraki Specific Attributed block
	APIName   string
	Notes     string
	Firmware  string
	Lng       string
	Lat       string
	LanIP     string
	Wan1IP    string
	Wan2IP    string
	NetworkID string
	Model     string
	///////////////////////////////////
	EndOfServiceContractRenewal string `xml:"EOXRecord>EndOfServiceContractRenewal"`
	EOLProductID                string `xml:"EOXRecord>EOLProductID"`
	EOXInputValue               string `xml:"EOXRecord>EOXInputValue"`
	LinkToProductBulletinURL    string `xml:"EOXRecord>LinkToProductBulletinURL"`
	EndOfSaleDate               string `xml:"EOXRecord>EndOfSaleDate"`
	EndOfSWMaintenanceReleases  string `xml:"EOXRecord>EndOfSWMaintenanceReleases"`
	LastDateOfSupport           string `xml:"EOXRecord>LastDateOfSupport"`
	MigrationProductId          string `xml:"EOXRecord>EOXMigrationDetails>MigrationProductId"`
	MigrationInformation        string `xml:"EOXRecord>EOXMigrationDetails>MigrationInformation"`
	MigrationProductInfoURL     string `xml:"EOXRecord>EOXMigrationDetails>MigrationProductInfoURL"`
	ErrorDataValue              string `xml:"EOXRecord>EOXError>ErrorDataValue"`
	ErrorDescription            string `xml:"EOXRecord>EOXError>ErrorDescription"`
}

type Company struct {
	CompanyID   string
	BillingName string
}

type AccessToken struct {
	Access_token string
	Token_type   string
	Expires_in   int
}

var authToken = "null"

///Variables//
var Company1 = Company{}
var Device1 = Device{}
var apiIterator int

/////////////

func RetrieveCurrentCustomers() {

	// Retrieve current managed services customer list from DynamoDB
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	proj := expression.NamesList(expression.Name("CompanyID"), expression.Name("BillingName"))

	expr, err := expression.NewBuilder().WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String("LMCompanies"),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}

	// Iterate thru DynamoDB Returned Rows
	for _, v := range result.Items {
		Company1 = Company{}
		err = dynamodbattribute.UnmarshalMap(v, &Company1)

		//Refresh Cisco API Token at beginning of each customer query loop
		apiAuthToken()

		// Incrementing the apiIerator variable - which is used with the API token function - will ensure
		// that a different API account/token is used on every other customer to attempt to ensure no one
		// API account hits request limit
		apiIterator++
		PerCustomerDeviceDBRetrieval(Company1.CompanyID)
	}
}

func PerCustomerDeviceDBRetrieval(CompanyID string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String("LMDevice"),
		KeyConditions: map[string]*dynamodb.Condition{
			"CompanyID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(CompanyID),
					},
				},
			},
		},
	}

	var result, _ = svc.Query(queryInput)

	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}

	// Iterate thru DynamoDB Returned Rows
	for _, v := range result.Items {
		Device1 = Device{}
		err = dynamodbattribute.UnmarshalMap(v, &Device1)
		if strings.Contains(strings.ToLower(Device1.Manufacturer), "cisco") {
			apiSN()
		} else if len(Device1.Manufacturer) > 0 && !strings.Contains(strings.ToLower(Device1.Manufacturer), "cisco") {
			continue
		} else if strings.Contains(Device1.ComponentSerialNumber, "NA_") {
			continue
		} else {
			apiSN()
		}
	}

}

func apiSN() {

	fmt.Println("Entering main component API function")

	//If the current instance is a sub-component - update the serial number used in the API request to the subcomponent SN
	apiRequestSN := "null"
	if len(Device1.ComponentSerialNumber) > 0 {
		apiRequestSN = Device1.ComponentSerialNumber
	} else if len(Device1.ParentSerialNumber) > 0 {
		apiRequestSN = Device1.ParentSerialNumber
	} else {
		return
	}

	url := "https://api.cisco.com/supporttools/eox/rest/5/EOXBySerialNumber/1/" + apiRequestSN + "?responseencoding=xml"

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Bearer "+authToken)

	res, err := client.Do(req)

	// The auth API experiences timeouts occasionally in the network request.  To ensure the flow continues the
	// logc below will re-attempt the API request four times before eventually erroring out.  A two second delay is
	// added between each of the two additional requests in an attempt to allow network timeouts to subside.
	if err != nil {
		fmt.Println("within res reattempt conditional")
		time.Sleep(2 * time.Second)
		res, err = client.Do(req)
		if err != nil {
			fmt.Println("within res reattempt2 conditional")
			time.Sleep(2 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt3 conditional")
			time.Sleep(2 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt4 conditional")
			time.Sleep(2 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))

	xml.Unmarshal(body, &Device1)

	if Device1.EndOfServiceContractRenewal == "" {
		Device1.EndOfServiceContractRenewal = "NA"
	}
	if Device1.LinkToProductBulletinURL == "" {
		Device1.LinkToProductBulletinURL = "NA"
	}
	if Device1.EndOfSaleDate == "" {
		Device1.EndOfSaleDate = "NA"
	}
	if Device1.EndOfSWMaintenanceReleases == "" {
		Device1.EndOfSWMaintenanceReleases = "NA"
	}
	if Device1.LastDateOfSupport == "" {
		Device1.LastDateOfSupport = "NA"
	}

	if Device1.MigrationInformation == "" {
		Device1.MigrationInformation = "NA"
	}
	if Device1.MigrationProductInfoURL == "" {
		Device1.MigrationProductInfoURL = "NA"
	}
	if Device1.EOLProductID != "" {
		Device1.EndpointModel = Device1.EOLProductID
	} else if Device1.ErrorDataValue != "" {
		Device1.EndpointModel = Device1.ErrorDataValue
		Device1.EOLProductID = "NA"
	} else {
		Device1.MigrationProductId = "NA"
	}

	fmt.Println(Device1)

	apiUpdateDynamoDB()

}

func apiUpdateDynamoDB() {

	fmt.Println("Entering DynamoDB update section")
	fmt.Println("device 1 as it enters update DB")
	fmt.Println(Device1)
	fmt.Println("######################")

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	Device1.UUID = string(uuid)
	// Device1.Parent_serial_number = Device1.Serial_number

	timeStamp := time.Now()
	timeFormatted := timeStamp.Format("20060102150405")
	strTimeStamp := string(timeFormatted)
	Device1.DBInsertionTimeStamp = strTimeStamp

	Device1.APIName = "CiscoEoX"

	// Insert into table - smartcentre_api_data

	// if strings.HasPrefix(Device1.Serial_number, "NA-") {
	// 	Device1.SubcomponentSN = Device1.Serial_number
	// }

	av, err := dynamodbattribute.MarshalMap(Device1)
	if err != nil {
		fmt.Println("Got error marshalling item:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	tableName := "smartcentre_eox_api_data"

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		// Conditional logic to re-attempt DynamoDB put if intiial attempt fails
		// Two seconds of delay is added to attempt to clear any network conditions that may have caused failure
		fmt.Println("within put reattempt conditional")
		time.Sleep(2 * time.Second)
		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

func apiAuthToken() {
	// API tokens remove for GitHub share at the end of these strings
	url1 := "https://cloudsso.cisco.com/as/token.oauth2?grant_type=client_credentials&client_id=shzy2tm8u5b6we5a7zqscr3k&client_secret"
	url2 := "https://cloudsso.cisco.com/as/token.oauth2?grant_type=client_credentials&client_id=bghqpdv3tqg66eedcuqd3hqr&client_secret"

	method := "POST"

	client := &http.Client{}

	// Logic to utilize two available API accounts to ensure one does not become exhausted during a large number
	// of component requests
	apiIterator = 1

	req, err := http.NewRequest(method, url1, nil)

	if apiIterator%2 != 0 {
		req, err = http.NewRequest(method, url2, nil)
		if err != nil {
			fmt.Println(err)
		}
	}

	res, err := client.Do(req)

	// The auth API experiences timeouts occasionally in the network request.  To ensure the flow continues the
	// logc below will re-attempt the API request four times before eventually erroring out.  A two second delay is
	// added between each of the two additional requests in an attempt to allow network timeouts to subside.
	if err != nil {
		fmt.Println("within res reattempt conditional")
		time.Sleep(2 * time.Second)
		res, err = client.Do(req)
		if err != nil {
			fmt.Println("within res reattempt2 conditional")
			time.Sleep(2 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt3 conditional")
			time.Sleep(2 * time.Second)
			res, err = client.Do(req)
		}
		if err != nil {
			fmt.Println("within res reattempt4 conditional")
			time.Sleep(2 * time.Second)
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
	RetrieveCurrentCustomers()
}
