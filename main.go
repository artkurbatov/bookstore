package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

var books = []Book{
	{ID: 0, Name: "The Shining", Author: "Stephen King", Quantity: 4},
	{ID: 2, Name: "Dune", Author: "Frank Herbert", Quantity: 7},
	{ID: 3, Name: "Fahrenheit 451", Author: "Ray Bradbury", Quantity: 1},
	{ID: 4, Name: "The Giver", Author: "Lois Lowry", Quantity: 3},
}

func main() {
	s := &http.Server{
		Addr: ":8080",
	}
	handleRoutes()

	fmt.Println("server staring on port %s", s.Addr)
	if err := http.ListenAndServe(s.Addr, nil); err != nil {
		fmt.Printf("server starting error: %s\n", err)
	}
}

func handleRoutes() {
	http.HandleFunc("/books", getBooks)

}

func getBooks(w http.ResponseWriter, r *http.Request) {
	var response JSONResponse

	if r.Method != http.MethodGet {
		response = JSONResponse{Status: Failure, Message: "Wrong method selected"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response = JSONResponse{Status: Success, Data: books}
	json.NewEncoder(w).Encode(response)
}
