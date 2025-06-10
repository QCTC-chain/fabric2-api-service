package main

import (
	"log"
	"net/http"

	"github.com/qctc/fabric2-api-server/router"
)

func main() {
	router := router.SetUpRouter()
	log.Println("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
