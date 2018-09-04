package verify

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"verify/controllers"
)

func main() {

	// Echo instance
	e := echo.New()
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Key Routes
	e.POST("/", controllers.VerifyPost)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}