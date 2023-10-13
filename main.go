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

var DB sql.DB

var books = []Book{
	{ID: 0, Name: "The Shining", Author: "Stephen King", Quantity: 4},
	{ID: 1, Name: "Dune", Author: "Frank Herbert", Quantity: 7},
	{ID: 2, Name: "Fahrenheit 451", Author: "Ray Bradbury", Quantity: 1},
	{ID: 3, Name: "The Giver", Author: "Lois Lowry", Quantity: 3},
}

func main() {
	SetupDBConnection()
	SetupRouter()
}

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

func SetupDBConnection() {
	const path = "user=postgres password=secret dbname=bookstore sslmode=disable"
	const dName = "postgres"

	db, err := sql.Open(dName, path)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	createBookTable(db)

	for i := range books {
		pk := insertBook(db, books[i])

		var name string
		var author string
		var quantity int

		quary := "SELECT name, author, quantity FROM bookstore WHERE id = $1"
		if err := db.QueryRow(quary, pk).Scan(&name, &author, &quantity); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Author: %s\n", author)
		fmt.Printf("Quantity: %d\n", quantity)
	}

	quary := "SELECT id, name, author, quantity FROM bookstore"
	dbBooks := []Book{}
	rows, err := db.Query(quary)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var id int
	var name string
	var author string
	var quantity int

	for rows.Next() {
		if err := rows.Scan(&id, &name, &author, &quantity); err != nil {
			log.Fatal(err)
		}
		dbBooks = append(dbBooks, Book{id, name, author, quantity})
	}
	fmt.Println(dbBooks)
}

func getBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, books)
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

func getBookByID(id string) (*Book, error) {
	ind, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("book not found")
	}

	for i, b := range books {
		if ind == b.ID {
			return &books[i], nil
		}
	}
	return nil, fmt.Errorf("book not found")
}

func addBook(c *gin.Context) {
	var book Book

	if err := c.BindJSON(&book); err != nil {
		return
	}
	books = append(books, book)
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

func insertBook(db *sql.DB, book Book) int {
	queary := `INSERT INTO bookstore(name, author, quantity)
		VALUES ($1, $2, $3) RETURNING id`

	var pq int
	if err := db.QueryRow(queary, book.Name, book.Author, book.Quantity).Scan(&pq); err != nil {
		log.Fatal(err)
	}
	return pq
}
