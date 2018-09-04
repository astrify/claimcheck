package controllers

import (
	"bytes"
	cryptoSha256 "crypto/sha256"
	"encoding/hex"
	"github.com/labstack/echo"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/xdr"
	"net/http"
)

// Define the input
type VerifyPostInput struct {
	AssetIssuer   	string `json:"asset_issuer"`
	AssetCode   	string `json:"asset_code"`
	Secret    		string `json:"secret"`
	TransactionHash string `json:"transaction_hash"`
}

func VerifyPost(c echo.Context) error {

	i := new(VerifyPostInput)
	if err := c.Bind(i); err != nil {
		return err
	}

	// Load the transaction
	t, err := loadTransaction(i.TransactionHash)
	if err != nil {
		return err
	}

	r := verify(i, t.EnvelopeXdr)
	return c.JSON(http.StatusOK, r)
}

func loadTransaction(t string) (transaction hProtocol.Transaction, err error) {
	transaction, err = horizon.DefaultTestNetClient.LoadTransaction(t)
	if err != nil {
		return transaction, err
	}

	return transaction, nil
}

func verify(i *VerifyPostInput, envelopeXdr string) bool {

	// unmarshal envelopeXdr into TransactionEnvelope
	var tx xdr.TransactionEnvelope
	err := tx.Scan(envelopeXdr)
	if err != nil {
		return false
	}

	// verify given issuer is a valid keypair
	issuerKeypair, err := keypair.Parse(i.AssetIssuer)
	if err != nil {
		return false
	}

	// verify the given transaction contains a single operation
	if len(tx.Tx.Operations) != 1 {
		return false
	}

	// verify the given transaction contains a payment operation
	if tx.Tx.Operations[0].Body.PaymentOp == nil {
		return false
	}

	// verify the transaction sent the asset back its the issuer
	payment := tx.Tx.Operations[0].Body.PaymentOp
	if !payment.Destination.Equals(payment.Asset.AlphaNum12.Issuer) {
		return false
	}

	// ensure the asset issuer matches the input
	if payment.Asset.AlphaNum12.Issuer.Address() != issuerKeypair.Address() {
		return false
	}

	// ensure the asset code matches the input
	var AssetCodeBytes [12]byte
	copy(AssetCodeBytes[:], i.AssetCode)
	if payment.Asset.AlphaNum12.AssetCode != AssetCodeBytes {
		return false
	}

	// verify the given secret hashes to the memo hash
	if!checkSecret(i.Secret, tx.Tx.Memo) {
		return false
	}

	return true
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
