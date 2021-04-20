package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type AccessToken struct {
	Access_token string
	Token_type   string
	Expires_in   int
}

type Device struct {
	UUID                     string
	Timestamp                int
	SystemModel              string
	EndpointModel            string
	SystemVersion            string
	FirmwareVersion          string
	Manufacturer             string
	DisplayName              string
	Description              string
	Categories               string
	SWproductID              string
	CompanyID                string
	BillingName              string
	ServiceBoard             string
	TicketType               string
	AgreementID              string
	SnmpVersion              string
	SubType                  string
	NetworkAddress           string
	NetworkMask              string
	MacAddress               string
	ComponentSerialNumber    string
	ParentSerialNumber       string
	IsParentDevice           string
	EntPhysicalDescription   string
	Memory                   string
	Vendor                   string
	HostSerialNumber         string
	Systeminfo               string
	DBInsertionTimeStamp     string
	APIName                  string
	UCVersion                string
	UCVersionParsed1         string
	UCVersionParsed2         string
	ExpresswayVerParsedFinal string
	ApiUrl                   string
	BugPaginationIndexCount  int
	BugID                    string
	BugHeadline              string
	BugSeverity              string
	BugStatus                string
	BugRecordID              string
	BugKnownAffectedReleases string
	BugKnownFixedReleases    string
	BugSupportCaseCount      string
	BugLastModifiedDate      string
	BugAssociatedURL         string
	BugPlatformMatched       string
	IosVerFinal              string
}

type Company struct {
	CompanyID   string
	BillingName string
}

type Bugs struct {
	Bugs                       []map[string]string
	Pagination_response_record map[string]int
}

///Variables//
var Company1 = Company{}
var Device1 = Device{}
var Bugs1 = Bugs{}
var authToken = "null"

/////////////

func RetrieveCurrentCustomers() {
	//Refresh Cisco API Token at beginning of each customer query loop
	apiAuthToken()

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
		apiAuthToken()
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

	var result, err2 = svc.Query(queryInput)

	if err2 != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}

	// Iterate thru DynamoDB Returned Rows
	for _, v := range result.Items {
		Device1 = Device{}
		err = dynamodbattribute.UnmarshalMap(v, &Device1)
		if err != nil {
			fmt.Println("Query API call failed:")
			fmt.Println((err.Error()))
			os.Exit(1)
		}

		if strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") && len(Device1.UCVersion) > 0 {
			// Logic below parses the DynamoDB LMDevice table > UCVersion field into Device1 struct fields and to allow
			// the construction of an API URL such as the following example
			// https://api.cisco.com/bug/v2.0/bugs/product_name/Communications%20Manager%20Version%2010.5/affected_releases/10.5(2)?modified_date=3
			tempVersionParsed := strings.Split(Device1.UCVersion, ".")
			Device1.UCVersionParsed1 = tempVersionParsed[0] + "." + tempVersionParsed[1]
			// Following conditional is necessary due to API fails to return data in Unity Connection and Expressway
			// instances in which the version is 12.5(0) - for example - and the version in the API URL is sent as 12.5(0).
			// Instead the API only returns data if the version is sent as 12.5.
			if strings.HasPrefix(Device1.SWproductID, "UNITYCN") && tempVersionParsed[2] == "0" {
				Device1.UCVersionParsed2 = tempVersionParsed[0] + "." + tempVersionParsed[1]
			} else if strings.HasPrefix(Device1.SWproductID, "SW-EXP") {
				// Currently UC version info for Expressway in LogicMonitor appear such as: systemkey-1.0-oak-v12.5.6-rc-2
				// Following parsing techniques will render the final version used in URL as 12.5.6 (I.e.)
				expresswayVerParsedV1 := strings.Split(Device1.UCVersion, "-v")
				expresswayVerParsedV2 := strings.Split(expresswayVerParsedV1[1], "-")
				Device1.ExpresswayVerParsedFinal = expresswayVerParsedV2[0]
			} else {
				Device1.UCVersionParsed2 = tempVersionParsed[0] + "." + tempVersionParsed[1] + "(" + tempVersionParsed[2] + ")"
			}
			BugAPIBySoftwareVerison()
		} else if !strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") {
			BugAPIBySoftwareVerison()
		}
	}
}

func BugAPIBySoftwareVerison() {

	fmt.Println("########################")
	fmt.Println("#####Within BugAPIBySoftwareVerison function##########")
	fmt.Println(Device1.DisplayName)
	fmt.Println(Device1.ParentSerialNumber)
	fmt.Println(Device1.SWproductID)
	fmt.Println("Device1.EndpointModel: ", Device1.EndpointModel)
	fmt.Println("Device1.IsParentDevice: ", Device1.IsParentDevice)
	fmt.Println("Device1.UCVersion", Device1.UCVersion)
	fmt.Println("Device1 complete: ", Device1)
	fmt.Println("########################")

	if strings.HasPrefix(Device1.SWproductID, "CUCM") {
		Device1.ApiUrl = "https://api.cisco.com/bug/v3.0/bugs/product_name/Communications%20Manager%20Version%20" + Device1.UCVersionParsed1 + "/affected_releases/" + Device1.UCVersionParsed2 + "?modified_date=2"
		Device1.BugPlatformMatched = "Cisco_CUCM"
	} else if strings.HasPrefix(Device1.SWproductID, "UNITYCN") {
		// Note that the URL used for Unity Connection is intentionally different than other UC products in that the version is not
		// included and only affected_releases.  This is manifest from testing and results revealed that Unity Connection bugs
		// are often not return in circumstances in which the version is 12.5(0) - for example - and the version is included in
		// the URL.  However removal of the version parameter in the URL seems to allow successful retrieval in all cases.  For
		// other UC products (I.e. CUCM) bug data is not returned if the version parameter is not included.
		Device1.ApiUrl = "https://api.cisco.com/bug/v3.0/bugs/product_name/Cisco%20Unity%20Connection/affected_releases/" + Device1.UCVersionParsed2 + "?modified_date=2"
	} else if strings.HasPrefix(Device1.SWproductID, "CCX") {
		Device1.ApiUrl = "https://api.cisco.com/bug/v3.0/bugs/product_name/Cisco%20Unified%20Contact%20Center%20Express%20" + Device1.UCVersionParsed1 + "/affected_releases/" + Device1.UCVersionParsed2 + "?modified_date=2"
		Device1.BugPlatformMatched = "Cisco_UCCX"
	} else if strings.Contains(Device1.Manufacturer, "Cisco") && Device1.EndpointModel != "" && Device1.IsParentDevice == "true" && !strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") {
		iosVerParsedV1 := strings.Split(Device1.Systeminfo, "Version ")
		iosVerParsedV2 := strings.Split(iosVerParsedV1[1], ",")
		Device1.IosVerFinal = iosVerParsedV2[0]
		Device1.ApiUrl = "https://api.cisco.com/bug/v3.0/bugs/products/product_id/" + Device1.EndpointModel + "/software_releases/" + Device1.IosVerFinal + "?modified_date=2"
		Device1.BugPlatformMatched = "Cisco_IOS_Device"
	} else if strings.HasPrefix(Device1.SWproductID, "SW-EXP") {
		Device1.ApiUrl = "https://api.cisco.com/bug/v3.0/bugs/product_name/Cisco%20TelePresence%20Video%20Communication%20Server%20%28VCS%29/affected_releases/X" + Device1.ExpresswayVerParsedFinal + "?modified_date=2"
		Device1.BugPlatformMatched = "Cisco_Expressway"
	} else {
		return
	}

	fmt.Println("entering Main API lookup")
	url := Device1.ApiUrl
	fmt.Println("URL: ", url)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "Bearer "+authToken)
	req.Header.Add("Accept", "*/*")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	Bugs1 := Bugs{}

	json.Unmarshal(body, &Bugs1)

	// Capture of number of records in bugs pagination record to device struct to allow iteration thru multiple
	// pages of returned data.  A seperate API call request must be made for each page index.
	Device1.BugPaginationIndexCount = Bugs1.Pagination_response_record["last_index"]
	//fmt.Println("Bugs page index: ", Bugs1.Pagination_response_record["last_index"])

	// Outer for loop that iterates thru page indexes of the API pagination
	for i := 1; i <= Device1.BugPaginationIndexCount; i++ {
		// Final URL composed of the product ID - formulated prior - and the current pagination page
		url := Device1.ApiUrl + "&page_index=" + strconv.Itoa(i)
		fmt.Println(url)
		method := "GET"

		client := &http.Client{}
		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			fmt.Println(err)
		}

		req.Header.Add("Authorization", "Bearer "+authToken)
		req.Header.Add("Accept", "*/*")

		res, err := client.Do(req)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		Bugs1 := Bugs{}

		json.Unmarshal(body, &Bugs1)

		// Inner for loop that iterates thru the bugs returned on each pagination page
		for _, v := range Bugs1.Bugs {
			// Cisco Bug API for IOS devices returns a match if the software version is listed either in known_affected_releases
			// or known_fixed_releases.  The following if conditional ensures that only matches on known affected releases are
			// written to the database
			if Device1.BugPlatformMatched == "Cisco_IOS_Device" {
				if !strings.Contains(v["known_affected_releases"], Device1.IosVerFinal) {
					fmt.Println("Bug DB insert skipped - current IOS version not an affected version")
					continue
				}
			}
			Device1.BugID = v["bug_id"]
			Device1.BugHeadline = v["headline"]
			Device1.BugSeverity = v["severity"]
			Device1.BugStatus = v["status"]
			Device1.Description = v["description"]
			Device1.BugRecordID = v["id"]
			Device1.BugKnownAffectedReleases = v["known_affected_releases"]
			Device1.BugKnownFixedReleases = v["known_fixed_releases"]
			Device1.BugSupportCaseCount = v["support_case_count"]
			Device1.BugLastModifiedDate = v["last_modified_date"]
			Device1.BugAssociatedURL = "https://bst.cloudapps.cisco.com/bugsearch/bug/" + Device1.BugID

			uuid, err := exec.Command("uuidgen").Output()
			if err != nil {
				log.Fatal(err)
			}

			Device1.UUID = string(uuid)

			timeStamp := time.Now()
			timeFormatted := timeStamp.Format("20060102150405")
			strTimeStamp := string(timeFormatted)
			Device1.DBInsertionTimeStamp = strTimeStamp

			Device1.APIName = "CiscoBugsAPI"

			fmt.Println("Device1 following API lookup: ")
			fmt.Println(Device1.DisplayName)
			fmt.Println(Device1.BugPaginationIndexCount)
			fmt.Println(Device1.BugID)
			fmt.Println(Device1.BugSeverity)
			fmt.Println(Device1.BugRecordID)
			fmt.Println("Page index: ", i)
			fmt.Println("##############################")

			// DynamoDB insert for each bug identified on current device
			// Only update the database of the status of the bug is O (Open) and F (Fixed).  If the status is T (Terminated)
			// it will not be written to the DB per the SMARTcentre's preferences.
			if Device1.BugStatus == "O" || Device1.BugStatus == "F" {
				apiUpdateDynamoDB()
			} else {
				continue
			}
		}
	}
}

func apiUpdateDynamoDB() {

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Update item in table smartcentre_eol_eos_component
	tableName := "smartcentre_bugs_data"

	av, err := dynamodbattribute.MarshalMap(Device1)
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

	fmt.Println("Successfully updated")

}

func apiAuthToken() {
	// Removing API key from end of string for GitHub upload
	url := "https://cloudsso.cisco.com/as/token.oauth2?grant_type=client_credentials&client_id=bghqpdv3tqg66eedcuqd3hqr&client_secret"

	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	AccessToken1 := AccessToken{}

	json.Unmarshal(body, &AccessToken1)

	authToken = AccessToken1.Access_token

}

func main() {
	RetrieveCurrentCustomers()
}
