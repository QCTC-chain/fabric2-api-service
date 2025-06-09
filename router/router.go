package router

import (
	"github.com/gorilla/mux"

	"github.com/qctc/fabric2-api-server/controller"
)

func SetUpRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/contractlist", controller.GetContractList).Methods("GET")
	return router
}
