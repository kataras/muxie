package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	r := echo.New()

	r.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome!\n")
	})

	r.GET("/user/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, c.Param("id"))
	})

	r.Start(":3000")
}
