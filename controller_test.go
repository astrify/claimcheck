package verify

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

var (
	verifyJSON = `{` +
		`"asset_issuer":"GAQ6EUNL46K4BYH5N7YG5VPOF6OD5FTSP5M3HYGLEGQ6SLYDU7OB4M45",` +
		`"asset_code":"WOOLLY",` +
		`"secret":"EC9CB7E10C0BDE144FFF36128EEE7B5B59E49A1BC7710FDAB583735880DB0ECA",` +
		`"transaction_hash":"54913d4c6069240f983b8609160b0d5876c57e0e864eeb93f75872f3a798c910"` +
	`}`)

func TestVerify(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader(verifyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, Verify(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Body.String())
	}
}
