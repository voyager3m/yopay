// yopay project yopay.go
package yopay

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type YoAPI struct {
	/* Username
	   The Yo! Payments API Username
	   Required.
	   You may obtain the API Username from the web interface of your Payment Account.
	*/
	Username string

	/* Password
	    The Yo! Payments API Password
		Required.
		You may obtain the API Password from the web interface of your Payment Account.
	*/
	Password string

	/*	NonBlocking
		The Non Blocking Request variable
		Optional.
		Whether the connection to the Yo! Payments Gateway is maintained until your request is
		fulfilled. "FALSE" maintains the connection till the request is complete.
		Default: "FALSE"
	*/
	NonBlocking bool

	/*	ExternalReference
		The External Reference variable
		Optional.
		An External Reference is something which yourself and the beneficiary agree upon e.g. an invoice number
		Default: Empty
	*/
	ExternalReference string

	/* 	InternalReference
	    The Internal Reference variable
	    Optional.
		An Internal Reference is a reference code related to another Yo! Payments system transaction
		If you are unsure about the meaning of this field, leave it as empty
		Default: Empty
	*/
	InternalReference string

	/*	ProviderReferenceText
		The Provider Reference Text variable
		Optional.
		A text you wish to be present in any confirmation message which the mobile money provider
		network sends to the subscriber upon successful completion of the transaction.
		Some mobile money providers automatically send a confirmatory text message to the subscriber
		upon completion of transactions. This parameter allows you to provide some text which will
		be appended to any such confirmatory message sent to the subscriber.
		Default: Empty
	*/
	ProviderReferenceText string

	/* InstantNotificationUrl
	   The Instant Notification URL variable
	   Optional.
	   A valid URL which is notified as soon as funds are successfully deposited into your account
	   A payment notification will be sent to this URL.
	   It must be properly URL encoded.
	   e.g. http://ipnurl?key1=This+value+has+encoded+white+spaces&key2=value
	   Any special XML Characters must be escaped or your request will fail
	   e.g. http://ipnurl?key1=This+value+has+encoded+white+spaces&amp;key2=value
	    Default: Empty
	*/
	InstantNotificationUrl string

	/* FailureNotificationUrl
	   The Failure Notification URL variable
	   Optional.
	   A valid URL which is notified as soon as your deposit request fails
	   A failure notification will be sent to this URL.
	   It must be properly URL encoded.
	   e.g. http://failureurl?key1=This+value+has+encoded+white+spaces&key2=value
	   Any special XML Characters must be escaped or your request will fail
	   e.g. http://failureurl?key1=This+value+has+encoded+white+spaces&amp;key2=value
	   Default: Empty
	*/
	FailureNotificationUrl string

	/* AuthenticationSignatureBase64
	   The Authentication Signature Base64 variable
	   Optional.
	   It may be required to authenticate certain deposit requests.
	   Contact Yo! Payments support services for clarification on the cases where this parameter
	   is required.
	   Default: Empty
	*/
	AuthenticationSignatureBase64 string

	/* DepositTransactionType
		The Deposit Transaction Type variable
	    Optional.
	    Set to "PUSH" if following up on the status of a push deposit funds transaction
	    Set to "PULL" if following up on the status of a pull deposit funds transaction
	    Default: "PULL"
	    Options: "PULL", "PUSH"
	*/
	DepositTransactionType string

	/* YoUrl
	   The Yo Payments API URL
	   Required.
	   Default: "https://paymentsapi1.yo.co.ug/ybs/task.php"
	   Options:
	   * "https://paymentsapi1.yo.co.ug/ybs/task.php",
	   * "https://paymentsapi2.yo.co.ug/ybs/task.php",
	   * "https://41.220.12.206/services/yopaymentsdev/task.php" For Sandbox tests
	*/
	YoUrl string

	/** LastQuery, LastResponse, LastError - for debug
	 */
	LastQuery      string
	LastResponse   string
	LastError      string
	LastStatusCode int
}

type DepositResponse struct {
	Status                    string `xml:"Status"`
	StatusCode                string `xml:"StatusCode"`
	StatusMessage             string `xml:"StatusMessage"`
	TransactionStatus         string `xml:"TransactionStatus,omitempty"`
	ErrorMessageCode          string `xml:"ErrorMessageCode,omitempty"`
	ErrorMessage              string `xml:"ErrorMessage,omitempty"`
	TransactionReference      string `xml:"TransactionReference,omitempty"`
	MNOTransactionReferenceId string `xml:"MNOTransactionReferenceId,omitempty"`
	IssuedReceiptNumber       string `xml:"IssuedReceiptNumber,omitempty"`
}

type TransactionStatus struct {
	DepositResponse

	Amount                    string `xml:"Amount,omitempty"`
	AmountFormatted           string `xml:"AmountFormatted,omitempty"`
	CurrencyCode              string `xml:"CurrencyCode,omitempty"`
	TransactionInitiationDate string `xml:"TransactionInitiationDate,omitempty"`
	TransactionCompletionDate string `xml:"TransactionCompletionDate,omitempty"`
}

type BalanceResponse struct {
	Status           string `xml:"Status"`
	StatusCode       string `xml:"StatusCode"`
	StatusMessage    string `xml:"StatusMessage"`
	ErrorMessageCode string `xml:"ErrorMessageCode,omitempty"`
	ErrorMessage     string `xml:"TransactionReference,omitempty"`
	Balance          struct {
		Currency []struct {
			Code    string `xml:"Code"`
			Balance string `xml:"Balance"`
		} `xml:"Currency"`
	} `xml:"Balance"`
}

type MinistatementResponse struct {
	Status               string `xml:"Status"`
	StatusCode           string `xml:"StatusCode"`
	ErrorMessageCode     string `xml:"ErrorMessageCode,omitempty"`
	ErrorMessage         string `xml:"TransactionReference,omitempty"`
	TotalTransactions    string `xml:"TotalTransactions"`
	ReturnedTransactions string `xml:"ReturnedTransactions"`
	Transactions         struct {
		Transaction []struct {
			TransactionSystemId                string `xml:"TransactionSystemId"`
			TransactionReference               string `xml:"TransactionReference"`
			TransactionStatus                  string `xml:"TransactionStatus"`
			InitiationDate                     string `xml:"InitiationDate"`
			CompletionDate                     string `xml:"CompletionDate"`
			NarrativeBase64                    string `xml:"NarrativeBase64"`
			Currency                           string `xml:"Currency"`
			Amount                             string `xml:"Amount"`
			Balance                            string `xml:"Balance"`
			GeneralType                        string `xml:"GeneralType"`
			DetailedType                       string `xml:"DetailedType"`
			BeneficiaryMsisdn                  string `xml:"BeneficiaryMsisdn"`
			BeneficiaryBase64                  string `xml:"BeneficiaryBase64"`
			SenderMsisdn                       string `xml:"SenderMsisdn"`
			SenderBase64                       string `xml:"SenderBase64"`
			Base64TransactionExternalReference string `xml:"Base64TransactionExternalReference"`
			TransactionEntryDesignation        string `xml:"TransactionEntryDesignation"`
		} `xml:"Transaction"`
	} `xml:"Transactions"`
}

type VerifyAccountResponse struct {
	Response struct {
		Status string `xml:"Status"`
		Valid  string `xml:"Valid"`
	} `xml:"Response"`
}

type PaymentNotificationResponse struct {
	Verified    bool
	DateTime    string
	Amount      string
	Narrative   string
	NetworkRef  string
	ExternalRef string
	Msisdn      string
}

type PaymentFailureNotificationResponse struct {
	Verified                   bool
	FailedTransactionReference string
	TransactionInitDate        string
}

const pub_crt string = `-----BEGIN CERTIFICATE-----
MIIEvTCCA6WgAwIBAgIJAN3e7VqDg5zQMA0GCSqGSIb3DQEBBQUAMIGaMQswCQYD
VQQGEwJVRzEQMA4GA1UECBMHS2FtcGFsYTEQMA4GA1UEBxMHS2FtcGFsYTEbMBkG
A1UECgwSWW8hIFVnYW5kYSBMaW1pdGVkMRUwEwYDVQQLDAxZbyEgUGF5bWVudHMx
FTATBgNVBAMTDHd3dy55by5jby51ZzEcMBoGCSqGSIb3DQEJARYNaW5mb0B5by5j
by51ZzAeFw0xMzA4MDkwNTQyMTRaFw0yMzA4MDcwNTQyMTRaMIGaMQswCQYDVQQG
EwJVRzEQMA4GA1UECBMHS2FtcGFsYTEQMA4GA1UEBxMHS2FtcGFsYTEbMBkGA1UE
CgwSWW8hIFVnYW5kYSBMaW1pdGVkMRUwEwYDVQQLDAxZbyEgUGF5bWVudHMxFTAT
BgNVBAMTDHd3dy55by5jby51ZzEcMBoGCSqGSIb3DQEJARYNaW5mb0B5by5jby51
ZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAPPo+N67Z56ebScXJ9tX
tFpSNBNNyDlqU/X8bqouZjWuvxpWOI4xZkPKXi0t205ooVbQL/+962NASjJRrouQ
IUhJq7xhwb+KKcWyFpA25742mNgaxeZJa9iofiHeKotBvHz6pswuqa2gXAyTTmYf
j6BOIFhDeUffOjfJYbzACy7WLtbK6VIRSTHypQY+zMQluw1euyY8524GYzf8E+c5
9qjIa5YY5PPianvvR25VDNRCm0Z6GPolhIGvYPUWHFZx+HtU8xoZumi5Kddvipew
uujxNVBRyQ8bVRoYxKKuDMFHiXA6V01oPzSOtfPK7JI+rd2JFU7dQgbFxTXI9+Qx
2yUCAwEAAaOCAQIwgf8wHQYDVR0OBBYEFPj0nwwE8lJByx243yV6cfXbTKbhMIHP
BgNVHSMEgccwgcSAFPj0nwwE8lJByx243yV6cfXbTKbhoYGgpIGdMIGaMQswCQYD
VQQGEwJVRzEQMA4GA1UECBMHS2FtcGFsYTEQMA4GA1UEBxMHS2FtcGFsYTEbMBkG
A1UECgwSWW8hIFVnYW5kYSBMaW1pdGVkMRUwEwYDVQQLDAxZbyEgUGF5bWVudHMx
FTATBgNVBAMTDHd3dy55by5jby51ZzEcMBoGCSqGSIb3DQEJARYNaW5mb0B5by5j
by51Z4IJAN3e7VqDg5zQMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEFBQADggEB
AGCaUMHBxGVtVsA8xMDWknjH6hV9yuca3s0qRrOoMfM7nyOjeYtUNgZlsLxuX2n3
FhoeK9DUBvIKVSlVfO5SXgsXyWKG54YFEkZ8D50Krsyl5NCfaAJezkQ0MNdtpG98
wlD/cYa6C6DC/s1eilUbI5QqaxLo+EFy5VuHQ8tAuxJbNTVPMW9GvTjxofeMUnug
SxUMDqHmEkzbQV7yCBVqf3yi4XOM4/6B7Tr6gaandpuR+v2XaKl4SOf8G5svn96g
Kn+Bk8p6rlBWAl+5hWxHWi4dkjiLsk8q+aeKh6ibwYtRjEt/sbWTgJAZjI1mTT8d
wsLYlL7k1O3wCjUeMQzi274=
-----END CERTIFICATE-----`

/*  New
Create new api object
*/
func NewYoApi(Username, Password string) YoAPI {
	yoapi := YoAPI{
		Username:               Username,
		Password:               Password,
		YoUrl:                  "https://paymentsapi1.yo.co.ug/ybs/task.php",
		DepositTransactionType: "PULL",
		NonBlocking:            false}

	return yoapi
}

func (api *YoAPI) GetXmlResponse(xmlbody string) ([]byte, error) {
	api.LastQuery = xmlbody
	api.LastError = ""
	api.LastResponse = ""
	api.LastStatusCode = 0
	var result []byte
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	timeout := time.Duration(180 * time.Second)
	client := &http.Client{Timeout: timeout, Transport: tr}
	req, err := http.NewRequest("POST", api.YoUrl, bytes.NewBuffer([]byte(xmlbody)))
	if err != nil {
		fmt.Println(err)
		return result, err
	}
	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		api.LastError = fmt.Sprintf("do quqey: %v", err)
		fmt.Println(api.LastError)
		return result, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Printf("query response: %s", body)

	api.LastStatusCode = resp.StatusCode
	api.LastResponse = string(body)
	if resp.StatusCode != 200 {
		api.LastError = fmt.Sprintf("Wrong xml response status %d %s", resp.StatusCode, resp.Status)
		return result, fmt.Errorf(api.LastError)
	}
	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	return body, nil
}

// small helper for request body
func (api *YoAPI) createXml(method string, ext string) string {
	return fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><AutoCreate><Request><APIUsername>%s</APIUsername><APIPassword>%s</APIPassword><Method>%s</Method>%s</Request></AutoCreate>`,
		api.Username, api.Password, method, ext,
	)
}

func (api *YoAPI) xmlInternalReference() string {
	if len(api.InternalReference) > 0 {
		return "<InternalReference>" + api.InternalReference + "</InternalReference>"
	}
	return ""
}

func (api *YoAPI) xmlExternalReference() string {
	if len(api.ExternalReference) > 0 {
		return "<ExternalReference>" + api.ExternalReference + "</ExternalReference>"
	}
	return ""
}

func (api *YoAPI) xmlNonBlocking() string {
	if api.NonBlocking {
		return "<NonBlocking>TRUE</NonBlocking>"
	}
	return ""
}

func (api *YoAPI) xmlProviderReferenceText() string {
	if len(api.ProviderReferenceText) > 0 {
		return "<ProviderReferenceText>" + api.ProviderReferenceText + "</ProviderReferenceText>"
	}
	return ""
}

func (api *YoAPI) xmlInstantNotificationUrl() string {
	if len(api.InstantNotificationUrl) > 0 {
		return "<InstantNotificationUrl>" + api.InstantNotificationUrl + "</InstantNotificationUrl>"
	}
	return ""
}

func (api *YoAPI) xmlFailureNotificationUrl() string {
	if len(api.FailureNotificationUrl) > 0 {
		return "<FailureNotificationUrl>" + api.FailureNotificationUrl + "</FailureNotificationUrl>"
	}
	return ""
}

func (api *YoAPI) xmlAuthenticationSignatureBase64() string {
	if len(api.AuthenticationSignatureBase64) > 0 {
		return "<AuthenticationSignatureBase64>" + api.AuthenticationSignatureBase64 + "</AuthenticationSignatureBase64>"
	}
	return ""
}

func (api *YoAPI) xmlDepositTransactionType() string {
	return "<DepositTransactionType>" + api.DepositTransactionType + "</DepositTransactionType>"
}

func (api *YoAPI) xmlAccount(account string) string {
	return "<Account>" + account + "</Account>"
}

func (api *YoAPI) xmlAmount(amount int64) string {
	return fmt.Sprintf("<Amount>%d</Amount>", amount)
}

func (api *YoAPI) xmlNarrative(narrative string) string {
	return "<Narrative>" + narrative + "</Narrative>"
}

func (api *YoAPI) xmlCurrencyCode(currency_code string) string {
	return "<CurrencyCode>" + currency_code + "</CurrencyCode>"
}

func (api *YoAPI) xmlBeneficiaryAccount(beneficiary_account int64) string {
	return "<BeneficiaryAccount>" + fmt.Sprint(beneficiary_account) + "</BeneficiaryAccount>"
}

func (api *YoAPI) xmlBeneficiaryEmail(beneficiary_email string) string {
	return "<BeneficiaryEmail>" + beneficiary_email + "</BeneficiaryEmail>"
}

func (api *YoAPI) xmlTransactionReference(transaction_reference string) string {
	if len(transaction_reference) > 0 {
		return "<TransactionReference>" + transaction_reference + "</TransactionReference>"
	}
	return ""
}

func (api *YoAPI) xmlPrivateTransactionReference(private_transaction_reference string) string {
	if len(private_transaction_reference) > 0 {
		return "<PrivateTransactionReference>" + private_transaction_reference + "</PrivateTransactionReference>"
	}
	return ""
}

/* DepositFunds
   Request Mobile Money User to deposit funds into your account
   Shortly after you submit this request, the mobile money user receives an on-screen
   notification on their mobile phone. The notification informs the mobile money user about
   your request to transfer funds out of their account and requests them to authorize the
   request to complete the transaction.
   This request is not supported by all mobile money operator networks
   * msisdn the mobile money phone number in the format 256772123456
   * amount the amount of money to deposit into your account (floats are supported)
   * narrative the reason for the mobile money user to deposit funds

*/
func (api *YoAPI) DepositFunds(msisdn string, amount int64, narrative string) (DepositResponse, error) {
	type Resp struct {
		XMLName  xml.Name        `xml:"AutoCreate"`
		Response DepositResponse `xml:"Response"`
	}
	xmlbody := api.xmlAccount(msisdn)
	xmlbody += api.xmlAmount(amount)
	xmlbody += api.xmlNarrative(narrative)
	xmlbody += api.xmlExternalReference()
	xmlbody += api.xmlInternalReference()
	xmlbody += api.xmlProviderReferenceText()
	xmlbody += api.xmlNonBlocking()
	xmlbody += api.xmlInstantNotificationUrl()
	xmlbody += api.xmlFailureNotificationUrl()
	xmlbody += api.xmlAuthenticationSignatureBase64()
	xmlbody = api.createXml("acdepositfunds", xmlbody)
	var response DepositResponse
	var r Resp
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/* TransactionCheckStatus
Check the status of a transaction that was earlier submitted for processing.
Its particularly useful where the NonBlocking is set to TRUE.
It can also be used to check on any other transaction on the system.
* transaction_reference the response from the Yo! Payments Gateway that uniquely identifies the transaction whose status you are checking
* private_transaction_reference The External Reference that was used to carry out a transaction
*/
func (api *YoAPI) CheckTransactionStatus(transaction_reference string, private_transaction_reference string) (TransactionStatus, error) {
	type Resp struct {
		XMLName  xml.Name          `xml:"AutoCreate"`
		Response TransactionStatus `xml:"Response"`
	}
	xmlbody := api.xmlDepositTransactionType()
	xmlbody += api.xmlTransactionReference(transaction_reference)
	xmlbody += api.xmlPrivateTransactionReference(private_transaction_reference)
	xmlbody = api.createXml("actransactioncheckstatus", xmlbody)
	var response TransactionStatus
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/* InternalTransfer
   Transfer funds from your Payment Account to another Yo! Payments Account
   currency_code Options
   * "UGX-MTNMM" -> Uganda Shillings - MTN Mobile Money
   * "UGX-MTNAT" -> Uganda Shillings - MTN Airtime
   * "UGX-WTLAT" -> Uganda Shillings - Warid Airtime
   * "UGX-OULAT" -> Uganda Shillings - Orange Airtime
   * "UGX-AIRAT" -> Uganda Shillings - Airtel Airtime
   amount  The amount to be transferred
   beneficiary_account Account number of Yo! Payments User
   beneficiary_email Email Address of the recipient of funds
   narrative Textual narrative about the transaction
*/
func (api *YoAPI) InternalTransfer(currency_code string, amount int64, beneficiary_account string, beneficiary_email string, narrative string) (DepositResponse, error) {
	type Resp struct {
		XMLName  xml.Name        `xml:"AutoCreate"`
		Response DepositResponse `xml:"Response"`
	}
	xmlbody := fmt.Sprintf(
		`<CurrencyCode>%s</CurrencyCode>
    	<BeneficiaryAccount>%s</BeneficiaryAccount>
    	<BeneficiaryEmail>%s</BeneficiaryEmail>`,
		currency_code, beneficiary_account, beneficiary_email)
	xmlbody += api.xmlNarrative(narrative)
	xmlbody += api.xmlAmount(amount)
	xmlbody += api.xmlInternalReference()
	xmlbody += api.xmlExternalReference()
	xmlbody = api.createXml("acinternaltransfer", xmlbody)

	var response DepositResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/* GetAcctBalance
Get the current balance of your Yo! Payments Account
Returned array contains an array of balances (including airtime)
*/
func (api *YoAPI) GetAcctBalance() (BalanceResponse, error) {
	type Resp struct {
		XMLName  xml.Name        `xml:"AutoCreate"`
		Response BalanceResponse `xml:"Response"`
	}
	xmlbody := api.createXml("acacctbalance", "")
	var response BalanceResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/* GetMinistatement
Return an array of transactions which were carried out on your account for a certain period of time
* start_date format YYYY-MM-DD HH:MM:SS
* end_date  format YYYY-MM-DD HH:MM:SS
* transaction_status
  - "FAILED"
  - "PENDING"
  - "INDETERMINATE"
  - "SUCCEEDED"
  - "FAILED,SUCCEEDED" (comma separated)
* currency_code
  	- "UGX-MTNMM" -> Uganda Shillings - MTN Mobile Money
 	- "UGX-WARIDMM" -> Uganda Shillings - Airtel Money
  	- "UGX-MTNAT" -> Uganda Shillings - MTN Airtime
  	- "UGX-WTLAT" -> Uganda Shillings - Warid Airtime
  	- "UGX-OULAT" -> Uganda Shillings - Orange Airtime
  	- "UGX-AIRAT" -> Uganda Shillings - Airtel Airtime
* result_set_limit A value of 0 returns all. Default limit = 15
* transaction_entry_designation
 	- "TRANSACTION"
	-	"CHARGES"
 	-	"ANY"
 external_reference Filter using this external_reference
*/
func (api *YoAPI) GetMinistatement(start_date, end_date, transaction_status, currency_code, result_set_limit, transaction_entry_designation, external_reference string) (MinistatementResponse, error) {
	type Resp struct {
		XMLName  xml.Name              `xml:"AutoCreate"`
		Response MinistatementResponse `xml:"Response"`
	}
	if len(transaction_entry_designation) == 0 {
		transaction_entry_designation = "ANY"
	}
	var xmlbody string
	xmlbody += "<TransactionEntryDesignation>" + transaction_entry_designation + "</TransactionEntryDesignation>"
	if len(start_date) > 0 {
		xmlbody += "<StartDate>" + start_date + "</StartDate>"
	}
	if len(end_date) > 0 {
		xmlbody += "<EndDate>" + end_date + "</EndDate>"
	}
	if len(transaction_status) > 0 {
		xmlbody += "<TransactionStatus>" + transaction_status + "</TransactionStatus>"
	}
	if len(currency_code) > 0 {
		xmlbody += "<CurrencyCode>" + currency_code + "</CurrencyCode>"
	}
	if len(result_set_limit) > 0 {
		xmlbody += "<ResultSetLimit>" + result_set_limit + "</ResultSetLimit>"
	}
	if len(external_reference) > 0 {
		xmlbody += "<ExternalReference>" + external_reference + "</ExternalReference>"
	}

	xmlbody = api.createXml("acgetministatement", xmlbody)
	var response MinistatementResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/*  SendAirtimeMobile
Send airtime to a mobile phone user
- msisdn the mobile phone number in the format 256772123456
- amount the amount of airtime to be sent to the mobile user
- narrative textual narrative about the transfer
*/
func (api *YoAPI) SendAirtimeMobile(msisdn string, amount int64, narrative string) (DepositResponse, error) {
	type Resp struct {
		XMLName  xml.Name        `xml:"AutoCreate"`
		Response DepositResponse `xml:"Response"`
	}
	xmlbody := api.xmlAccount(msisdn)
	xmlbody += api.xmlAmount(amount)
	xmlbody += api.xmlNarrative(narrative)
	xmlbody += api.xmlNonBlocking()
	xmlbody += api.xmlExternalReference()
	xmlbody += api.xmlInternalReference()
	xmlbody += api.xmlProviderReferenceText()
	xmlbody = api.createXml("acsendairtimemobile", xmlbody)

	var response DepositResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/* SendAirtimeInternal
Send airtime from your Yo! Payments account to another Yo! Payments user account
currency_code
* "UGX-MTNAT" -> Uganda Shillings - MTN Airtime
* "UGX-WTLAT" -> Uganda Shillings - Warid Airtime
* "UGX-OULAT" -> Uganda Shillings - Orange Airtime
* "UGX-AIRAT" -> Uganda Shillings - Airtel Airtime
amount the amount of airtime to be sent to the beneficiary Yo! Payments User
beneficiary_account
beneficiary_email
narrative textual narrative about the transfer
*/
func (api *YoAPI) SendAirtimeInternal(currency_code string, amount int64, beneficiary_account int64, beneficiary_email string, narrative string) (DepositResponse, error) {
	type Resp struct {
		XMLName  xml.Name        `xml:"AutoCreate"`
		Response DepositResponse `xml:"Response"`
	}
	xmlbody := api.xmlAmount(amount)
	xmlbody += api.xmlNarrative(narrative)
	xmlbody += api.xmlInternalReference()
	xmlbody += api.xmlCurrencyCode(currency_code)
	xmlbody += api.xmlBeneficiaryAccount(beneficiary_account)
	xmlbody += api.xmlBeneficiaryEmail(beneficiary_email)
	xmlbody += api.xmlInternalReference()
	xmlbody += api.xmlExternalReference()
	xmlbody = api.createXml("acsendairtimeinternal", xmlbody)

	var response DepositResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/* VerifyAccountValidity
Verify the validity of a given mobile money account
 msisdn the mobile phone number in the format 256772123456
 return boolean true if valid
*/
func (api *YoAPI) VerifyAccountValidity(msisdn string) (bool, error) {
	type Resp struct {
		XMLName  xml.Name              `xml:"AutoCreate"`
		Response VerifyAccountResponse `xml:"Response"`
	}
	xmlbody := api.xmlAccount(msisdn)
	xmlbody = api.createXml("acverifyaccountvalidity", xmlbody)

	var isvalid bool = false
	var response VerifyAccountResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return isvalid, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	if response.Response.Status == "OK" {
		if response.Response.Valid == "TRUE" {
			isvalid = true
		} else {
			isvalid = false
		}
	}
	return isvalid, err
}

/* WithdrawFunds
   Withdraw funds from your YO! Payments Account to a mobile money user
   This transaction transfers funds from your YO! Payments Account to a mobile money user.
   Please handle this request with care because if compromised, it can lead to
   withdrawal of funds from your account.
   This request is not supported by all mobile money operator networks
   This request requires permission that is granted by the issuance of an API Access Letter
   * msisdn the mobile money phone number in the format 256772123456
   * amount the amount of money to withdraw from your account (floats are supported)
   * narrative the reason for withdrawal of funds from your account
*/
func (api *YoAPI) WithdrawFunds(msisdn string, amount int64, narrative string) (DepositResponse, error) {
	type Resp struct {
		XMLName  xml.Name        `xml:"AutoCreate"`
		Response DepositResponse `xml:"Response"`
	}
	xmlbody := api.xmlNonBlocking()
	xmlbody += api.xmlAccount(msisdn)
	xmlbody += api.xmlAmount(amount)
	xmlbody += api.xmlNarrative(narrative)
	xmlbody += api.xmlExternalReference()
	xmlbody += api.xmlInternalReference()
	xmlbody += api.xmlProviderReferenceText()
	xmlbody = api.createXml("acwithdrawfunds", xmlbody)

	var response DepositResponse
	resp, err := api.GetXmlResponse(xmlbody)
	if err != nil {
		return response, err
	}
	var r Resp
	err = xml.Unmarshal(resp, &r)
	response = r.Response
	return response, err
}

/*
 */
func (api *YoAPI) ReceivePaymentNotification(date_time, amount, narrative, network_ref, external_ref, msisdn, signature string) (PaymentNotificationResponse, error) {
	var result PaymentNotificationResponse
	verified, err := api.verifyPaymentNotification(date_time, amount, narrative, network_ref, external_ref, msisdn, signature)
	result.Verified = verified
	result.DateTime = date_time
	result.Amount = amount
	result.Narrative = narrative
	result.NetworkRef = network_ref
	result.ExternalRef = external_ref
	result.Msisdn = msisdn
	return result, err
}

func (api *YoAPI) ReceivePaymentFailureNotification(failed_transaction_reference, transaction_init_date, verification string) (result PaymentFailureNotificationResponse, err error) {
	result.Verified, err = api.verifyPaymentFailureNotification(failed_transaction_reference, transaction_init_date, verification)
	result.FailedTransactionReference = failed_transaction_reference
	result.TransactionInitDate = transaction_init_date
	return
}

func (api *YoAPI) verifyPaymentNotification(date_time, amount, narrative, network_ref, external_ref, msisdn, signature string) (bool, error) {
	data := date_time + amount + narrative + network_ref + external_ref + msisdn
	sign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	block, _ := pem.Decode([]byte(pub_crt))
	if err != nil {
		return false, err
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	pubkey := cert.PublicKey.(*rsa.PublicKey)

	hash := sha1.New()
	hash.Write([]byte(data))
	digest := hash.Sum(nil)
	err = rsa.VerifyPKCS1v15(pubkey, crypto.SHA1, digest, sign)
	if err != nil {
		return false, err
	}
	return true, nil
	/*
	   $key_id = openssl_pkey_get_public($key);
	   $verified = openssl_verify($data, $signature, $key_id);
	   openssl_free_key($key_id);
	   if($verified == 1){
	       return true;
	   }
	   return false;
	*/
}

func (api *YoAPI) verifyPaymentFailureNotification(failed_transaction_reference, transaction_init_date, verification string) (bool, error) {
	data := failed_transaction_reference + transaction_init_date
	sign, err := base64.StdEncoding.DecodeString(verification)
	if err != nil {
		return false, err
	}

	block, _ := pem.Decode([]byte(pub_crt))
	if err != nil {
		return false, err
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	pubkey := cert.PublicKey.(*rsa.PublicKey)

	hash := sha1.New()
	hash.Write([]byte(data))
	digest := hash.Sum(nil)
	err = rsa.VerifyPKCS1v15(pubkey, crypto.SHA1, digest, sign)
	if err != nil {
		return false, err
	}
	return true, nil
}
