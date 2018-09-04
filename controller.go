package verify

import (
	"github.com/labstack/echo"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"net/http"
)

func Verify(c echo.Context) error {

	// Define the input
	type input struct {
		AssetIssuer   	string `json:"asset_issuer"`
		AssetCode   	string `json:"asset_code"`
		Secret    		string `json:"secret"`
		TransactionHash string `json:"transaction_hash"`
	}

	i := new(input)
	if err := c.Bind(i); err != nil {
		return err
	}

	//
	_, err := keypair.Parse(i.AssetIssuer)
	if err != nil {
		return err
	}

	_, err = horizon.DefaultTestNetClient.LoadTransaction(i.TransactionHash)
	if err != nil {
		return err
	}

	// Define the output
	type output struct{
		SignedTransaction string  `json:"signed_transaction"`
	}

	// Bind the output
/*	r := &output{
		SignedTransaction:  "Ok",
	}*/

	return c.JSON(http.StatusOK, i)
}