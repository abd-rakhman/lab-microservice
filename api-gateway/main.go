package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

type Range struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}

type One struct {
	Index int `json:"index"`
}

func main() {
	e := echo.New()

	e.GET("/getList", func(c echo.Context) error {
		_, err := http.Get("http://localhost:4000")

		if err != nil {
			return c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
		} else {
			return c.String(http.StatusOK, "OK")
		}
	})

	e.GET("/crud/:method/", func(c echo.Context) error {
		method := c.Param("method")

		var err error

		fmt.Println(method)

		if method == "getRange" {
			var rng Range
			c.Bind(&rng)
			res, err := http.Get(fmt.Sprintf("http://localhost:5000/getRange?left=%d&right=%d", rng.Left, rng.Right))

			if err != nil {
				return c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
			} else {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					return c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
				}

				return c.String(http.StatusOK, string(body))
			}

		} else if method == "getOne" {
			var gOne One
			c.Bind(&gOne)
			res, err := http.Get(fmt.Sprintf("http://localhost:5000/getOne?index=%d", gOne.Index))

			if err != nil {
				return c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
			} else {
				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
				c.Response().WriteHeader(http.StatusOK)
				return json.NewEncoder(c.Response()).Encode(res.Body)
			}

		} else if method == "delete" {
			var gOne One
			c.Bind(&gOne)
		}

		if err != nil {
			return c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
		} else {
			return c.String(http.StatusOK, "OK")
		}
	})

	e.Logger.Fatal(e.Start(":3000"))
}
