package main

import (
	"log"
	"net/http"

	"github.com/qctc/fabric2-api-server/router"
	"github.com/qctc/fabric2-api-server/service"
)

func main() {
	// Initialize Fabric2Service
	if err := service.InitFabric2Service("./config/config.yaml"); err != nil {
		log.Fatalf("Failed to initialize Fabric2Service: %v", err)
	}

	router := router.SetUpRouter()
	log.Println("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
