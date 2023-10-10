package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	Success = "success"
	Failure = "failure"
)

type Book struct {
	ID       string `json: id`
	Name     string `json: name`
	Author   string `json: author`
	Quantity int    `json: quantity`
}

type JSONResponse struct {
	Status  string `json: status`
	Data    []Book `json: data`
	Message string `json: message`
}

var books = []Book{
	{ID: "0", Name: "The Shining", Author: "Stephen King", Quantity: 4},
	{ID: "1", Name: "Dune", Author: "Frank Herbert", Quantity: 7},
	{ID: "2", Name: "Fahrenheit 451", Author: "Ray Bradbury", Quantity: 1},
	{ID: "3", Name: "The Giver", Author: "Lois Lowry", Quantity: 3},
}

func main() {
	router := gin.Default()
	router.GET("/books", getBooks)
	router.GET("/books/:id", getBook)
	router.POST("/book", addBook)
	router.PATCH("/return", returnBook)
	router.PATCH("/checkout", checkoutBook)
	router.PATCH("/swap", swapBooks)
	router.Run(":8080")
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
	for i, b := range books {
		if id == b.ID {
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
