package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type CustomValidator struct {
validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func testGetInput1() *input {
	i := &input{
		AssetIssuer: "GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",
		AssetCode: "WOOLLY",
		Secret: "EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",
		TransactionHash: "54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",
	}

	return i
}

func testGetOutput1() *output {
	o := &output{
		AssetIssuer: "GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",
		AssetCode: "WOOLLY",
		TransactionHash: "54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",
		Valid: true,
		Amount: "10000000",
	}

	return o
}

func testGetTxXdr1() string {
	return "AAAAAJApr4VjRhcrExTJNWuv6LeK8h4Hk5LMBRfRIhGvfe+FAAAAZABumd0AAAAPAAAAAAAAAANBeV8g6mqV315SDVhZ/G8mxRQLMdkTtRI94Lp39DkINAAAAAEAAAAAAAAAAQAAAAAh4lGr55XA4P1v8G7V7i+cPpZyf1mz4MshoekvA6fcHgAAAAJXT09MTFkAAAAAAAAAAAAAIeJRq+eVwOD9b/Bu1e4vnD6Wcn9Zs+DLIaHpLwOn3B4AAAAAAJiWgAAAAAAAAAABr33vhQAAAECTODuhctoTJ0CMUxInyp6lwedWvpsQrEpt5kqQTK1zO3ZtOlXtEhJMnbtBlZN7Lkx+NhCDKgVh5fZlyp/KU2wD"
}

func testGetTx1() xdr.TransactionEnvelope {
	var tx xdr.TransactionEnvelope
	_ = tx.Scan(testGetTxXdr1())
	return tx
}

// full controller integration test
func TestClaimCheck(t *testing.T) {
	// setup request body
	i := testGetInput1();
	inputBytes, _ := json.Marshal(i)
	inputJson := string(inputBytes)

	// setup expected response body
	o := testGetOutput1();
	outputBytes, _ := json.Marshal(o)
	outputJson := string(outputBytes)

	// setup echo request
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader(inputJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// assertions
	if assert.NoError(t, ClaimCheck(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, outputJson, rec.Body.String())
	}
}

func TestGetInputIssuerAccount(t *testing.T) {
	i := testGetInput1()
	var expected xdr.AccountId
	expected.SetAddress(i.AssetIssuer)

	a, err := getInputIssuerAccount(i.AssetIssuer)
	assert.Equal(t, expected, a)
	fmt.Print(err)
}

func TestGetInputAsset(t *testing.T) {
	i := testGetInput1()
	accountId, _ := getInputIssuerAccount(i.AssetIssuer)
	expected, _ := tempSetCredit(i.AssetCode, accountId)

	a, err := getInputAsset(i.AssetCode,accountId)
	assert.Equal(t, expected, a)
	fmt.Print(err)
}

func TestLoadAndValidateTransaction(t *testing.T) {
	i := testGetInput1()
	txXdr := testGetTxXdr1()
	tx := testGetTx1()

	trans, err := loadTransaction(i.TransactionHash)
	assert.Equal(t, txXdr, trans.EnvelopeXdr)
	fmt.Print(err)

	assert.NoError(t, validateTransaction(trans, tx))
}

func TestGetTransactionEnvelope(t *testing.T) {
	txXdr := testGetTxXdr1()
	expected := testGetTx1()

	tx, err := getTransactionEnvelope(txXdr)
	assert.Equal(t, expected, tx)
	fmt.Print(err)
}

func TestGetPaymentOperation(t *testing.T) {
	tx := testGetTx1()
	expected := tx.Tx.Operations[0].Body.PaymentOp

	op, err := getPaymentOperation(tx)
	fmt.Print(err)
	assert.Equal(t, expected, op)
}

func TestValidateTransactionAsset(t *testing.T) {
	tx := testGetTx1()
	op, _ := getPaymentOperation(tx)

	assert.NoError(t, validateTransactionAsset(op))
}

func TestValidateTransactionIsRedemption(t *testing.T) {
	tx := testGetTx1()
	op, _ := getPaymentOperation(tx)

	assert.NoError(t, validateTransactionIsRedemption(op))
}

func TestValidateInputAssetEqualsTransactionAsset(t *testing.T) {
	i := testGetInput1()
	tx := testGetTx1()
	op, _ := getPaymentOperation(tx)
	accountId, _ := getInputIssuerAccount(i.AssetIssuer)
	inputAsset, _ := getInputAsset(i.AssetCode, accountId)

	assert.NoError(t, validateInputAssetEqualsTransactionAsset(op, inputAsset))
}

func TestGetTransactionAmountAsString(t *testing.T) {
	tx := testGetTx1()
	op, _ := getPaymentOperation(tx)
	amount := getTransactionAmountAsString(op)
	assert.Equal(t, "10000000", amount)
}

func TestValidateSecret(t *testing.T) {
	i := testGetInput1()
	tx := testGetTx1()
	assert.NoError(t, validateSecret(i.Secret, tx.Tx.Memo))
}
