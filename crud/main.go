package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
)

func connectToDatabase() *sql.DB {
	godotenv.Load("../.env")
	dbInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", "localhost", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_NAME"))

	db, err := sql.Open("postgres", dbInfo)

	if err != nil {
		log.Fatalf("Error: could not connect to DB; Issue: %v", err)
	}
	err = db.Ping()

	if err != nil {
		log.Fatalf("Error: could not connect to DB. Issue: %v\n", err)
	}
	return db
}

type Response struct {
	Meta Meta   `json:"meta"`
	Data []Data `json:"data"`
}

type Data struct {
	Id      int    `json:"id"`
	User_id int    `json:"user_id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

type Meta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total int   `json:"total"`
	Pages int   `json:"pages"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Links Links `json:"links"`
}

type Links struct {
	Previous string `json:"previous"`
	Current  string `json:"current"`
	Next     string `json:"next"`
}

func getRows(left, right int, db *sql.DB) []Response {
	rows, err := db.Query("SELECT * FROM posts WHERE post_id >= $1 AND post_id <= $2", left, right)

	if err != nil {
		log.Fatalf("Error: could not get rows from DB. Issue: %v\n", err)
	}

	responses := make([]Response, 0, right-left+1)

	for rows.Next() {
		var post_id, total, pages, page, limit int
		var previous, current, next string

		err = rows.Scan(&post_id, &total, &pages, &page, &limit, &previous, &current, &next)

		pagination := Pagination{
			Total: total,
			Pages: pages,
			Page:  page,
			Limit: limit,
			Links: Links{
				Previous: previous,
				Current:  current,
				Next:     next,
			},
		}
		response := Response{
			Meta: Meta{
				Pagination: pagination,
			},
		}
		responses = append(responses, response)

		if err != nil {
			log.Fatalf("Error: could not scan rows. Issue: %v\n", err)
		}
	}

	rows, err = db.Query("SELECT * FROM datas WHERE post_id BETWEEN $1 AND $2;", left, right)

	if err != nil {
		fmt.Printf("Error: failed to get rows from DB. Issue: %v\n", err)
	}

	for rows.Next() {
		var post_id, id, user_id int
		var title, body string

		err = rows.Scan(&post_id, &id, &user_id, &title, &body)

		data := Data{
			Id:      id,
			User_id: user_id,
			Title:   title,
			Body:    body,
		}
		responses[post_id-left].Data = append(responses[post_id-left].Data, data)

		if err != nil {
			fmt.Printf("Error: failed to scan rows. Issue: %v\n", err)
		}
	}
	return responses
}

func deleteRow(index int, db *sql.DB) error {
	_, err := db.Exec("DELETE FROM posts WHERE post_id = $1", index)

	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM datas WHERE post_id = $1", index)

	if err != nil {
		return err
	}
	return nil
}

func main() {
	db := connectToDatabase()
	e := echo.New()

	e.GET("/getRange", func(c echo.Context) error {
		left, err := strconv.Atoi(c.QueryParam("left"))

		if err != nil {
			fmt.Printf("Error: failed to convert left to int. Issue: %v\n", err)
		}
		right, err := strconv.Atoi(c.QueryParam("right"))

		if err != nil {
			fmt.Printf("Error: failed to convert right to int. Issue: %v\n", err)
		}

		responses := getRows(left, right, db)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return json.NewEncoder(c.Response()).Encode(responses)
	})

	e.GET("/getOne", func(c echo.Context) error {
		index, err := strconv.Atoi(c.QueryParam("index"))

		if err != nil {
			fmt.Printf("Error: failed to convert index to int. Issue: %v\n", err)
		}

		responses := getRows(index, index, db)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return json.NewEncoder(c.Response()).Encode(responses[0])
	})

	// e.PATCH("/", func(c echo.Context) error {

	// })

	e.DELETE("/delete", func(c echo.Context) error {
		index, err := strconv.Atoi(c.QueryParam("index"))

		if err != nil {
			fmt.Printf("Error: failed to convert index to int. Issue: %v\n", err)
		}

		err = deleteRow(index, db)

		if err != nil {
			return c.String(http.StatusBadGateway, fmt.Sprintf("Error: failed to delete row. Issue: %v\n", err))
		} else {
			return c.String(http.StatusOK, "Row deleted successfully")
		}
	})

	e.Logger.Fatal(e.Start(":5000"))
}
