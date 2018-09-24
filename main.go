package main

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_echo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/go-playground/validator.v9"
	"claimcheck/controllers"
	"net/http"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {

	// Echo instance
	e := echo.New()
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit("1K"))

	// Validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Create a limiter struct.
	limiter := tollbooth.NewLimiter(1, nil)

	// Key Routes
	e.POST("/", controllers.ClaimCheck, tollbooth_echo.LimitHandler(limiter))
	e.GET("/", func(c echo.Context) error { return c.String(http.StatusOK, "ready") })


	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}