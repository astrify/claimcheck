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
)

// Define the input
type verifyPostInput struct {
	AssetIssuer   	string `json:"asset_issuer" validate:"required,base64,len=56"`
	AssetCode   	string `json:"asset_code" validate:"required,min=1,max=12"`
	Secret    		string `json:"secret" validate:"required,hexadecimal,len=64"`
	TransactionHash string `json:"transaction_hash" validate:"required,hexadecimal,len=64"`
}

type verifyPostOutput struct{
	AssetIssuer   	string  	`json:"asset_issuer"`
	AssetCode   	string  	`json:"asset_code"`
	Secret    		string  	`json:"secret"`
	TransactionHash string  	`json:"transaction_hash"`
	Verified		bool    	`json:"verified"`
	Amount			xdr.Int64   `json:"amount"`
}

func VerifyPost(c echo.Context) error {
	i := new(verifyPostInput)

	if err := c.Bind(i); err != nil {
		return err
	}

	if err := c.Validate(i); err != nil {
		return err
	}

	// Load the transaction
	t, err := loadTransaction(i.TransactionHash)
	if err != nil {
		return err
	}

	// verify the given transaction contains a memo hash
	if t.MemoType != "hash" {
		return errors.New("the given transaction does not contain a memo of type 'hash'")
	}

	// verify the given transaction contains a single operation
	if t.OperationCount != 1 {
		return errors.New("the transaction contains more then one operation")
	}

	// verify the given transaction was successful
	var rx xdr.TransactionResult
	err = rx.Scan(t.ResultXdr)
	if err != nil {
		return err
	}
	if rx.Result.Code != xdr.TransactionResultCodeTxSuccess {
		return errors.New("the given transaction was not successful")
	}

	o, err := getOutput(i, t.EnvelopeXdr)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

func loadTransaction(t string) (transaction hProtocol.Transaction, err error) {
	transaction, err = horizon.DefaultTestNetClient.LoadTransaction(t)
	if err != nil {
		return transaction, err
	}

	return transaction, nil
}

func getOutput(i *verifyPostInput, envelopeXdr string) (*verifyPostOutput, error) {
	o := &verifyPostOutput{
		AssetIssuer: i.AssetIssuer,
		AssetCode: i.AssetCode,
		Secret: i.Secret,
		TransactionHash: i.TransactionHash,
		Verified: false,
		Amount: 0,
	}

	// unmarshal envelopeXdr into TransactionEnvelope
	var tx xdr.TransactionEnvelope
	err := tx.Scan(envelopeXdr)
	if err != nil {
		return o, err
	}

	o.Verified, err = verify(i, tx)
	if err != nil {
		return o, err
	}

	o.Amount = tx.Tx.Operations[0].Body.PaymentOp.Amount
	return o, nil
}

func verify(i *verifyPostInput, tx xdr.TransactionEnvelope) (bool, error) {

	// verify given issuer is a valid keypair
	issuerKeypair, err := keypair.Parse(i.AssetIssuer)
	if err != nil {
		return false, errors.New("the given issuer public key is invalid")
	}

	// verify the given transaction contains a payment operation
	if tx.Tx.Operations[0].Body.PaymentOp == nil {
		return false, errors.New("the transaction does not contain a payment operation")
	}

	// verify the transaction sent the asset back its the issuer
	payment := tx.Tx.Operations[0].Body.PaymentOp
	if !payment.Destination.Equals(payment.Asset.AlphaNum12.Issuer) {
		return false, errors.New("the payment operation did not return the asset its issuer")
	}

	// ensure the asset issuer matches the input
	if payment.Asset.AlphaNum12.Issuer.Address() != issuerKeypair.Address() {
		return false, errors.New("the payment operation's asset issuer does not match the given issuer")
	}

	// ensure the asset code matches the input
	var AssetCodeBytes [12]byte
	copy(AssetCodeBytes[:], i.AssetCode)
	if payment.Asset.AlphaNum12.AssetCode != AssetCodeBytes {
		return false, errors.New("the payment operation's asset code does not match the given code")
	}

	// verify the given secret hashes to the memo hash
	if!checkSecret(i.Secret, tx.Tx.Memo) {
		return false, errors.New("the transaction's memo is not an sha256 hash of the given secret")
	}

	return true, nil
}

func checkSecret(secretString string, memo xdr.Memo) (bool) {
	secretBytes, err := hex.DecodeString(secretString)
	if err != nil {
		return false
	}

	publicBytes := memo.Hash[:] // https://stackoverflow.com/a/27762870
	return bytes.Equal(publicBytes, sha256(secretBytes))
}

func sha256(in []byte) ([]byte) {
	hasher := cryptoSha256.New()
	hasher.Write(in)
	return hasher.Sum(nil)
}
