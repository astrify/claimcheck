package controllers

import (
	"bytes"
	cryptoSha256 "crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/labstack/echo"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/xdr"
	"net/http"
	"strconv"
)

type input struct {
	AssetIssuer   	string `json:"asset_issuer" validate:"required,base64,len=56"`
	AssetCode   	string `json:"asset_code" validate:"required,min=1,max=12"`
	Secret    		string `json:"secret" validate:"required,hexadecimal,len=64"`
	TransactionHash string `json:"transaction_hash" validate:"required,hexadecimal,len=64"`
}

type output struct{
	AssetIssuer   	string  	`json:"asset_issuer"`
	AssetCode   	string  	`json:"asset_code"`
	TransactionHash string  	`json:"transaction_hash"`
	Valid			bool    	`json:"verified"`
	Amount			string  	`json:"amount"`
}


func ClaimCheck(c echo.Context) error {

	i, err := bindAndValidateInput(c)
	if err != nil {
		return err
	}

	inputIssuerAccount, err := getInputIssuerAccount(i.AssetIssuer)
	if err != nil {
		return err
	}

	inputAsset, err := getInputAsset(i.AssetCode, inputIssuerAccount)
	if err != nil {
		return err
	}

	t, err := loadTransaction(i.TransactionHash)
	if err != nil {
		return err
	}

	tx, err := getTransactionEnvelope(t.EnvelopeXdr)
	if err != nil {
		return err
	}

	err = validateTransaction(t, tx)
	if err != nil {
		return err
	}

	op, err := getPaymentOperation(tx)
	if err != nil {
		return err
	}

	err = validateTransactionAsset(op)
	if err != nil {
		return err
	}

	err = validateTransactionIsRedemption(op)
	if err != nil {
		return err
	}

	err = validateInputAssetEqualsTransactionAsset(op, inputAsset)
	if err != nil {
		return err
	}

	err = validateSecret(i.Secret, tx.Tx.Memo)
	if err != nil {
		return err
	}

	a := getTransactionAmountAsString(op)
	o := &output{
		AssetIssuer: i.AssetIssuer,
		AssetCode: i.AssetCode,
		TransactionHash: i.TransactionHash,
		Valid: true,
		Amount: a,
	}

	return c.JSON(http.StatusOK, o)
}

func bindAndValidateInput(c echo.Context) (*input, error){
	i := new(input)

	if err := c.Bind(i); err != nil {
		return i, err
	}

	if err := c.Validate(i); err != nil {
		return i, err
	}

	return i, nil
}

func getInputIssuerAccount(i string) (xdr.AccountId, error){
	var a xdr.AccountId

	_, err := keypair.Parse(i)
	if err != nil {
		return a, echo.NewHTTPError(http.StatusUnprocessableEntity, "the given issuer public key is invalid")
	}

	a.SetAddress(i)
	if err != nil {
		return a, echo.NewHTTPError(http.StatusUnprocessableEntity, "the given issuer public key is invalid")
	}

	return a, nil
}

func getInputAsset(code string, issuer xdr.AccountId) (xdr.Asset, error) {
	a, err := tempSetCredit(code, issuer)
	if err != nil {
		return a, echo.NewHTTPError(http.StatusUnprocessableEntity, "the given asset code is invalid")
	}

	return a, nil
}

func loadTransaction(t string) (hProtocol.Transaction, error) {
	transaction, err := horizon.DefaultTestNetClient.LoadTransaction(t)
	if err != nil {
		return transaction, err
	}

	return transaction, nil
}

func getTransactionEnvelope(envelopeXdr string) (xdr.TransactionEnvelope, error) {
	var tx xdr.TransactionEnvelope
	err := tx.Scan(envelopeXdr)
	if err != nil {
		return tx, echo.NewHTTPError(http.StatusInternalServerError, "error decoding TransactionEnvelopeXdr")
	}

	return tx, nil
}

func validateTransaction(t hProtocol.Transaction, tx xdr.TransactionEnvelope) (error){

	if t.MemoType != "hash" {
		return echo.NewHTTPError(http.StatusUnprocessableEntity,"the given transaction does not contain a memo of type 'hash'")
	}

	if t.OperationCount != 1 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "the given transaction contains more then one operation")
	}

	var rx xdr.TransactionResult
	err := rx.Scan(t.ResultXdr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error decoding TransactionResultXdr")
	}
	if rx.Result.Code != xdr.TransactionResultCodeTxSuccess {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "the given transaction was not a successful operation")
	}

	return nil
}

func getPaymentOperation(tx xdr.TransactionEnvelope) (*xdr.PaymentOp, error){

	op := tx.Tx.Operations[0].Body.PaymentOp;
	if op == nil {
		return op, echo.NewHTTPError(http.StatusUnprocessableEntity, "the given transaction does not contain a payment operation")
	}

	return op, nil
}

func validateTransactionAsset(op *xdr.PaymentOp) (error) {
	if op.Asset.Type == xdr.AssetTypeAssetTypeNative {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "the given transaction contains a native payment operation")
	}

	return nil
}

// verify the transaction was a redemption whereby the asset was sent back to it's issuer
func validateTransactionIsRedemption(op *xdr.PaymentOp) error {
	asset := op.Asset
	var typ xdr.AssetType
	var code, issuer string

	err := asset.Extract(&typ, &code, &issuer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error extracting transaction asset")
	}

	var issuerAccount xdr.AccountId
	issuerAccount.SetAddress(issuer)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "the transaction issuer public key is invalid")
	}

	if !op.Destination.Equals(issuerAccount) {
		return errors.New("the payment operation in the given transaction did not return the asset to its issuer")
	}

	return nil
}

func validateInputAssetEqualsTransactionAsset(op *xdr.PaymentOp, inAsset xdr.Asset) (error) {
	if op.Asset.Equals(inAsset) {
		return nil
	}

	return echo.NewHTTPError(http.StatusUnprocessableEntity, "the given asset_code and asset_issuer do not match the asset contained in the given transaction")
}

func getTransactionAmountAsString(op *xdr.PaymentOp) (string) {
	return strconv.FormatInt(int64(op.Amount), 10)
}

func validateSecret(secretString string, memo xdr.Memo) error {
	secretBytes, err := hex.DecodeString(secretString)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error decoding string secret")
	}

	publicBytes := memo.Hash[:] // https://stackoverflow.com/a/27762870

	if(bytes.Equal(publicBytes, sha256(secretBytes))){
		return nil
	}

	return errors.New("the transaction's memo is not an sha256 hash of the given secret")
}

func sha256(in []byte) ([]byte) {
	hasher := cryptoSha256.New()
	hasher.Write(in)
	return hasher.Sum(nil)
}

// Temp function until https://github.com/stellar/go/issues/682 is fixed
func tempSetCredit(code string, issuer xdr.AccountId) (xdr.Asset, error) {
	length := len(code)
	var a xdr.Asset
	var typ xdr.AssetType
	var body interface{}


	switch {
	case length >= 1 && length <= 4:
		newbody := xdr.AssetAlphaNum4{Issuer: issuer}
		copy(newbody.AssetCode[:], []byte(code)[:length])
		typ = xdr.AssetTypeAssetTypeCreditAlphanum4
		body = newbody
	case length >= 5 && length <= 12:
		newbody := xdr.AssetAlphaNum12{Issuer: issuer}
		copy(newbody.AssetCode[:], []byte(code)[:length])
		typ = xdr.AssetTypeAssetTypeCreditAlphanum12
		body = newbody
	default:
		return a, errors.New("Asset code length is invalid")
	}

	newa, err := xdr.NewAsset(typ, body)
	if err != nil {
		return a, err
	}

	return newa, nil
}