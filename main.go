package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	Success = "success"
	Failure = "failure"
)

type Book struct {
	ID       int    `json: id`
	Name     string `json: name`
	Author   string `json: author`
	Quantity int    `json: quantity`
}

type JSONResponse struct {
	Status  string `json: status`
	Data    []Book `json: data`
	Message string `json: message`
}

var db *sql.DB

func main() {
	// Data base configuration
	db = SetupDBConnection()
	createBookTable(db)
	defer db.Close()

	// Router configuration
	SetupRouter()
}

// Router

func SetupRouter() {
	router := gin.Default()
	router.GET("/books", getBooks)
	router.GET("/books/:id", getBook)
	router.POST("/book", addBook)
	router.PATCH("/return", returnBook)
	router.PATCH("/checkout", checkoutBook)
	router.PATCH("/swap", swapBooks)
	router.Run(":8080")
}

// Data base setup

func SetupDBConnection() *sql.DB {
	const path = "user=postgres password=secret dbname=bookstore sslmode=disable"
	const dName = "postgres"

	db, err := sql.Open(dName, path)

	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db

	// createBookTable(db)

	// for i := range books {
	// 	pk := insertBook(db, books[i])

	// 	var name string
	// 	var author string
	// 	var quantity int

	// 	query := "SELECT name, author, quantity FROM bookstore WHERE id = $1"
	// 	if err := db.QueryRow(query, pk).Scan(&name, &author, &quantity); err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	fmt.Printf("Name: %s\n", name)
	// 	fmt.Printf("Author: %s\n", author)
	// 	fmt.Printf("Quantity: %d\n", quantity)
	// }
}

// API handlers

func getBooks(c *gin.Context) {
	query := "SELECT id, name, author, quantity FROM bookstore"
	dbBooks := make([]Book, 0)
	rows, err := db.Query(query)

	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	var id int
	var name string
	var author string
	var quantity int

	for rows.Next() {
		err := rows.Scan(&id, &name, &author, &quantity)

		if err != nil {
			log.Println(err)
			return
		}

		dbBooks = append(dbBooks, Book{id, name, author, quantity})
	}
	log.Println(dbBooks)
	c.IndentedJSON(http.StatusOK, dbBooks)
}

func getBook(c *gin.Context) {
	id := c.Param("id")

	book, err := getBookByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, book)
}

func getBookByID(queryID string) (Book, error) {
	query := `SELECT id, name, author, quantity FROM bookstore WHERE id = $1`
	ind, err := strconv.Atoi(queryID)
	if err != nil {
		return Book{}, fmt.Errorf("book not found")
	}

	var id int
	var name string
	var author string
	var quantity int

	err = db.QueryRow(query, ind).Scan(&id, &name, &author, &quantity)
	if err != nil {
		log.Println(err)
		return Book{}, fmt.Errorf("book not found")
	}
	book := Book{id, name, author, quantity}
	log.Println(book)

	return book, nil
}

// func getBooksByAuthor(author string) (Book, error) {
// 	query := `SELECT id, name, author, quantity FROM bookstore WHERE author = $1`

// 	var book Book

// 	err = db.QueryRow(query, ind).Scan(&book)
// 	handleError(err)

// 	return book, nil
// }

func addBook(c *gin.Context) {
	var book Book

	err := c.BindJSON(&book)
	if err != nil {
		log.Println(err)
		return
	}

	query := `INSERT INTO bookstore(name, author, quantity)
		VALUES ($1, $2, $3) RETURNING id`

	var id int
	err = db.QueryRow(query, book.Name, book.Author, book.Quantity).Scan(&id)

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Added book with id=%d\n", id)
	c.IndentedJSON(http.StatusOK, book)
}

func returnBook(c *gin.Context) {
	id, ok := c.GetQuery("id")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "id not found"})
		return
	}

	book, err := getBookByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}

	book.Quantity++
	c.IndentedJSON(http.StatusOK, book)
}

func checkoutBook(c *gin.Context) {
	id, ok := c.GetQuery("id")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "id not found"})
		return
	}

	book, err := getBookByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}

	if book.Quantity <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "book is out of stock"})
		return
	}
	book.Quantity--
	c.IndentedJSON(http.StatusOK, book)
}

func swapBooks(c *gin.Context) {
	returnID, ok := c.GetQuery("id")
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "return id not found"})
		return
	}

	checkoutID, ok := c.GetQuery("to")
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "checkout id not found"})
		return
	}

	returnB, err := getBookByID(returnID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "return book not found"})
		return
	}

	checkoutB, err := getBookByID(checkoutID)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "checkout book not found"})
		return
	}

	if checkoutB.Quantity <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "checkout book is out of stock"})
		return
	}

	returnB.Quantity++
	checkoutB.Quantity--
	c.IndentedJSON(http.StatusOK, checkoutB)
}

func createBookTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS bookstore(
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		author VARCHAR(100) NOT NULL,
		quantity NUMERIC(4,0) NOT NULL
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// func insertBook(db *sql.DB, book Book) int {
// 	queary := `INSERT INTO bookstore(name, author, quantity)
// 		VALUES ($1, $2, $3) RETURNING id`

// 	var pq int
// 	if err := db.QueryRow(queary, book.Name, book.Author, book.Quantity).Scan(&pq); err != nil {
// 		log.Fatal(err)
// 	}
// 	return pq
// }
