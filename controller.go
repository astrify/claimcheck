package verify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	// "encoding/base64"
	"github.com/labstack/echo"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/xdr"
	"net/http"
	// "strings"
)

// Define the input
type input struct {
	AssetIssuer   	string `json:"asset_issuer"`
	AssetCode   	string `json:"asset_code"`
	Secret    		string `json:"secret"`
	TransactionHash string `json:"transaction_hash"`
}

func Verify(c echo.Context) error {

	i := new(input)
	if err := c.Bind(i); err != nil {
		return err
	}

	//
	issuerKeypair, err := keypair.Parse(i.AssetIssuer)
	if err != nil {
		return err
	}
	fmt.Printf("%T\n", issuerKeypair)

	// Load the transaction
	t, err := LoadTransaction(i.TransactionHash)
	if err != nil {
		return err
	}

	// Unmarshal into TransactionEnvelope
	var tx xdr.TransactionEnvelope
	tx.Scan(t.EnvelopeXdr)

	secretMatch := checkSecret(i, tx)

	op := tx.Tx.Operations[0]
	fmt.Printf("%T\n", op)
	if op.Body.PaymentOp == nil {
		return err
	}

	// ensure the transaction was sending the asset back to the issuer
	payment := op.Body.PaymentOp
	redemption := false
	if payment.Destination.Equals(payment.Asset.AlphaNum12.Issuer) {
		redemption = true
		fmt.Printf("%T\n", redemption)
	}

	// ensure the issuer matches the input
	issuerMatch := false
	if payment.Asset.AlphaNum12.Issuer.Address() == issuerKeypair.Address() {
		issuerMatch = true
		fmt.Printf("%T\n", issuerMatch)
	}

	// ensure the asset code matches the input
	codeMatch := false
	var AssetCodeBytes [12]byte
	copy(AssetCodeBytes[:], i.AssetCode)
	if payment.Asset.AlphaNum12.AssetCode == AssetCodeBytes {
		codeMatch = true
		fmt.Printf("%T\n", codeMatch)
	}

	result := secretMatch && redemption && issuerMatch && codeMatch

	return c.JSON(http.StatusOK, result)
}

func LoadTransaction(t string) (transaction hProtocol.Transaction, err error) {
	transaction, err = horizon.DefaultTestNetClient.LoadTransaction(t)
	if err != nil {
		return transaction, err
	}

	return transaction, nil
}

func checkSecret(i *input, tx xdr.TransactionEnvelope) (bool) {
	secretBytes, err := hex.DecodeString(i.Secret)
	if err != nil {
		panic(err)
	}

	publicBytes := getMemoBytes(tx)
	proofBytes := getSha256(secretBytes)
	return bytes.Equal(publicBytes, proofBytes)
}

func getMemoBytes(tx xdr.TransactionEnvelope) ([]byte) {
	publicHash := tx.Tx.Memo.Hash
	publicBytes := publicHash[:] // https://stackoverflow.com/a/27762870
	return publicBytes
}

func getSha256(in []byte) ([]byte) {
	hasher := sha256.New()
	hasher.Write(in)
	return hasher.Sum(nil)
}
