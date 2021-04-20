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
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type AccessToken struct {
	Access_token string
	Token_type   string
	Expires_in   int
}

type Psirt struct {
	Advisories []struct {
		AdvisoryID    string
		AdvisoryTitle string
		BugIDs        []string
		IpsSignatures []struct {
			LegacyIpsID     string
			ReleaseVersion  string
			SoftwareVersion string
			LegacyIpsURL    string
		}
		Cves           []string
		CvrfURL        string
		CvssBaseScore  string
		Cwe            []string
		FirstPublished string
		LastUpdated    string
		ProductNames   []string
		PublicationURL string
		Sir            string
		Summary        string
	}
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
	DBInsertionTimeStamp   string
	IosVerFinal            string
	NXOSFinal              string
	IOSXEFinal             string
	UCPlatform             string
	UCVersion              string
	UCProductName          string
	PlatformIdentifier     string
	CiscoID                string
	APIName                string
	ApiUrl                 string
	PSIRTAdvisoryID        string
	PSIRTAdvisoryTitle     string
	PSIRTBugIDs            []string
	PSIRTIpsSignatures     []struct {
		LegacyIpsID     string
		ReleaseVersion  string
		SoftwareVersion string
		LegacyIpsURL    string
	}
	PSIRTCves           []string
	PSIRTCvrfURL        string
	PSIRTCvssBaseScore  string
	PSIRTCwe            []string
	PSIRTFirstPublished string
	PSIRTLastUpdated    string
	PSIRTProductNames   []string
	PSIRTPublicationURL string
	PSIRTSir            string
	PSIRTSummary        string
}

type PsirtIsolated struct {
	APIName            string
	APIURL             string
	PSIRTAdvisoryID    string
	PSIRTAdvisoryTitle string
	PSIRTBugIDs        []string
	PSIRTIpsSignatures []struct {
		LegacyIpsID     string
		ReleaseVersion  string
		SoftwareVersion string
		LegacyIpsURL    string
	}
	PSIRTCves                []string
	PSIRTCvrfURL             string
	PSIRTCvssBaseScore       string
	PSIRTCwe                 []string
	PSIRTFirstPublished      string
	PSIRTLastUpdated         string
	PSIRTProductNames        []string
	PSIRTPublicationURL      string
	PSIRTSir                 string
	PSIRTSummary             string
	DBInsertionTimeStamp     string
	Description              string
	EndpointModel            string
	EndpointModelPlusVersion string
	SWproductID              string
	UCVersion                string
	IosVerFinal              string
	NXOSFinal                string
	IOSXEFinal               string
	PlatformIdentifier       string
	CiscoID                  string
	UUID                     string
}

type Company struct {
	CompanyID   string
	BillingName string
}

///Variables//
var Company1 = Company{}
var Device1 = Device{}
var authToken = "null"
var Psirt1 Psirt
var ucPlatform = "null"
var iosPlatformAndVersion = "null"
var PsirtIsolated1 = PsirtIsolated{}

//////////////

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
		DynamoDBExistingRecordCheck(Device1)
	}
}

func DynamoDBExistingRecordCheck(Device1 Device) {
	// Purpose of this function is to detemine if bug data exists in the smartcentre_api_master table.
	// Instead of writing the PSIRT data - which can be considerable for IOS and other devices - for device
	// a master DB has been created that will house all device/type and versions for SMARTcentre customers and
	// thus only one DB record is created for each version and then customer devices tables and joined with this
	// master table.  The master table is populated in this code by iterrating thru all customers/all devices >
	// checking to see if a record for that device type/software version exist in the DB and will skip over API
	// lookup and insertion if the record exists and it is not more than 24 hours old.

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	queryInput := &dynamodb.QueryInput{}

	// Validation of Cisco collab software pre-existing in the master API DB

	if strings.HasPrefix(Device1.SWproductID, "CUCM") || strings.HasPrefix(Device1.SWproductID, "UNITYCN") || strings.HasPrefix(Device1.SWproductID, "CCX") || strings.HasPrefix(Device1.SWproductID, "SW-EXP") || strings.HasPrefix(Device1.SWproductID, "ER") {
		Device1.UCPlatform = Device1.SWproductID

		if strings.HasPrefix(Device1.SWproductID, "CUCM") {
			Device1.PlatformIdentifier = "CollabServer__CiscoUnifiedCommunicationsManager"
		} else if strings.HasPrefix(Device1.SWproductID, "UNITYCN") {
			Device1.PlatformIdentifier = "CollabServer__CiscoUnityConnection"
		} else if strings.HasPrefix(Device1.SWproductID, "CCX") {
			Device1.PlatformIdentifier = "CollabServer__CiscoUnifiedContactCenterExpress"
		} else if strings.HasPrefix(Device1.SWproductID, "SW-EXP") {
			Device1.PlatformIdentifier = "CollabServer__CiscoExpressway"
		} else if strings.HasPrefix(Device1.SWproductID, "ER") {
			Device1.PlatformIdentifier = "CollabServer__CiscoEmergencyResponder"
		} else {
			return
		}

		fmt.Println("Within UC existing record check - PlatformIdentifier: ", Device1.PlatformIdentifier)
		queryInput = &dynamodb.QueryInput{
			TableName: aws.String("smartcentre_api_master"),
			IndexName: aws.String("APIName-PlatformIdentifier-index"),
			KeyConditions: map[string]*dynamodb.Condition{
				"APIName": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String("CiscoPsirtAPI"),
						},
					},
				},
				"PlatformIdentifier": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(Device1.PlatformIdentifier),
						},
					},
				},
			},
		}
		// Validation of Cisco NX-OS software pre-existing in the master API DB
	} else if strings.Contains(Device1.Systeminfo, "Version") && strings.Contains(Device1.Manufacturer, "Cisco") && Device1.EndpointModel != "" && Device1.IsParentDevice == "true" && !strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") && strings.Contains(Device1.Systeminfo, "NX-OS") {
		nxosVerParsedV1 := strings.Split(Device1.Systeminfo, "Version ")
		nxosVerParsedV2 := strings.Split(nxosVerParsedV1[1], ",")
		Device1.NXOSFinal = nxosVerParsedV2[0]
		Device1.PlatformIdentifier = "NXOS__" + Device1.NXOSFinal
		fmt.Println("Within NX-OS existing record check - PlatformIdentifier: ", Device1.PlatformIdentifier)
		queryInput = &dynamodb.QueryInput{
			TableName: aws.String("smartcentre_api_master"),
			IndexName: aws.String("APIName-PlatformIdentifier-index"),
			KeyConditions: map[string]*dynamodb.Condition{
				"APIName": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String("CiscoPsirtAPI"),
						},
					},
				},
				"PlatformIdentifier": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(Device1.PlatformIdentifier),
						},
					},
				},
			},
		}
		// Validation of Cisco IOS-XE or Cisco IOS Everest software pre-existing in the master API DB
	} else if strings.Contains(Device1.Systeminfo, "Version") && strings.Contains(Device1.Manufacturer, "Cisco") && Device1.EndpointModel != "" && Device1.IsParentDevice == "true" && !strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") && (strings.Contains(Device1.Systeminfo, "IOS-XE") || strings.Contains(Device1.Systeminfo, "Everest") || strings.Contains(Device1.Systeminfo, "Gibraltar")) {
		iosxeVerParsedV1 := strings.Split(Device1.Systeminfo, "Version ")
		iosxeVerParsedV2 := strings.Split(iosxeVerParsedV1[1], " ")
		Device1.IOSXEFinal = iosxeVerParsedV2[0]
		// Cisco IOS Everest version info in Device1.Systeminfo field has a trailing comma which is removed below for
		// proper URL formulation
		if strings.HasSuffix(Device1.IOSXEFinal, ",") {
			iosxeVerParsedV3 := strings.Split(Device1.IOSXEFinal, ",")
			Device1.IOSXEFinal = iosxeVerParsedV3[0]
		}
		Device1.PlatformIdentifier = "IOSXE__" + Device1.IOSXEFinal
		fmt.Println("IOSXE Final: ", Device1.IOSXEFinal)
		fmt.Println("Within IOSXE existing record check -PlatformIdentifier: ", Device1.PlatformIdentifier)
		queryInput = &dynamodb.QueryInput{
			TableName: aws.String("smartcentre_api_master"),
			IndexName: aws.String("APIName-PlatformIdentifier-index"),

			KeyConditions: map[string]*dynamodb.Condition{
				"APIName": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String("CiscoPsirtAPI"),
						},
					},
				},
				"PlatformIdentifier": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(Device1.PlatformIdentifier),
						},
					},
				},
			},
		}
		// Validation of Cisco IOS software pre-existing in the master API DB
	} else if strings.Contains(Device1.Systeminfo, "Version") && strings.Contains(Device1.Manufacturer, "Cisco") && Device1.EndpointModel != "" && Device1.IsParentDevice == "true" && !strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") {
		iosVerParsedV1 := strings.Split(Device1.Systeminfo, "Version ")
		iosVerParsedV2 := strings.Split(iosVerParsedV1[1], ",")
		Device1.IosVerFinal = iosVerParsedV2[0]
		Device1.PlatformIdentifier = "IOS__" + Device1.IosVerFinal
		fmt.Println("Within IOS existing record check -PlatformIdentifier: ", Device1.PlatformIdentifier)
		queryInput = &dynamodb.QueryInput{
			TableName: aws.String("smartcentre_api_master"),
			IndexName: aws.String("APIName-PlatformIdentifier-index"),
			KeyConditions: map[string]*dynamodb.Condition{
				"APIName": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String("CiscoPsirtAPI"),
						},
					},
				},
				"PlatformIdentifier": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String(Device1.PlatformIdentifier),
						},
					},
				},
			},
		}
	} else {
		// If the current device is neither Cisco collab nor IOS product - return out of the function and proceed to
		// next device.  No further query will be of non collab/IOS products in current scheme.
		return
	}

	var result, _ = svc.Query(queryInput)

	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}

	// Iterate thru DynamoDB Returned Rows
	if len(result.Items) > 0 {
		fmt.Println("####################Match in master API table###########")
		//  Need to add in logic to check how old bug record is and update if older than 24 hours
		// Iterate thru DynamoDB Returned Rows
		for _, v := range result.Items {
			PsirtIsolated1 = PsirtIsolated{}
			err = dynamodbattribute.UnmarshalMap(v, &PsirtIsolated1)
			if err != nil {
				fmt.Println("Query API call failed:")
				fmt.Println((err.Error()))
				os.Exit(1)
			}

			Device1.PSIRTAdvisoryID = PsirtIsolated1.PSIRTAdvisoryID
			Device1.PSIRTAdvisoryTitle = PsirtIsolated1.PSIRTAdvisoryTitle
			Device1.PSIRTBugIDs = PsirtIsolated1.PSIRTBugIDs
			Device1.PSIRTCves = PsirtIsolated1.PSIRTCves
			Device1.PSIRTCvrfURL = PsirtIsolated1.PSIRTCvrfURL
			Device1.PSIRTCvssBaseScore = PsirtIsolated1.PSIRTCvssBaseScore
			Device1.PSIRTCwe = PsirtIsolated1.PSIRTCwe
			Device1.PSIRTFirstPublished = PsirtIsolated1.PSIRTFirstPublished
			Device1.PSIRTLastUpdated = PsirtIsolated1.PSIRTLastUpdated
			Device1.PSIRTProductNames = PsirtIsolated1.PSIRTProductNames
			Device1.PSIRTPublicationURL = PsirtIsolated1.PSIRTPublicationURL
			Device1.PSIRTSir = PsirtIsolated1.PSIRTSir
			Device1.PSIRTSummary = PsirtIsolated1.PSIRTSummary
			Device1.CiscoID = PsirtIsolated1.PSIRTAdvisoryID

			Device1.APIName = "CiscoPsirtAPI"

			uuid, err := exec.Command("uuidgen").Output()
			if err != nil {
				log.Fatal(err)
			}

			Device1.UUID = string(uuid)

			apiUpdateDynamoDB(Device1)
		}
	} else {
		fmt.Println("####################No match in master API table###########")
		if len(Device1.UCPlatform) > 0 {
			PsirtAPIByProductName(Device1)
			// If the device type is either IOS, IOS-XE, or NX-OS the flow is sent to the PsirtAPIByIOSVerison function
			// Unique API endpoints exist for IOS, IOS-XE, and NX-OS and thus the API URL used will be determined within
			// the PsirtAPIByIOSVerison function
		} else if len(Device1.IosVerFinal) > 0 || len(Device1.NXOSFinal) > 0 || len(Device1.IOSXEFinal) > 0 {
			PsirtAPIByIOSVerison(Device1)
		}
	}
}

func PsirtAPIByIOSVerison(Device1 Device) {
	// Condition attempts to identify the device as a Cisco networking IOS device
	if strings.Contains(Device1.Manufacturer, "Cisco") && Device1.EndpointModel != "" && Device1.IsParentDevice == "true" && !strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") {
		// Parse the IOS version from the Systeminfo device info made available from the LogicMonitor API
		fmt.Println("##########Device1 within API by IOS version function#########")
		fmt.Println("Device1.DisplayName: ", Device1.DisplayName)
		fmt.Println("#######################################")

		// Construction of URL to use in IOS API request.  Sample/comments out version is shown for reference of how
		//  URL should be structured as the resultant.
		// url := "https://api.cisco.com/security/advisories/ios?version=15.4(3)M3"
		if len(Device1.IosVerFinal) > 0 {
			url := "https://api.cisco.com/security/advisories/ios?version=" + Device1.IosVerFinal
			fmt.Println("IOS URL: ", url)
			Device1.ApiUrl = url
			// Construction of URL to use in NX-OS API request.  Sample/comments out version is shown for reference of how
			//  URL should be structured as the resultant.
			// url := "https://api.cisco.com/security/advisories/nxos?version=5.2(1)N1(4)"
		} else if len(Device1.NXOSFinal) > 0 {
			url := "https://api.cisco.com/security/advisories/nxos?version=" + Device1.NXOSFinal
			fmt.Println("NX-OS URL: ", url)
			Device1.ApiUrl = url
			// Construction of URL to use in IOS-XE API request.  Sample/comments out version is shown for reference of how
			//  URL should be structured as the resultant.
			// url := "hhttps://api.cisco.com/security/advisories/iosxe?version=03.04.02.SG"
		} else if len(Device1.IOSXEFinal) > 0 {
			url := "https://api.cisco.com/security/advisories/iosxe?version=" + Device1.IOSXEFinal
			fmt.Println("IOS-XE URL: ", url)
			Device1.ApiUrl = url
		}

	} else if strings.HasPrefix(Device1.ComponentSerialNumber, "NA_") && len(Device1.UCVersion) > 0 {
		// Conditional scrutiny for devices that fall thru IOS initial If statement isolates Cisco UC servers and upon
		// match send to a function that will conduct API request by product name (I.e.)
		PsirtAPIByProductName(Device1)
	}

	// Final check if IosVerFinal - essential for this API request is populated.  If it is not - return and add flow
	// for device.
	if len(Device1.IosVerFinal) == 0 && len(Device1.NXOSFinal) == 0 && len(Device1.IOSXEFinal) == 0 {
		return
	}

	fmt.Println("entering ios API lookup")

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, Device1.ApiUrl, nil)

	if err != nil {
		fmt.Println(err)
	}

	// req.Header.Add("Authorization", "Bearer pUQq2tJvUIut80CajWoevz5Xy811")
	req.Header.Add("Authorization", "Bearer "+authToken)
	req.Header.Add("Accept", "*/*")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	Psirt1 = Psirt{}
	json.Unmarshal(body, &Psirt1)

	now := time.Now()
	for _, v := range Psirt1.Advisories {

		Device1.PSIRTAdvisoryID = v.AdvisoryID
		Device1.PSIRTAdvisoryTitle = v.AdvisoryTitle
		Device1.PSIRTBugIDs = v.BugIDs
		Device1.PSIRTCves = v.Cves
		Device1.PSIRTCvrfURL = v.CvrfURL
		Device1.PSIRTCvssBaseScore = v.CvssBaseScore
		Device1.PSIRTCwe = v.Cwe
		Device1.PSIRTFirstPublished = v.FirstPublished
		Device1.PSIRTLastUpdated = v.LastUpdated
		Device1.PSIRTProductNames = v.ProductNames
		Device1.PSIRTPublicationURL = v.PublicationURL
		Device1.PSIRTSir = v.Sir
		Device1.PSIRTSummary = v.Summary

		Device1.APIName = "CiscoPsirtAPI"

		fmt.Println("############Device PSIRT Data############")
		fmt.Println("PSIRT Title: ", Device1.PSIRTAdvisoryTitle)
		fmt.Println("PSIRT Last updated: ", Device1.PSIRTLastUpdated)
		fmt.Println("PSIRT BUG ids: ", Device1.PSIRTBugIDs)

		fmt.Println("########################")

		uuid, err := exec.Command("uuidgen").Output()
		if err != nil {
			log.Fatal(err)
		}

		Device1.UUID = string(uuid)

		// timeStamp := time.Now()
		// timeFormatted := timeStamp.Format("20060102150405")
		// strTimeStamp := string(timeFormatted)
		// Device1.DBInsertionTimeStamp = strTimeStamp
		Device1.DBInsertionTimeStamp = time.Now().Format(time.RFC3339)

		layout := "2006-01-02T15:04:05"
		t, err := time.Parse(layout, Device1.PSIRTLastUpdated)
		if err != nil {
			fmt.Println(err)
		}

		// Check if PSIRT is older than 1 year old.  Only alerts 1 year old or newer will be writen to database.
		if t.Before(now.AddDate(-1, 0, 0)) {
			continue
		}

		apiUpdateDynamoDB(Device1)
		apiUpdateMasterDB(Device1)
	}

}

func PsirtAPIByProductName(Device1 Device) {

	fmt.Println("entering product name API lookup")
	// Identification of Cisco UC Product Name required for he PSIRT URL. Product names can be found in PSIRT
	// bulletins such as the following example - https://tools.cisco.com/security/center/contentxml/CiscoSecurityAdvisory/cisco-sa-20171115-vos/cvrf/cisco-sa-20171115-vos_cvrf.xml
	if strings.HasPrefix(Device1.SWproductID, "CUCM") {
		Device1.UCProductName = "Cisco%20Unified%20Communications%20Manager"
	} else if strings.HasPrefix(Device1.SWproductID, "UNITYCN") {
		Device1.UCProductName = "Cisco%20Unity%20Connection"
	} else if strings.HasPrefix(Device1.SWproductID, "CCX") {
		Device1.UCProductName = "Cisco%20Unified%20Contact%20Center%20Express"
	} else if strings.HasPrefix(Device1.SWproductID, "SW-EXP") {
		Device1.UCProductName = "Cisco%20Expressway"
	} else if strings.HasPrefix(Device1.SWproductID, "ER") {
		Device1.UCProductName = "Cisco%20Emergency%20Responder"
	} else {
		return
	}

	// Construction of URL to use in API request.  Sample/comments out version is shown for reference of how
	//  URL should be structured as the resultant.
	// url := "https://api.cisco.com/security/advisories/product?product=Cisco Unified Communications Manager"
	url := "https://api.cisco.com/security/advisories/product?product=" + Device1.UCProductName
	fmt.Println(url)
	Device1.ApiUrl = url
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, Device1.ApiUrl, nil)

	if err != nil {
		fmt.Println(err)
	}

	// req.Header.Add("Authorization", "Bearer pUQq2tJvUIut80CajWoevz5Xy811")
	req.Header.Add("Authorization", "Bearer "+authToken)
	req.Header.Add("Accept", "*/*")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	Psirt1 = Psirt{}
	json.Unmarshal(body, &Psirt1)

	now := time.Now()
	for _, v := range Psirt1.Advisories {

		Device1.PSIRTAdvisoryID = v.AdvisoryID
		Device1.PSIRTAdvisoryTitle = v.AdvisoryTitle
		Device1.PSIRTBugIDs = v.BugIDs
		Device1.PSIRTCves = v.Cves
		Device1.PSIRTCvrfURL = v.CvrfURL
		Device1.PSIRTCvssBaseScore = v.CvssBaseScore
		Device1.PSIRTCwe = v.Cwe
		Device1.PSIRTFirstPublished = v.FirstPublished
		Device1.PSIRTLastUpdated = v.LastUpdated
		Device1.PSIRTProductNames = v.ProductNames
		Device1.PSIRTPublicationURL = v.PublicationURL
		Device1.PSIRTSir = v.Sir
		Device1.PSIRTSummary = v.Summary

		uuid, err := exec.Command("uuidgen").Output()
		if err != nil {
			log.Fatal(err)
		}

		Device1.UUID = string(uuid)

		// timeStamp := time.Now()
		// timeFormatted := timeStamp.Format("20060102150405")
		// strTimeStamp := string(timeFormatted)
		// Device1.DBInsertionTimeStamp = strTimeStamp
		Device1.DBInsertionTimeStamp = time.Now().Format(time.RFC3339)

		Device1.APIName = "CiscoPsirtAPI"

		layout := "2006-01-02T15:04:05"
		t, err := time.Parse(layout, Device1.PSIRTLastUpdated)
		if err != nil {
			fmt.Println(err)
		}

		// Check if PSIRT is older than 1 year old.  Only alerts 1 year old or newer will be writen to database.
		if t.Before(now.AddDate(-1, 0, 0)) {
			continue
		}

		apiUpdateDynamoDB(Device1)
		apiUpdateMasterDB(Device1)

	}

}

// func apiUpdateDynamoDB(SubcomponentSN string, Companyid string, PsirtDeviceArray1 PsirtDeviceArray) {
func apiUpdateDynamoDB(Device1 Device) {

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	Device1.UUID = string(uuid)
	fmt.Println("UUID: ", Device1.UUID)

	fmt.Println("#######Entered DB update function############")

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Update item in table smartcentre_eol_eos_component
	tableName := "smartcentre_psirt_data"

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

	fmt.Println("Successfully updated device table for customer: ", Device1.CompanyID)

}

func apiUpdateMasterDB(Device1 Device) {
	fmt.Println("############Updating bug master DynamoDB table############")
	PsirtIsolated1 = PsirtIsolated{}

	PsirtIsolated1.PlatformIdentifier = Device1.PlatformIdentifier

	PsirtIsolated1.PSIRTCvssBaseScore = Device1.PSIRTCvssBaseScore
	PsirtIsolated1.PSIRTAdvisoryTitle = Device1.PSIRTAdvisoryTitle
	PsirtIsolated1.PSIRTAdvisoryID = Device1.PSIRTAdvisoryID
	PsirtIsolated1.PSIRTFirstPublished = Device1.PSIRTFirstPublished
	PsirtIsolated1.PSIRTLastUpdated = Device1.PSIRTLastUpdated
	PsirtIsolated1.PSIRTCvrfURL = Device1.PSIRTCvrfURL
	PsirtIsolated1.PSIRTProductNames = Device1.PSIRTProductNames
	PsirtIsolated1.PSIRTSir = Device1.PSIRTSir
	PsirtIsolated1.PSIRTBugIDs = Device1.PSIRTBugIDs
	PsirtIsolated1.APIName = Device1.APIName
	PsirtIsolated1.APIURL = Device1.ApiUrl
	PsirtIsolated1.DBInsertionTimeStamp = Device1.DBInsertionTimeStamp
	PsirtIsolated1.Description = Device1.Description
	PsirtIsolated1.EndpointModel = Device1.EndpointModel
	PsirtIsolated1.SWproductID = Device1.SWproductID
	PsirtIsolated1.UCVersion = Device1.UCVersion
	PsirtIsolated1.IosVerFinal = Device1.IosVerFinal
	PsirtIsolated1.NXOSFinal = Device1.NXOSFinal
	PsirtIsolated1.IOSXEFinal = Device1.IOSXEFinal
	PsirtIsolated1.CiscoID = Device1.PSIRTAdvisoryID

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	PsirtIsolated1.UUID = string(uuid)

	// As SWproductID is an index sortt key and must be populated - set the value to the string of null for IOS devices
	// that would not have such an attribute otherwise
	if len(PsirtIsolated1.SWproductID) == 0 {
		PsirtIsolated1.SWproductID = "null"
	}

	// As EndpointModelPlusVersion is an index sor key and must be populated - set the value to the string of
	// null for PSIRT IOS devices that would not have such an attribute otherwise
	if len(PsirtIsolated1.EndpointModelPlusVersion) == 0 {
		PsirtIsolated1.EndpointModelPlusVersion = "null"
	}

	// As IosVerFinal is an index sor key and must be populated - set the value to the string of
	// null for PSIRT collab servers that would not have such an attribute otherwise
	if len(PsirtIsolated1.IosVerFinal) == 0 {
		PsirtIsolated1.IosVerFinal = "null"
	}

	// As NXOSFinal is an index sor key and must be populated - set the value to the string of
	// null for PSIRT collab servers that would not have such an attribute otherwise
	if len(PsirtIsolated1.NXOSFinal) == 0 {
		PsirtIsolated1.NXOSFinal = "null"
	}

	// As IOSXEFinal is an index sor key and must be populated - set the value to the string of
	// null for PSIRT collab servers that would not have such an attribute otherwise
	if len(PsirtIsolated1.IOSXEFinal) == 0 {
		PsirtIsolated1.IOSXEFinal = "null"
	}

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Update item in table smartcentre_eol_eos_component
	tableName := "smartcentre_api_master"

	av, err := dynamodbattribute.MarshalMap(PsirtIsolated1)
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

	fmt.Println("Successfully updated master API table")

}

func apiAuthToken() {
	// APIU keys removed at end of string for GitHub upload
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
