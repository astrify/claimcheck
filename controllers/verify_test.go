package controllers

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

var (
	inputJSON = `{` +
		`"asset_issuer":"GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",` +
		`"asset_code":"WOOLLY",` +
		`"secret":"EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",` +
		`"transaction_hash":"54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910"` +
	`}`

	envelopeXdr = "AAAAAJApr4VjRhcrExTJNWuv6LeK8h4Hk5LMBRfRIhGvfe+FAAAAZABumd0AAAAPAAAAAAAAAANBeV8g6mqV315SDVhZ/G8mxRQLMdkTtRI94Lp39DkINAAAAAEAAAAAAAAAAQAAAAAh4lGr55XA4P1v8G7V7i+cPpZyf1mz4MshoekvA6fcHgAAAAJXT09MTFkAAAAAAAAAAAAAIeJRq+eVwOD9b/Bu1e4vnD6Wcn9Zs+DLIaHpLwOn3B4AAAAAAJiWgAAAAAAAAAABr33vhQAAAECTODuhctoTJ0CMUxInyp6lwedWvpsQrEpt5kqQTK1zO3ZtOlXtEhJMnbtBlZN7Lkx+NhCDKgVh5fZlyp/KU2wD"
)

func TestVerifyPost(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, VerifyPost(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Body.String())
	}
}

func TestVerifyPasses(t *testing.T) {

	i := VerifyPostInput{
		AssetIssuer: "GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",
		AssetCode: "WOOLLY",
		Secret: "EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",
		TransactionHash: "54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",
	}


	r := verify(&i,envelopeXdr)
	assert.Equal(t, true, r)
}

func TestVerifyFails(t *testing.T) {

	i := VerifyPostInput{
		AssetIssuer: "GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",
		AssetCode: "WOOLLY",
		Secret: "EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",
		TransactionHash: "54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910",
	}


	r := verify(&i,envelopeXdr)
	assert.Equal(t, true, r)
}