package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	Meta struct {
		Pagination struct {
			Total int `stbl:"total"`
			Pages int `stbl:"pages"`
			Page  int `stbl:"page"`
			Limit int `stbl:"limit"`
			Links struct {
				Previous string `stbl:"previous"`
				Current  string `stbl:"current"`
				Next     string `stbl:"next"`
			}
		}
	}
	Data []struct {
		Id      int
		User_id int
		Title   string
		Body    string
	}
}

func resetTable(db *sql.DB) bool {
	_, err := db.Exec("DROP TABLE IF EXISTS posts")
	if err != nil {
		fmt.Println("Error: failed to delete posts table in DB")
		return false
	}
	_, err = db.Exec("DROP TABLE IF EXISTS datas")

	if err != nil {
		fmt.Println("Error: failed to delete datas table in DB")
		return false
	}

	res, err := db.Exec("CREATE TABLE IF NOT EXISTS posts (post_id INT unique, total INT, pages INT, page INT, lim INT, previous VARCHAR(256), current VARCHAR(256), next VARCHAR(256));")

	fmt.Println(res)

	if err != nil {
		fmt.Println("Error: failed to insert posts table in DB. %v\n", err)
		return false
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS datas (post_id int, id int, user_id int, title VARCHAR(256), body VARCHAR(512));")

	if err != nil {
		fmt.Println("Error: failed to create datas table to DB")
		return false
	} else {
		return true
	}
}

func insertRow(postId int, response Response, db *sql.DB) bool {
	p := response.Meta.Pagination
	res, err := db.Exec("INSERT INTO posts (post_id, total, pages, page, lim, previous, current, next) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);", postId, p.Total, p.Pages, p.Page, p.Limit, p.Links.Previous, p.Links.Current, p.Links.Next)

	fmt.Println(res)

	if err != nil {
		fmt.Println("Error: failed to insert posts table in DB")
		return false
	}
	for i := 0; i < len(response.Data); i++ {
		d := response.Data[i]

		_, err = db.Exec("INSERT INTO datas (post_id, id, user_id, title, body) VALUES ($1, $2, $3, $4, $5)", postId, d.Id, d.User_id, d.Title, d.Body)
		if err != nil {
			fmt.Printf("Error: failed to insert row into table datas in DB. %v\n", err)
			return false
		}
	}
	return true
}

func extractData(db *sql.DB) error {
	url := "https://gorest.co.in/public/v1/posts"

	resetTable(db)

	for i := 1; i <= 50; i++ {
		query := url + fmt.Sprintf("?page=%d", i)
		resp, err := http.Get(query)
		if err != nil {
			log.Fatalf("Error: could not get data from query. Issue: %v\n", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error: parsing body from the response query. Issue: %v\n", err)
		}
		var response Response
		json.Unmarshal(body, &response)
		insertRow(i, response, db)

		defer resp.Body.Close()
	}
	return nil
}

func main() {
	db := connectToDatabase()

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		err := extractData(db)

		if err != nil {
			return c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
		} else {
			return c.String(http.StatusOK, "OK")
		}
	})

	e.Logger.Fatal(e.Start(":4000"))
}
