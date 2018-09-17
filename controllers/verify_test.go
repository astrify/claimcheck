package controllers

import (
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

func TestVerifyPost(t *testing.T) {
	var inputJSON =
		`{` +
			`"asset_issuer":"GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",` +
			`"asset_code":"WOOLLY",` +
			`"secret":"EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",` +
			`"transaction_hash":"54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910"` +
		`}`

	var outputJson =
		`{` +
			`"asset_issuer":"GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",` +
			`"asset_code":"WOOLLY",` +
			`"transaction_hash":"54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",` +
			`"verified":true,` +
			`"amount":"10000000"` +
		`}`

	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, VerifyPost(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, outputJson, rec.Body.String())
	}
}

func TestVerifyPasses(t *testing.T) {

	i := verifyPostInput{
		AssetIssuer: "GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",
		AssetCode: "WOOLLY",
		Secret: "EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",
		TransactionHash: "54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",
	}

	var envelopeXdr = "AAAAAJApr4VjRhcrExTJNWuv6LeK8h4Hk5LMBRfRIhGvfe+FAAAAZABumd0AAAAPAAAAAAAAAANBeV8g6mqV315SDVhZ/G8mxRQLMdkTtRI94Lp39DkINAAAAAEAAAAAAAAAAQAAAAAh4lGr55XA4P1v8G7V7i+cPpZyf1mz4MshoekvA6fcHgAAAAJXT09MTFkAAAAAAAAAAAAAIeJRq+eVwOD9b/Bu1e4vnD6Wcn9Zs+DLIaHpLwOn3B4AAAAAAJiWgAAAAAAAAAABr33vhQAAAECTODuhctoTJ0CMUxInyp6lwedWvpsQrEpt5kqQTK1zO3ZtOlXtEhJMnbtBlZN7Lkx+NhCDKgVh5fZlyp/KU2wD"
	var tx xdr.TransactionEnvelope
	tx.Scan(envelopeXdr)


	r, err := verify(&i, tx)
	assert.Equal(t, true, r)
	fmt.Print(err)
}

func TestVerifyFails(t *testing.T) {

	i := verifyPostInput{
		AssetIssuer: "GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M48",
		AssetCode: "WOOLLY",
		Secret: "EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",
		TransactionHash: "54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",
	}

	var envelopeXdr = "AAAAAJApr4VjRhcrExTJNWuv6LeK8h4Hk5LMBRfRIhGvfe+FAAAAZABumd0AAAAPAAAAAAAAAANBeV8g6mqV315SDVhZ/G8mxRQLMdkTtRI94Lp39DkINAAAAAEAAAAAAAAAAQAAAAAh4lGr55XA4P1v8G7V7i+cPpZyf1mz4MshoekvA6fcHgAAAAJXT09MTFkAAAAAAAAAAAAAIeJRq+eVwOD9b/Bu1e4vnD6Wcn9Zs+DLIaHpLwOn3B4AAAAAAJiWgAAAAAAAAAABr33vhQAAAECTODuhctoTJ0CMUxInyp6lwedWvpsQrEpt5kqQTK1zO3ZtOlXtEhJMnbtBlZN7Lkx+NhCDKgVh5fZlyp/KU2wD"
	var tx xdr.TransactionEnvelope
	tx.Scan(envelopeXdr)


	r, err := verify(&i, tx)
	assert.Equal(t, false, r)
	fmt.Print(err)
}