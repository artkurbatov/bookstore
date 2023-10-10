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
