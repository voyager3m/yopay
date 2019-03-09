package yopay

import (
	"fmt"
	"testing"
	"time"
)

func newTestingApi(t *testing.T) *YoAPI {
	result := NewYoApi("90003851865", "1168170290")
	result.YoUrl = "https://sandbox.yo.co.ug/services/yopaymentsdev/task.php" //< sandbox url
	return &result
}

func TestGetAcctBalance(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.GetAcctBalance()
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
	t.Logf("%+v", r)
}

var transaction_id string

func TestDepositFunds(t *testing.T) {
	yo := newTestingApi(t)
	//yo.NonBlocking = true
	r, err := yo.DepositFunds("256771234567", 2000, fmt.Sprintf("%d", time.Now().Unix()))

	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
	transaction_id = r.TransactionReference
	t.Log(transaction_id)
}

func TestCheckTransactionStatus(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.CheckTransactionStatus(transaction_id, "")
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)

}

func qTestInternalTransfer(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.InternalTransfer("111", 111, "222", "333", "444")
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
}

func TestGetMinistatement(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.GetMinistatement("2019-01-01 00:00:00", "2019-02-26 12:00:00", "", "", "", "", "")
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
	t.Log(r.Transactions)
}

func qTestSendAirtimeMobile(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.SendAirtimeMobile("111", 111, "222")
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
}

func qTestSendAirtimeInternal(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.SendAirtimeInternal("11", 22, 333, "44", "55")
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
}

func TestVerifyAccountValidity(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.VerifyAccountValidity("256771234567")
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r)
}

func TestWithdrawFunds(t *testing.T) {
	yo := newTestingApi(t)
	r, err := yo.WithdrawFunds("256771234567", 100, fmt.Sprintf("test withdarw %d", time.Now().Unix()))
	if err != nil {
		t.Fatalf("server return error: %v", err)
	}
	t.Log(r.Status)
}
