package verify

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	// Echo instance
	e := echo.New()
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Key Routes
	e.POST("/", Verify)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}